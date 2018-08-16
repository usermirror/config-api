package vault

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	metrics "github.com/armon/go-metrics"
	"github.com/hashicorp/errwrap"
	multierror "github.com/hashicorp/go-multierror"
	uuid "github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/audit"
	"github.com/hashicorp/vault/helper/consts"
	"github.com/hashicorp/vault/helper/jsonutil"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/physical"
	"github.com/oklog/run"
)

// Standby checks if the Vault is in standby mode
func (c *Core) Standby() (bool, error) {
	c.stateLock.RLock()
	standby := c.standby
	c.stateLock.RUnlock()
	return standby, nil
}

// Leader is used to get the current active leader
func (c *Core) Leader() (isLeader bool, leaderAddr, clusterAddr string, err error) {
	// Check if HA enabled. We don't need the lock for this check as it's set
	// on startup and never modified
	if c.ha == nil {
		return false, "", "", ErrHANotEnabled
	}

	// Check if sealed
	if c.Sealed() {
		return false, "", "", consts.ErrSealed
	}

	c.stateLock.RLock()

	// Check if we are the leader
	if !c.standby {
		c.stateLock.RUnlock()
		return true, c.redirectAddr, c.clusterAddr, nil
	}

	// Initialize a lock
	lock, err := c.ha.LockWith(coreLockPath, "read")
	if err != nil {
		c.stateLock.RUnlock()
		return false, "", "", err
	}

	// Read the value
	held, leaderUUID, err := lock.Value()
	if err != nil {
		c.stateLock.RUnlock()
		return false, "", "", err
	}
	if !held {
		c.stateLock.RUnlock()
		return false, "", "", nil
	}

	c.clusterLeaderParamsLock.RLock()
	localLeaderUUID := c.clusterLeaderUUID
	localRedirAddr := c.clusterLeaderRedirectAddr
	localClusterAddr := c.clusterLeaderClusterAddr
	c.clusterLeaderParamsLock.RUnlock()

	// If the leader hasn't changed, return the cached value; nothing changes
	// mid-leadership, and the barrier caches anyways
	if leaderUUID == localLeaderUUID && localRedirAddr != "" {
		c.stateLock.RUnlock()
		return false, localRedirAddr, localClusterAddr, nil
	}

	c.logger.Trace("found new active node information, refreshing")

	defer c.stateLock.RUnlock()
	c.clusterLeaderParamsLock.Lock()
	defer c.clusterLeaderParamsLock.Unlock()

	// Validate base conditions again
	if leaderUUID == c.clusterLeaderUUID && c.clusterLeaderRedirectAddr != "" {
		return false, localRedirAddr, localClusterAddr, nil
	}

	key := coreLeaderPrefix + leaderUUID
	// Use background because postUnseal isn't run on standby
	entry, err := c.barrier.Get(context.Background(), key)
	if err != nil {
		return false, "", "", err
	}
	if entry == nil {
		return false, "", "", nil
	}

	var oldAdv bool

	var adv activeAdvertisement
	err = jsonutil.DecodeJSON(entry.Value, &adv)
	if err != nil {
		// Fall back to pre-struct handling
		adv.RedirectAddr = string(entry.Value)
		c.logger.Debug("parsed redirect addr for new active node", "redirect_addr", adv.RedirectAddr)
		oldAdv = true
	}

	if !oldAdv {
		c.logger.Debug("parsing information for new active node", "active_cluster_addr", adv.ClusterAddr, "active_redirect_addr", adv.RedirectAddr)

		// Ensure we are using current values
		err = c.loadLocalClusterTLS(adv)
		if err != nil {
			return false, "", "", err
		}

		// This will ensure that we both have a connection at the ready and that
		// the address is the current known value
		// Since this is standby, we don't use the active context. Later we may
		// use a process-scoped context
		err = c.refreshRequestForwardingConnection(context.Background(), adv.ClusterAddr)
		if err != nil {
			return false, "", "", err
		}
	}

	// Don't set these until everything has been parsed successfully or we'll
	// never try again
	c.clusterLeaderRedirectAddr = adv.RedirectAddr
	c.clusterLeaderClusterAddr = adv.ClusterAddr
	c.clusterLeaderUUID = leaderUUID

	return false, adv.RedirectAddr, adv.ClusterAddr, nil
}

// StepDown is used to step down from leadership
func (c *Core) StepDown(httpCtx context.Context, req *logical.Request) (retErr error) {
	defer metrics.MeasureSince([]string{"core", "step_down"}, time.Now())

	if req == nil {
		retErr = multierror.Append(retErr, errors.New("nil request to step-down"))
		return retErr
	}

	c.stateLock.RLock()
	defer c.stateLock.RUnlock()
	if c.Sealed() {
		return nil
	}
	if c.ha == nil || c.standby {
		return nil
	}

	ctx, cancel := context.WithCancel(c.activeContext)
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
		case <-httpCtx.Done():
			cancel()
		}
	}()

	acl, te, entity, identityPolicies, err := c.fetchACLTokenEntryAndEntity(req)
	if err != nil {
		if errwrap.ContainsType(err, new(TemplateError)) {
			c.logger.Warn("permission denied due to a templated policy being invalid or containing directives not satisfied by the requestor", "error", err)
			err = logical.ErrPermissionDenied
		}
		retErr = multierror.Append(retErr, err)
		return retErr
	}

	// Audit-log the request before going any further
	auth := &logical.Auth{
		ClientToken:      req.ClientToken,
		Policies:         identityPolicies,
		IdentityPolicies: identityPolicies,
	}
	if te != nil {
		auth.TokenPolicies = te.Policies
		auth.Policies = append(te.Policies, identityPolicies...)
		auth.Metadata = te.Meta
		auth.DisplayName = te.DisplayName
		auth.EntityID = te.EntityID
	}

	logInput := &audit.LogInput{
		Auth:    auth,
		Request: req,
	}
	if err := c.auditBroker.LogRequest(ctx, logInput, c.auditedHeaders); err != nil {
		c.logger.Error("failed to audit request", "request_path", req.Path, "error", err)
		retErr = multierror.Append(retErr, errors.New("failed to audit request, cannot continue"))
		return retErr
	}

	if entity != nil && entity.Disabled {
		c.logger.Warn("permission denied as the entity on the token is disabled")
		retErr = multierror.Append(retErr, logical.ErrPermissionDenied)
		c.stateLock.RUnlock()
		return retErr
	}

	if te != nil && te.EntityID != "" && entity == nil {
		c.logger.Warn("permission denied as the entity on the token is invalid")
		retErr = multierror.Append(retErr, logical.ErrPermissionDenied)
		c.stateLock.RUnlock()
		return retErr
	}

	// Attempt to use the token (decrement num_uses)
	if te != nil {
		te, err = c.tokenStore.UseToken(ctx, te)
		if err != nil {
			c.logger.Error("failed to use token", "error", err)
			retErr = multierror.Append(retErr, ErrInternalError)
			return retErr
		}
		if te == nil {
			// Token has been revoked
			retErr = multierror.Append(retErr, logical.ErrPermissionDenied)
			return retErr
		}
	}

	// Verify that this operation is allowed
	authResults := c.performPolicyChecks(ctx, acl, te, req, entity, &PolicyCheckOpts{
		RootPrivsRequired: true,
	})
	if !authResults.Allowed {
		c.stateLock.RUnlock()
		retErr = multierror.Append(retErr, authResults.Error)
		if authResults.Error.ErrorOrNil() == nil || authResults.DeniedError {
			retErr = multierror.Append(retErr, logical.ErrPermissionDenied)
		}
		return retErr
	}

	if te != nil && te.NumUses == tokenRevocationPending {
		// Token needs to be revoked. We do this immediately here because
		// we won't have a token store after sealing.
		leaseID, err := c.expiration.CreateOrFetchRevocationLeaseByToken(c.activeContext, te)
		if err == nil {
			err = c.expiration.Revoke(c.activeContext, leaseID)
		}
		if err != nil {
			c.logger.Error("token needed revocation before step-down but failed to revoke", "error", err)
			retErr = multierror.Append(retErr, ErrInternalError)
		}
	}

	select {
	case c.manualStepDownCh <- struct{}{}:
	default:
		c.logger.Warn("manual step-down operation already queued")
	}

	return retErr
}

// runStandby is a long running process that manages a number of the HA
// subsystems.
func (c *Core) runStandby(doneCh, manualStepDownCh, stopCh chan struct{}) {
	defer close(doneCh)
	defer close(manualStepDownCh)
	c.logger.Info("entering standby mode")

	var g run.Group
	{
		// This will cause all the other actors to close when the stop channel
		// is closed.
		g.Add(func() error {
			<-stopCh
			return nil
		}, func(error) {})
	}
	{
		// Monitor for key rotation
		keyRotateDone := make(chan struct{})
		keyRotateStop := make(chan struct{})

		g.Add(func() error {
			c.periodicCheckKeyUpgrade(context.Background(), keyRotateDone, keyRotateStop)
			return nil
		}, func(error) {
			close(keyRotateStop)
			c.logger.Debug("shutting down periodic key rotation checker")
			<-keyRotateDone
		})
	}
	{
		// Monitor for new leadership
		checkLeaderDone := make(chan struct{})
		checkLeaderStop := make(chan struct{})

		g.Add(func() error {
			c.periodicLeaderRefresh(checkLeaderDone, checkLeaderStop)
			return nil
		}, func(error) {
			close(checkLeaderStop)
			c.logger.Debug("shutting down periodic leader refresh")
			<-checkLeaderDone
		})
	}
	{
		// Wait for leadership
		leaderDoneCh := make(chan struct{})
		leaderStopCh := make(chan struct{})

		g.Add(func() error {
			c.waitForLeadership(leaderDoneCh, manualStepDownCh, leaderStopCh)
			return nil
		}, func(error) {
			close(leaderStopCh)
			c.logger.Debug("shutting down leader elections")
			<-leaderDoneCh
		})
	}

	// Start all the actors
	g.Run()
}

// waitForLeadership is a long running routine that is used when an HA backend
// is enabled. It waits until we are leader and switches this Vault to
// active.
func (c *Core) waitForLeadership(doneCh, manualStepDownCh, stopCh chan struct{}) {
	defer close(doneCh)

	c.logger.Info("entering standby mode")

	var manualStepDown bool
	for {
		// Check for a shutdown
		select {
		case <-stopCh:
			c.logger.Debug("stop channel triggered in runStandby")
			return
		default:
			// If we've just down, we could instantly grab the lock again. Give
			// the other nodes a chance.
			if manualStepDown {
				time.Sleep(manualStepDownSleepPeriod)
				manualStepDown = false
			}
		}

		// Create a lock
		uuid, err := uuid.GenerateUUID()
		if err != nil {
			c.logger.Error("failed to generate uuid", "error", err)
			return
		}
		lock, err := c.ha.LockWith(coreLockPath, uuid)
		if err != nil {
			c.logger.Error("failed to create lock", "error", err)
			return
		}

		// Attempt the acquisition
		leaderLostCh := c.acquireLock(lock, stopCh)

		// Bail if we are being shutdown
		if leaderLostCh == nil {
			return
		}
		c.logger.Info("acquired lock, enabling active operation")

		// This is used later to log a metrics event; this can be helpful to
		// detect flapping
		activeTime := time.Now()

		// Grab the lock as we need it for cluster setup, which needs to happen
		// before advertising;

		lockGrabbedCh := make(chan struct{})
		go func() {
			// Grab the lock
			c.stateLock.Lock()
			// If stopCh has been closed, which only happens while the
			// stateLock is held, we have actually terminated, so we just
			// instantly give up the lock, otherwise we notify that it's ready
			// for consumption
			select {
			case <-stopCh:
				c.stateLock.Unlock()
			default:
				close(lockGrabbedCh)
			}
		}()

		select {
		case <-stopCh:
			lock.Unlock()
			metrics.MeasureSince([]string{"core", "leadership_setup_failed"}, activeTime)
			return
		case <-lockGrabbedCh:
			// We now have the lock and can use it
		}

		if c.Sealed() {
			c.logger.Warn("grabbed HA lock but already sealed, exiting")
			lock.Unlock()
			c.stateLock.Unlock()
			metrics.MeasureSince([]string{"core", "leadership_setup_failed"}, activeTime)
			return
		}

		// Store the lock so that we can manually clear it later if needed
		c.heldHALock = lock

		// Create the active context
		activeCtx, activeCtxCancel := context.WithCancel(context.Background())
		c.activeContext = activeCtx
		c.activeContextCancelFunc.Store(activeCtxCancel)

		// This block is used to wipe barrier/seal state and verify that
		// everything is sane. If we have no sanity in the barrier, we actually
		// seal, as there's little we can do.
		{
			c.seal.SetBarrierConfig(activeCtx, nil)
			if c.seal.RecoveryKeySupported() {
				c.seal.SetRecoveryConfig(activeCtx, nil)
			}

			if err := c.performKeyUpgrades(activeCtx); err != nil {
				// We call this in a goroutine so that we can give up the
				// statelock and have this shut us down; sealInternal has a
				// workflow where it watches for the stopCh to close so we want
				// to return from here
				c.logger.Error("error performing key upgrades", "error", err)
				go c.Shutdown()
				c.heldHALock = nil
				lock.Unlock()
				c.stateLock.Unlock()
				metrics.MeasureSince([]string{"core", "leadership_setup_failed"}, activeTime)
				return
			}
		}

		{
			// Clear previous local cluster cert info so we generate new. Since the
			// UUID will have changed, standbys will know to look for new info
			c.localClusterParsedCert.Store((*x509.Certificate)(nil))
			c.localClusterCert.Store(([]byte)(nil))
			c.localClusterPrivateKey.Store((*ecdsa.PrivateKey)(nil))

			if err := c.setupCluster(activeCtx); err != nil {
				c.heldHALock = nil
				lock.Unlock()
				c.stateLock.Unlock()
				c.logger.Error("cluster setup failed", "error", err)
				metrics.MeasureSince([]string{"core", "leadership_setup_failed"}, activeTime)
				continue
			}
		}
		// Advertise as leader
		if err := c.advertiseLeader(activeCtx, uuid, leaderLostCh); err != nil {
			c.heldHALock = nil
			lock.Unlock()
			c.stateLock.Unlock()
			c.logger.Error("leader advertisement setup failed", "error", err)
			metrics.MeasureSince([]string{"core", "leadership_setup_failed"}, activeTime)
			continue
		}

		// Attempt the post-unseal process
		err = c.postUnseal(activeCtx, activeCtxCancel)
		if err == nil {
			c.standby = false
		}

		c.stateLock.Unlock()

		// Handle a failure to unseal
		if err != nil {
			c.logger.Error("post-unseal setup failed", "error", err)
			lock.Unlock()
			metrics.MeasureSince([]string{"core", "leadership_setup_failed"}, activeTime)
			continue
		}

		cancelCtxAndLock := func() {
			go func() {
				select {
				case <-activeCtx.Done():
				// Attempt to drain any inflight requests
				case <-time.After(DefaultMaxRequestDuration):
					activeCtxCancel()
				}
			}()
			c.stateLock.Lock()
			activeCtxCancel()
		}

		runSealing := func() {
			metrics.MeasureSince([]string{"core", "leadership_lost"}, activeTime)

			c.standby = true

			if err := c.preSeal(); err != nil {
				c.logger.Error("pre-seal teardown failed", "error", err)
			}
		}

		releaseHALock := func() {
			// We may hit this from leaderLostCh or manualStepDownCh if they
			// triggered before stopCh, so we check here instead of only in the
			// stopCh case so we can try to do the right thing then, too
			if atomic.LoadUint32(c.keepHALockOnStepDown) == 1 {
				return
			}
			if err := c.clearLeader(uuid); err != nil {
				c.logger.Error("clearing leader advertisement failed", "error", err)
			}
			c.heldHALock.Unlock()
			c.heldHALock = nil
		}

		// Monitor a loss of leadership
		select {
		case <-leaderLostCh:
			c.logger.Warn("leadership lost, stopping active operation")
			cancelCtxAndLock()
			runSealing()
			releaseHALock()
			c.stateLock.Unlock()

		case <-stopCh:
			activeCtxCancel()
			runSealing()
			releaseHALock()
			return

		case <-manualStepDownCh:
			manualStepDown = true
			c.logger.Warn("stepping down from active operation to standby")

			cancelCtxAndLock()
			runSealing()
			releaseHALock()
			c.stateLock.Unlock()
		}
	}
}

// This checks the leader periodically to ensure that we switch RPC to a new
// leader pretty quickly. There is logic in Leader() already to not make this
// onerous and avoid more traffic than needed, so we just call that and ignore
// the result.
func (c *Core) periodicLeaderRefresh(doneCh, stopCh chan struct{}) {
	defer close(doneCh)
	opCount := new(int32)
	for {
		select {
		case <-time.After(leaderCheckInterval):
			count := atomic.AddInt32(opCount, 1)
			if count > 1 {
				atomic.AddInt32(opCount, -1)
				continue
			}
			// We do this in a goroutine because otherwise if this refresh is
			// called while we're shutting down the call to Leader() can
			// deadlock, which then means stopCh can never been seen and we can
			// block shutdown
			go func() {
				// Bind locally, as the race detector is tripping here
				lopCount := opCount
				c.Leader()
				atomic.AddInt32(lopCount, -1)
			}()
		case <-stopCh:
			return
		}
	}
}

// periodicCheckKeyUpgrade is used to watch for key rotation events as a standby
func (c *Core) periodicCheckKeyUpgrade(ctx context.Context, doneCh, stopCh chan struct{}) {
	defer close(doneCh)
	opCount := new(int32)
	for {
		select {
		case <-time.After(keyRotateCheckInterval):
			count := atomic.AddInt32(opCount, 1)
			if count > 1 {
				atomic.AddInt32(opCount, -1)
				continue
			}

			go func() {
				// Bind locally, as the race detector is tripping here
				lopCount := opCount

				// Only check if we are a standby
				c.stateLock.RLock()
				standby := c.standby
				c.stateLock.RUnlock()
				if !standby {
					atomic.AddInt32(lopCount, -1)
					return
				}

				// Check for a poison pill. If we can read it, it means we have stale
				// keys (e.g. from replication being activated) and we need to seal to
				// be unsealed again.
				entry, _ := c.barrier.Get(ctx, poisonPillPath)
				if entry != nil && len(entry.Value) > 0 {
					c.logger.Warn("encryption keys have changed out from underneath us (possibly due to replication enabling), must be unsealed again")
					go c.Shutdown()
					atomic.AddInt32(lopCount, -1)
					return
				}

				if err := c.checkKeyUpgrades(ctx); err != nil {
					c.logger.Error("key rotation periodic upgrade check failed", "error", err)
				}

				atomic.AddInt32(lopCount, -1)
				return
			}()
		case <-stopCh:
			return
		}
	}
}

// checkKeyUpgrades is used to check if there have been any key rotations
// and if there is a chain of upgrades available
func (c *Core) checkKeyUpgrades(ctx context.Context) error {
	for {
		// Check for an upgrade
		didUpgrade, newTerm, err := c.barrier.CheckUpgrade(ctx)
		if err != nil {
			return err
		}

		// Nothing to do if no upgrade
		if !didUpgrade {
			break
		}
		if c.logger.IsInfo() {
			c.logger.Info("upgraded to new key term", "term", newTerm)
		}
	}
	return nil
}

func (c *Core) performKeyUpgrades(ctx context.Context) error {
	if err := c.checkKeyUpgrades(ctx); err != nil {
		return errwrap.Wrapf("error checking for key upgrades: {{err}}", err)
	}

	if err := c.barrier.ReloadMasterKey(ctx); err != nil {
		return errwrap.Wrapf("error reloading master key: {{err}}", err)
	}

	if err := c.barrier.ReloadKeyring(ctx); err != nil {
		return errwrap.Wrapf("error reloading keyring: {{err}}", err)
	}

	if err := c.scheduleUpgradeCleanup(ctx); err != nil {
		return errwrap.Wrapf("error scheduling upgrade cleanup: {{err}}", err)
	}

	return nil
}

// scheduleUpgradeCleanup is used to ensure that all the upgrade paths
// are cleaned up in a timely manner if a leader failover takes place
func (c *Core) scheduleUpgradeCleanup(ctx context.Context) error {
	// List the upgrades
	upgrades, err := c.barrier.List(ctx, keyringUpgradePrefix)
	if err != nil {
		return errwrap.Wrapf("failed to list upgrades: {{err}}", err)
	}

	// Nothing to do if no upgrades
	if len(upgrades) == 0 {
		return nil
	}

	// Schedule cleanup for all of them
	time.AfterFunc(keyRotateGracePeriod, func() {
		sealed, err := c.barrier.Sealed()
		if err != nil {
			c.logger.Warn("failed to check barrier status at upgrade cleanup time")
			return
		}
		if sealed {
			c.logger.Warn("barrier sealed at upgrade cleanup time")
			return
		}
		for _, upgrade := range upgrades {
			path := fmt.Sprintf("%s%s", keyringUpgradePrefix, upgrade)
			if err := c.barrier.Delete(ctx, path); err != nil {
				c.logger.Error("failed to cleanup upgrade", "path", path, "error", err)
			}
		}
	})
	return nil
}

// acquireLock blocks until the lock is acquired, returning the leaderLostCh
func (c *Core) acquireLock(lock physical.Lock, stopCh <-chan struct{}) <-chan struct{} {
	for {
		// Attempt lock acquisition
		leaderLostCh, err := lock.Lock(stopCh)
		if err == nil {
			return leaderLostCh
		}

		// Retry the acquisition
		c.logger.Error("failed to acquire lock", "error", err)
		select {
		case <-time.After(lockRetryInterval):
		case <-stopCh:
			return nil
		}
	}
}

// advertiseLeader is used to advertise the current node as leader
func (c *Core) advertiseLeader(ctx context.Context, uuid string, leaderLostCh <-chan struct{}) error {
	go c.cleanLeaderPrefix(ctx, uuid, leaderLostCh)

	var key *ecdsa.PrivateKey
	switch c.localClusterPrivateKey.Load().(type) {
	case *ecdsa.PrivateKey:
		key = c.localClusterPrivateKey.Load().(*ecdsa.PrivateKey)
	default:
		c.logger.Error("unknown cluster private key type", "key_type", fmt.Sprintf("%T", c.localClusterPrivateKey.Load()))
		return fmt.Errorf("unknown cluster private key type %T", c.localClusterPrivateKey.Load())
	}

	keyParams := &clusterKeyParams{
		Type: corePrivateKeyTypeP521,
		X:    key.X,
		Y:    key.Y,
		D:    key.D,
	}

	locCert := c.localClusterCert.Load().([]byte)
	localCert := make([]byte, len(locCert))
	copy(localCert, locCert)
	adv := &activeAdvertisement{
		RedirectAddr:     c.redirectAddr,
		ClusterAddr:      c.clusterAddr,
		ClusterCert:      localCert,
		ClusterKeyParams: keyParams,
	}
	val, err := jsonutil.EncodeJSON(adv)
	if err != nil {
		return err
	}
	ent := &Entry{
		Key:   coreLeaderPrefix + uuid,
		Value: val,
	}
	err = c.barrier.Put(ctx, ent)
	if err != nil {
		return err
	}

	sd, ok := c.ha.(physical.ServiceDiscovery)
	if ok {
		if err := sd.NotifyActiveStateChange(); err != nil {
			if c.logger.IsWarn() {
				c.logger.Warn("failed to notify active status", "error", err)
			}
		}
	}
	return nil
}

func (c *Core) cleanLeaderPrefix(ctx context.Context, uuid string, leaderLostCh <-chan struct{}) {
	keys, err := c.barrier.List(ctx, coreLeaderPrefix)
	if err != nil {
		c.logger.Error("failed to list entries in core/leader", "error", err)
		return
	}
	for len(keys) > 0 {
		select {
		case <-time.After(leaderPrefixCleanDelay):
			if keys[0] != uuid {
				c.barrier.Delete(ctx, coreLeaderPrefix+keys[0])
			}
			keys = keys[1:]
		case <-leaderLostCh:
			return
		}
	}
}

// clearLeader is used to clear our leadership entry
func (c *Core) clearLeader(uuid string) error {
	key := coreLeaderPrefix + uuid
	err := c.barrier.Delete(context.Background(), key)

	// Advertise ourselves as a standby
	sd, ok := c.ha.(physical.ServiceDiscovery)
	if ok {
		if err := sd.NotifyActiveStateChange(); err != nil {
			if c.logger.IsWarn() {
				c.logger.Warn("failed to notify standby status", "error", err)
			}
		}
	}

	return err
}

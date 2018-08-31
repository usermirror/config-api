package storage

import (
	"context"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
)

// NewEtcd creates a new Store through the etcdv3 client.
func NewEtcd(etcdEndpoint string) (*Etcd, error) {
	if etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdEndpoint},
		DialTimeout: 5 * time.Second,
	}); err != nil {
		return nil, err
	} else {
		return &Etcd{
			client: etcd,
		}, nil
	}

}

// Etcd backed persistence for arbitrary key/values.
type Etcd struct {
	client *clientv3.Client
}

// implements Store interface
var _ Store = new(Etcd)

func (e *Etcd) Init() error {
	return nil
}

func (e *Etcd) Get(input GetInput) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resp, err := e.client.Get(ctx, input.Key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	return resp.Kvs[0].Value, nil
}

func (e *Etcd) Set(input SetInput) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := e.client.Put(ctx, input.Key, string(input.Value))
	return err
}

func (e *Etcd) Scan(input ScanInput) (KeyList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	kl := KeyList{}
	resp, err := e.client.Get(ctx, input.Prefix, clientv3.WithPrefix())
	if err != nil {
		return kl, err
	}

	if len(resp.Kvs) == 0 {
		return kl, nil
	}

	for _, item := range resp.Kvs {
		kl.Keys = append(kl.Keys, strings.Replace(string(item.Key), input.Prefix+"::", "", 1))
	}

	return kl, err
}

func (e *Etcd) Close() error {
	return e.client.Close()
}

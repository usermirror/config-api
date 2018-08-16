package storage

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
)

// NewVault creates a new Store through the vault api client.
func NewVault(vaultAddress string, vaultToken string) (*Vault, error) {
	if vault, err := api.NewClient(&api.Config{
		Address: vaultAddress,
		Timeout: 5 * time.Second,
	}); err != nil {
		return nil, err
	} else {
		vault.SetToken(vaultToken)

		return &Vault{
			client: vault,
		}, nil
	}
}

// Vault backed persistence for arbitrary key/values.
type Vault struct {
	client *api.Client
}

// implements Store interface
var _ Store = new(Vault)

const valueKey = "value"

func (v *Vault) Get(input GetInput) ([]byte, error) {
	keyName := "secret/data/" + strings.Replace(input.Key, "::", "-", 1)

	fmt.Println(fmt.Sprintf("storage.vault.get: %s", keyName))
	resp, err := v.client.Logical().Read(keyName)
	if err != nil {
		return nil, err
	}

	if resp == nil || resp.Data["data"] == nil {
		return nil, nil
	}

	data, ok := resp.Data["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("vault.get: failed to convert data to map[string]interface{}")
	}

	if data["value"] == nil {
		return nil, nil
	}

	value, err := toBytes(data[valueKey])
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (v *Vault) Set(input SetInput) error {
	keyName := "secret/data/" + strings.Replace(input.Key, "::", "-", 1)

	fmt.Println(fmt.Sprintf("storage.vault.set: %s %v", keyName, map[string]interface{}{
		valueKey: string(input.Value),
	}))
	_, err := v.client.Logical().Write(keyName, map[string]interface{}{
		"data": map[string]interface{}{
			valueKey: string(input.Value),
		},
	})

	return err
}

func (v *Vault) Close() error {
	v.client.ClearToken()
	return nil

	// TODO: dereference client?
	// return v.client.Close()
}

func toBytes(value interface{}) ([]byte, error) {
	str, ok := value.(string)
	if !ok {
		return nil, errors.New("vault.toBytes: failed to convert to string")
	}

	return []byte(str), nil
}

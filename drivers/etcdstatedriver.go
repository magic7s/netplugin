package drivers

import (
	"encoding/base64"
	"github.com/coreos/go-etcd/etcd"

	"github.com/mapuri/netplugin/core"
)

// implements the StateDriver interface for an etcd based distributed
// key-value store used to store config and runtime state for the netplugin.

type EtcdStateDriverConfig struct {
	Etcd struct {
		Machines []string
	}
}

type EtcdStateDriver struct {
	Client *etcd.Client
}

func (d *EtcdStateDriver) Init(config *core.Config) error {
	// XXX: initialize with default setting i.e. etcd cluster running on
	// localhost. Revisit once we figure out how the config is passed to
	// the state driver
	cfg, ok := config.V.(EtcdStateDriverConfig)

	if !ok {
		return &core.Error{Desc: "Invalid config type passed!"}
	}

	d.Client = etcd.NewClient(cfg.Etcd.Machines)

	return nil
}

func (d *EtcdStateDriver) Deinit() {
}

func (d *EtcdStateDriver) Write(key string, value []byte) error {
	// XXX: etcd go client right now accepts only string values, so
	// encode the received byte array as base64 string before storing it.
	encodedStr := base64.URLEncoding.EncodeToString(value)
	_, err := d.Client.Set(key, encodedStr, 5)

	return err
}

func (d *EtcdStateDriver) Read(key string) ([]byte, error) {
	resp, err := d.Client.Get(key, false, false)
	if err != nil {
		return []byte{}, err
	}

	// XXX: etcd go client right now accepts only string values, so
	// decode the received data as base64 string.
	decodedStr := []byte{}
	decodedStr, err = base64.URLEncoding.DecodeString(resp.Node.Value)
	if err != nil {
		return []byte{}, err
	}

	return decodedStr, err
}

func (d *EtcdStateDriver) ClearState(key string) error {
	_, err := d.Client.Delete(key, false)
	return err
}

func (d *EtcdStateDriver) ReadState(key string, value core.State,
	unmarshal func([]byte, interface{}) error) error {
	encodedState, err := d.Read(key)
	if err != nil {
		return err
	}

	err = unmarshal(encodedState, value)
	if err != nil {
		return err
	}

	return nil
}

func (d *EtcdStateDriver) WriteState(key string, value core.State,
	marshal func(interface{}) ([]byte, error)) error {
	encodedState, err := marshal(value)
	if err != nil {
		return err
	}

	err = d.Write(key, encodedState)
	if err != nil {
		return err
	}

	return nil
}
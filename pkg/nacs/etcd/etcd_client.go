package etcd

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// etcd URI format: etcd://host:2379?namespace=$namespace&service=$service&version=$version
type BaseEtcdClient struct {
	refCnt             atomic.Int32
	client             *clientv3.Client
	maxCallSendMsgSize int
	maxCallRecvMsgSize int
}

// NewEtcdClient creates a new etcd-backed client.
func NewBaseEtcdClient(u *url.URL) (*BaseEtcdClient, error) {
	maxCallSendMsgSize := 1024 * 1024 * 5
	if str := u.Query().Get("MaxCallSendMsgSize"); str != "" {
		if val, err := strconv.Atoi(str); err == nil {
			maxCallSendMsgSize = val
		}
	}

	maxCallRecvMsgSize := 1024 * 1024 * 5
	if str := u.Query().Get("MaxCallRecvMsgSize"); str != "" {
		if val, err := strconv.Atoi(str); err == nil {
			maxCallRecvMsgSize = val
		}
	}

	password, _ := u.User.Password()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:          strings.Split(u.Host, ","),
		DialTimeout:        5 * time.Second,
		MaxCallSendMsgSize: maxCallSendMsgSize,
		MaxCallRecvMsgSize: maxCallRecvMsgSize,
		Username:           u.User.Username(),
		Password:           password,
	})
	if err != nil {
		return nil, err
	}

	c := &BaseEtcdClient{
		client:             cli,
		maxCallSendMsgSize: maxCallSendMsgSize,
		maxCallRecvMsgSize: maxCallRecvMsgSize,
	}

	c.refCnt.Add(1)
	return c, nil
}

func (c *BaseEtcdClient) Spawn(namespace, service, version string) (*EtcdClient, error) {
	if namespace == "" {
		return nil, errors.New("namespace is empty")
	}
	if service == "" {
		return nil, errors.New("service is empty")
	}
	if version == "" {
		return nil, errors.New("version is empty")
	}

	c.refCnt.Add(1)
	return &EtcdClient{
		namespace:  namespace,
		service:    service,
		version:    version,
		ttl:        10,
		baseClient: c,
	}, nil
}

func (c *BaseEtcdClient) MaxCallSendMsgSize() int { return c.maxCallSendMsgSize }
func (c *BaseEtcdClient) MaxCallRecvMsgSize() int { return c.maxCallRecvMsgSize }

func (c *BaseEtcdClient) Close() error {
	if c.refCnt.Add(-1) == 0 {
		return c.client.Close()
	}
	return nil
}

func (c *BaseEtcdClient) MustClose() {
	c.refCnt.Add(-10000) // set a big negative number to make sure it will be closed
	c.client.Close()
}

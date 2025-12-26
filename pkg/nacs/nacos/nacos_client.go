package nacos

import (
	"net/url"
	"strconv"
	"sync/atomic"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/pkg/errors"
)

type BaseNacosClient struct {
	refCnt       atomic.Int32
	namingClient naming_client.INamingClient
	configClient config_client.IConfigClient

	// Namespace is the namespace of the nacos client
	// If not specified, use default namespace. It's immutable once created
	namespace string
}

// nacos://$user:$password@$host:$port?namespace=$namespace&app=$app
func NewBaseNacosClient(u *url.URL) (*BaseNacosClient, error) {

	query := u.Query()
	namespace := query.Get("namespace")
	if namespace == "" {
		namespace = constant.DEFAULT_NAMESPACE_ID
	}

	app := query.Get("app")
	if app == "" {
		app = "default"
	}

	password, _ := u.User.Password()
	cc := constant.NewClientConfig(
		constant.WithNamespaceId(namespace),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithAppName(app),
		constant.WithUsername(u.User.Username()),
		constant.WithPassword(password),
		constant.WithTimeoutMs(3000),
		constant.WithUpdateThreadNum(2),
	)

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}

	sc := []constant.ServerConfig{{
		IpAddr: u.Hostname(),
		Port:   uint64(port),
	}}

	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  cc,
		ServerConfigs: sc,
	})
	if err != nil {
		return nil, err
	}

	configClient, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  cc,
		ServerConfigs: sc,
	})
	if err != nil {
		return nil, err
	}

	c := &BaseNacosClient{
		namingClient: namingClient,
		configClient: configClient,
		namespace:    namespace,
	}
	c.refCnt.Add(1)
	return c, nil
}

func (c *BaseNacosClient) Namespace() string { return c.namespace }

// Spawn will create a new client with different group, service, version
// But shared the same instance of namingClient and configClient
// Notice, namingClient and configClient will be close only when the reference count is 0
func (c *BaseNacosClient) Spawn(group, service, version string) (*NacosClient, error) {
	if group == "" {
		group = constant.DEFAULT_GROUP
	}
	if service == "" {
		return nil, errors.New("service is empty")
	}
	if version == "" {
		return nil, errors.New("version is empty")
	}

	c.refCnt.Add(1)
	return &NacosClient{
		inGroup:    group,
		inService:  service,
		inVersion:  version,
		baseClient: c,
	}, nil
}

func (c *BaseNacosClient) Close() error {
	if c.refCnt.Add(-1) == 0 {
		c.namingClient.CloseClient()
		c.configClient.CloseClient()
	}
	return nil
}

func (c *BaseNacosClient) MustClose() {
	c.refCnt.Add(-10000) // set a big negative number to make sure it will be closed
	c.namingClient.CloseClient()
	c.configClient.CloseClient()
}

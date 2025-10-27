package nacos

import (
	"sync/atomic"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
)

type refNamingClient struct {
	naming_client.INamingClient
	ref atomic.Int32
}

func (r *refNamingClient) Close() {
	if r.ref.Add(-1) == 0 {
		r.INamingClient.CloseClient()
	}
}

func (r *refNamingClient) MustClose() {
	r.ref.Add(-100000000) // set a big negative number to make sure it will be closed
	r.INamingClient.CloseClient()
}

func (r *refNamingClient) Ref() int32 {
	return r.ref.Load()
}

func (r *refNamingClient) Share() *refNamingClient {
	r.ref.Add(1)
	return r
}

type refConfigClient struct {
	config_client.IConfigClient
	ref atomic.Int32
}

func (r *refConfigClient) Close() {
	if r.ref.Add(-1) == 0 {
		r.IConfigClient.CloseClient()
	}
}

func (r *refConfigClient) MustClose() {
	r.ref.Add(-100000000) // set a big negative number to make sure it will be closed
	r.IConfigClient.CloseClient()
}

func (r *refConfigClient) Ref() int32 {
	return r.ref.Load()
}

func (r *refConfigClient) Share() *refConfigClient {
	r.ref.Add(1)
	return r
}

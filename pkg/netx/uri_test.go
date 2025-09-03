package netx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURI(t *testing.T) {
	uri, err := ParseURI("mysql://admin:123456@192.168.1.10:3306,www.baidu.com:3306/mydb?charset=utf8")
	if err != nil {
		fmt.Println("解析失败:", err)
		return
	}

	assert.Equal(t, "mysql", uri.Scheme)
	assert.Equal(t, "admin:123456", uri.User+":"+uri.Pass)
	assert.Equal(t, "192.168.1.10:3306", uri.Hosts[0].Raw)
	assert.Equal(t, "www.baidu.com:3306", uri.Hosts[1].Raw)
	assert.Equal(t, "/mydb", uri.Path)
	assert.Equal(t, "utf8", uri.Query.Get("charset"))
}

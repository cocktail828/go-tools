package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator(t *testing.T) {
	validDSL := DSL{
		Syntax:  "v1",
		Project: "demo",
		Services: []ServiceAST{
			{
				Interceptors: []string{"log", "recover"},
				Groups: []GroupAST{
					{
						Prefix:       "api",
						Interceptors: []string{"auth"},
						Routes: []RouteAST{
							{
								Method:   "POST",
								Path:     "/user/login",
								Request:  "LoginReq",
								Response: "LoginResp",
							},
						},
					},
				},
			},
		},
	}
	assert.NoError(t, Validate(validDSL))

	invalidDSL := DSL{
		Syntax:  "", // 违反required
		Project: "demo",
		Services: []ServiceAST{ // 违反len=1（此处有2个元素）
			{
				Interceptors: []string{"", "recover"}, // 第一个元素违反required
				Groups:       []GroupAST{},            // 违反min=1
			},
			{
				Interceptors: []string{"metric"},
				Groups: []GroupAST{
					{
						Prefix:       "",           // 违反required
						Interceptors: []string{""}, // 违反required
						Routes: []RouteAST{
							{
								Method:   "", // 违反required
								Path:     "/user/login",
								Request:  "", // 违反required
								Response: "LoginResp",
							},
						},
					},
				},
			},
		},
	}
	assert.Error(t, Validate(invalidDSL))
}

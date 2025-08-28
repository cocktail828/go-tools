package ast

import (
	"github.com/go-playground/validator/v10"
)

type DSL struct {
	Syntax   string    `validate:"required"`   // 不为空
	Project  string    `validate:"required"`   // 不为空
	Services []Service `validate:"len=1,dive"` // 仅允许有一个成员，dive表示深入校验切片元素
}

type Service struct {
	Interceptors []string `validate:"dive,required"` // 每个成员都不能为空（dive深入校验）
	Groups       []Group  `validate:"min=1,dive"`    // 至少有一个成员，dive深入校验
}

type Group struct {
	Prefix       string   `validate:"required"`      // 不为空
	Interceptors []string `validate:"dive,required"` // 每个成员都不能为空
	Routes       []Route  `validate:"min=1,dive"`    // 至少有一个成员
}

type Route struct {
	Method   string `validate:"required"` // 不为空
	Path     string `validate:"required"` // 不为空
	Request  string `validate:"required"` // 不为空
	Response string `validate:"required"` // 不为空
}

func Validate(in any) error {
	return validator.New().Struct(in)
}

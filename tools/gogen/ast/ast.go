package ast

import (
	"github.com/go-playground/validator/v10"
)

type DSL struct {
	Syntax   string       `validate:"required"`   // 不为空
	Project  string       `validate:"required"`   // 不为空
	Services []ServiceAST `validate:"len=1,dive"` // 仅允许有一个成员，dive表示深入校验切片元素
	Structs  []StructDef
}

type ServiceAST struct {
	Interceptors []string   `validate:"dive,required"` // 每个成员都不能为空（dive深入校验）
	Groups       []GroupAST `validate:"min=1,dive"`    // 至少有一个成员，dive深入校验
}

type GroupAST struct {
	Prefix       string     `validate:"required"`      // 不为空
	Interceptors []string   `validate:"dive,required"` // 每个成员都不能为空
	Routes       []RouteAST `validate:"min=1,dive"`    // 至少有一个成员
}

type RouteAST struct {
	HandlerName string `validate:"required"` // 不为空
	Method      string `validate:"required"` // 不为空
	Path        string `validate:"required"` // 不为空
	Request     string
	Response    string
}

type StructDef struct {
	Name   string
	Fields []StructField
}

type StructField struct {
	Name    string
	Type    string
	Tag     string
	Comment string
}

func Validate(in any) error {
	return validator.New().Struct(in)
}

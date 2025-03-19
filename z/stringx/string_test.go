package stringx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToCamel(t *testing.T) {
	cases := []struct{ input, want string }{
		{input: "__", want: ""},
		{input: "go_zero", want: "GoZero"},
		{input: "goZero", want: "GoZero"},
		{input: "goZero", want: "GoZero"},
		{input: "goZero_", want: "GoZero"},
		{input: "go_Zero_", want: "GoZero"},
		{input: "_go_Zero_", want: "GoZero"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, ToCamel(c.input))
	}
}

func TestToSnake(t *testing.T) {
	cases := []struct{ input, want string }{
		{input: "goZero", want: "go_zero"},
		{input: "Gozero", want: "gozero"},
		{input: "GoZero", want: "go_zero"},
		{input: "Go_Zero", want: "go_zero"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, ToSnake(c.input))
	}
}

func TestTitle(t *testing.T) {
	cases := []struct{ input, want string }{
		{input: "go zero", want: "Go Zero"},
		{input: "goZero", want: "GoZero"},
		{input: "GoZero", want: "GoZero"},
		{input: "Gozero", want: "Gozero"},
		{input: "Go_zero", want: "Go_zero"},
		{input: "go_zero", want: "Go_zero"},
		{input: "Go_Zero", want: "Go_Zero"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, Title(c.input))
	}
}

func TestUntitle(t *testing.T) {
	cases := []struct{ input, want string }{
		{input: "go Zero", want: "go zero"},
		{input: "GoZero", want: "goZero"},
		{input: "Gozero", want: "gozero"},
		{input: "Go_zero", want: "go_zero"},
		{input: "go_zero", want: "go_zero"},
		{input: "Go_Zero", want: "go_Zero"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, Untitle(c.input))
	}
}

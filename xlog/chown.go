//go:build !linux
// +build !linux

package xlog

import (
	"os"
)

func chown(_ string, _ os.FileInfo) error {
	return nil
}

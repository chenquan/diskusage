//go:build windows

package internal

import (
	"syscall"
)

func accessDeniedSyscall(err error) bool {
	return syscall.ERROR_ACCESS_DENIED == err
}

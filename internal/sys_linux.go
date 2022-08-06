//go:build linux

package internal

func accessDeniedSyscall(err error) bool {
	return false
}

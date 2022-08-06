//go:build darwin

package internal

func accessDeniedSyscall(err error) bool {
	return false
}

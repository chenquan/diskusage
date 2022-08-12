//go:build darwin

package internal

func accessDeniedSyscall(_ error) bool {
	return false
}

func sysFilter(_ string) bool {
	return true
}

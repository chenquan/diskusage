//go:build linux

package internal

func accessDeniedSyscall(_ error) bool {
	return false
}

func sysFilter(dir string) bool {
	return dir != "/proc"
}

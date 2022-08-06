//go:build linux

package internal

import (
	"io/fs"
	"syscall"
	"time"
)

func getFileTimeInfo(fi fs.FileInfo) fileTimeInfo {
	win32FileAttributeData := fi.Sys().(*syscall.Stat_t)
	return fileTimeInfo{
		createTime: time.Unix(win32FileAttributeData.Ctim.Unix()),
		modifyTime: time.Unix(win32FileAttributeData.Mtim.Unix()),
	}
}

func accessDeniedSyscall(err error) bool {
	return false
}

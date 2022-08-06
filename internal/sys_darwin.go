//go:build darwin

package internal

import (
	"io/fs"
	"syscall"
	"time"
)

func getFileTimeInfo(fi fs.FileInfo) fileTimeInfo {
	statT := fi.Sys().(*syscall.Stat_t)
	return fileTimeInfo{
		createTime: time.Unix(statT.Ctimespec.Unix()),
		modifyTime: time.Unix(statT.Mtimespec.Unix()),
	}
}

func accessDeniedSyscall(err error) bool {
	return false
}

//go:build windows

package internal

import (
	"io/fs"
	"syscall"
	"time"
)

func getFileTimeInfo(fi fs.FileInfo) fileTimeInfo {
	win32FileAttributeData := fi.Sys().(*syscall.Win32FileAttributeData)
	return fileTimeInfo{
		createTime: time.Unix(0, win32FileAttributeData.CreationTime.Nanoseconds()),
		modifyTime: time.Unix(0, win32FileAttributeData.LastWriteTime.Nanoseconds()),
	}
}

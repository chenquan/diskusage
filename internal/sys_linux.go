//go:build linux

//   Copyright 2023 chenquan
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package internal

import (
	"os"
	"syscall"
)

func sysFilter(dir string) bool {
	return dir != "/proc"
}

// diskSize returns the actual number of bytes allocated on disk for the file,
// i.e. allocated blocks (st_blocks * 512), matching `du`'s default behavior.
// For sparse files this is smaller than the apparent logical size.
func diskSize(info os.FileInfo, _ string) int64 {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return stat.Blocks * 512 // st_blocks is always in 512-byte units (POSIX)
	}
	return info.Size()
}

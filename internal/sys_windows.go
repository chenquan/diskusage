//go:build windows

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
	"unsafe"
)

var (
	modKernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGetCompressedFileSizeW = modKernel32.NewProc("GetCompressedFileSizeW")
)

func sysFilter(_ string) bool {
	return true
}

// diskSize returns the actual number of bytes allocated on disk for the file
// via GetCompressedFileSizeW, matching `du`'s default behavior. For sparse or
// compressed files this excludes unallocated holes, so it is smaller than the
// apparent logical size returned by os.FileInfo.Size().
func diskSize(info os.FileInfo, name string) int64 {
	p, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return info.Size()
	}

	var high uint32
	low, _, _ := procGetCompressedFileSizeW.Call(
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(&high)),
	)
	if uint32(low) == 0xFFFFFFFF { // INVALID_FILE_SIZE → call failed
		return info.Size()
	}

	return int64(uint64(high)<<32 | uint64(uint32(low)))
}

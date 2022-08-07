//   Copyright 2022 chenquan
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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/chenquan/diskusage/internal/worker"
	"github.com/fatih/color"
	flag "github.com/spf13/pflag"

	"github.com/spf13/cobra"
)

type file struct {
	sub        []file
	name       string
	isDir      bool
	size       int64
	modifyTime time.Time
	mode       fs.FileMode
}

const (
	Bytes = 1
	KB    = 1024
	MB    = KB * 1024
	GB    = MB * 1024
	TB    = GB * 1024
)

var (
	errChan           = make(chan error)
	errorAccessDenied = errors.New("access denied")
	units             = []int{Bytes, KB, MB, GB, TB}
	unitStrings       = []string{"B", "K", "M", "G", "T"}
	w                 = worker.New(5120)
)

func Stat(cmd *cobra.Command, _ []string) error {
	flags := cmd.Flags()
	dir, err := flags.GetString("dir")
	if err != nil {
		return err
	}

	depth, err := flags.GetInt("depth")
	if err != nil {
		return err
	}

	unit, err := getUnit(flags)
	if err != nil {
		return err
	}

	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}

	types, err := flags.GetStringSlice("type")
	if err != nil {
		return err
	}
	typeMap := make(map[string]struct{}, len(types))
	for _, s := range types {
		typeMap["."+s] = struct{}{}
	}

	filter, err := flags.GetString("filter")
	if err != nil {
		return err
	}
	compile, err := regexp.Compile(filter)
	if err != nil {
		return err
	}

	go func() {
		files, err := find(dir, func(info fs.FileInfo) bool {
			if info.IsDir() {
				return true
			}

			name := info.Name()
			ext := filepath.Ext(name)
			_, ok := typeMap[ext]
			typeB := ok || len(types) == 0

			filterB := compile.MatchString(name)

			return typeB && filterB
		})
		if err != nil {
			errChan <- err
		}

		totalSize := int64(0)
		for _, f := range files {
			totalSize += f.size
		}

		header := fmt.Sprintf("total size:%s\tdir:%s", getReduce(unit, totalSize), color.HiGreenString(dir))
		colorPrintln(header)
		colorPrintln(strings.Repeat("-", len(header)+2))
		printFiles(files, 0, depth, unit)
		errChan <- nil
	}()

	err = <-errChan
	if err != nil {
		return err
	}

	return nil
}

func getUnit(flags *flag.FlagSet) (string, error) {
	unit, err := flags.GetString("unit")
	if err != nil {
		return "", err
	}

	switch unit {
	case "B", "K", "M", "G", "T":
		return unit, err
	default:
		return "", errors.New("invalid unit")
	}
}

func find(dir string, filter func(info fs.FileInfo) bool) ([]file, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		if pathError, ok := err.(*fs.PathError); ok {
			// currently only supports Windows
			if accessDeniedSyscall(pathError.Err) {
				return nil, errorAccessDenied
			}
		}
		return nil, err
	}

	var wg = sync.WaitGroup{}
	fileChan := make(chan file, len(dirEntries))
	for _, entry := range dirEntries {
		entry := entry
		fileInfo, err := entry.Info()
		if err != nil {
			return nil, err
		}

		if !filter(fileInfo) {
			continue
		}

		if !entry.IsDir() {
			fileChan <- file{
				name:       entry.Name(),
				size:       fileInfo.Size(),
				modifyTime: fileInfo.ModTime(),
				mode:       fileInfo.Mode(),
			}
			continue
		}

		wg.Add(1)
		do := func() {
			defer wg.Done()

			subFiles, err := find(path.Join(dir, entry.Name()), filter)
			if err != nil {
				if accessDenied(err) {
					return
				}
				errChan <- err
			}

			totalSize := int64(0)
			for _, subFile := range subFiles {
				totalSize += subFile.size
			}

			fileChan <- file{
				sub:        subFiles,
				name:       entry.Name(),
				isDir:      true,
				size:       totalSize,
				modifyTime: fileInfo.ModTime(),
				mode:       fileInfo.Mode(),
			}
		}
		w.Run(do)
	}
	wg.Wait()
	close(fileChan)

	files := make([]file, 0, len(dirEntries))
	for f := range fileChan {
		files = append(files, f)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })

	return files, nil
}

func printFiles(files []file, n, depth int, unit string) {
	if n == depth {
		return
	}

	bar := strings.Repeat("   ", n) + "|--"
	for _, f := range files {

		part1 := fmt.Sprintf("%s\t%s\t%s", f.modifyTime.Format("20060102 15:04:05"), f.mode, getReduce(unit, f.size))
		part2 := color.HiGreenString(f.name)
		var s = bar
		if f.isDir {
			s += color.HiBlueString(part1) + "\t" + part2
		} else {
			s += part1 + "\t" + part2
		}
		colorPrintln(s)

		if f.isDir {
			printFiles(f.sub, n+1, depth, unit)
		}
	}
}

func getReduce(unit string, n int64) string {
	reduce := 0
	switch unit {
	case "B":
		reduce = 0
	case "K":
		reduce = 1
	case "M":
		reduce = 2
	case "G":
		reduce = 3
	case "T":
		reduce = 4
	}
	for {
		if reduce <= 0 {
			break
		}

		if int(float64(n)/float64(units[reduce])*1000) > 0 {
			break
		}

		reduce--

	}

	return fmt.Sprintf("%0.3f%s", float64(n)/float64(units[reduce]), unitStrings[reduce])
}

func colorPrintln(a ...any) {
	_, _ = fmt.Fprintln(color.Output, a...)
}

func accessDenied(err error) bool {
	return err == errorAccessDenied
}

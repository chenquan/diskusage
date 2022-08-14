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
	"strconv"
	"strings"
	"sync"

	"github.com/chenquan/diskusage/internal/worker"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/list"
	flag "github.com/spf13/pflag"

	"github.com/spf13/cobra"
)

type (
	file struct {
		sub   []file
		name  string
		isDir bool
		size  int64
	}

	infoFile struct {
		size      float64
		str       string
		usageRate float64
		uint      string
		isDir     bool
	}
)

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

	all, err := flags.GetBool("all")
	if err != nil {
		return err
	}

	err = handleColor(flags)
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

		val, reduceUnit := getReduce(unit, totalSize)
		header := fmt.Sprintf("Total: %0.3f%s\t%s", val, reduceUnit, color.HiGreenString(dir))
		colorPrintln(header)
		colorPrintln(strings.Repeat("-", len(header)+2))

		l := list.NewWriter()
		l.SetStyle(list.StyleConnectedLight)

		infoFiles := buildInfoFile(l, files, 0, depth, unit, totalSize, all)
		maxLen := 0
		for _, info := range infoFiles {
			size := len(info.str)
			if maxLen < size {
				maxLen = size
			}
		}
		printTree(l.Render(), infoFiles, maxLen)

		errChan <- nil
	}()

	err = <-errChan
	if err != nil {
		return err
	}

	return nil
}

func handleColor(flags *flag.FlagSet) error {
	colorVal, err := flags.GetString("color")
	if err != nil {
		return err
	}

	switch colorVal {
	case "auto":
	case "always":
		color.NoColor = false
	case "ignore":
		color.NoColor = true
	default:
		return errors.New("invalid color mode")
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
	if !sysFilter(dir) {
		return nil, nil
	}

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
				name: entry.Name(),
				size: fileInfo.Size(),
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
				sub:   subFiles,
				name:  entry.Name(),
				isDir: true,
				size:  totalSize,
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
	sort.Slice(files, func(i, j int) bool { return files[i].size > files[j].size })

	return files, nil
}

func buildInfoFile(l list.Writer, files []file, n, depth int, unit string, totalSize int64, all bool) []infoFile {
	if n == depth {
		return nil
	}

	var infoFiles []infoFile
	for _, f := range files {
		if f.isDir && f.size == 0 && !all {
			continue
		}

		val, reduceUnit := getReduce(unit, f.size)
		infoFiles = append(infoFiles, infoFile{
			size:      val,
			uint:      reduceUnit,
			usageRate: float64(f.size) / float64(totalSize) * 100,
			str:       fmt.Sprintf("%0.1f", val),
			isDir:     f.isDir,
		})

		name := f.name
		if f.isDir {
			name = color.HiGreenString(name)
		}
		l.AppendItem(name)

		if f.isDir {
			l.Indent()
			subUsageSizes := buildInfoFile(l, f.sub, n+1, depth, unit, totalSize, all)
			infoFiles = append(infoFiles, subUsageSizes...)
			l.UnIndent()
		}
	}

	return infoFiles
}

func getReduce(unit string, n int64) (float64, string) {
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

		if int(float64(n)/float64(units[reduce])*10) > 0 {
			break
		}

		reduce--

	}

	return float64(n) / float64(units[reduce]), unitStrings[reduce]
}

func colorPrintln(a ...any) {
	_, _ = fmt.Fprintln(color.Output, a...)
}

func accessDenied(err error) bool {
	return err == errorAccessDenied
}

func printTree(content string, infoFiles []infoFile, maxLen int) {
	size := len(infoFiles)
	for i, line := range strings.Split(content, "\n") {
		if i >= size {
			continue
		}

		info := infoFiles[i]

		str := " %" + strconv.Itoa(maxLen) + ".1f%s %4.1f%%"
		str = fmt.Sprintf(str, info.size, info.uint, info.usageRate)
		if info.isDir {
			str = color.HiRedString(str)
		}

		colorPrintln(str, line)
	}
	colorPrintln()
}

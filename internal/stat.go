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
	"bytes"
	clist "container/list"
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

const (
	Bytes = 1
	KB    = 1024
	MB    = KB * 1024
	GB    = MB * 1024
	TB    = GB * 1024
)

var (
	errChan     = make(chan error)
	units       = []int64{Bytes, KB, MB, GB, TB}
	unitStrings = []string{"B", "K", "M", "G", "T"}
	w           *worker.Worker
	out         = &bytes.Buffer{}
)

type (
	file struct {
		sub   []*file
		name  string
		isDir bool
		size  int64
		print bool
	}

	fileInfo struct {
		size      float64
		strLen    int
		usageRate float64
		uint      string
		isDir     bool
	}
)

func Stat(cmd *cobra.Command, _ []string) error {
	flags := cmd.Flags()
	dir, err := flags.GetString("dir")
	if err != nil {
		return err
	}

	depth, err := flags.GetInt64("depth")
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

	regexpFilter, err := genRegexpFilter(filter)
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

	err = setWorker(flags)
	if err != nil {
		return err
	}

	limit, err := flags.GetInt64("limit")
	if err != nil {
		return err
	}

	recursion, err := flags.GetBool("recursion")
	if err != nil {
		return err
	}

	directory, err := getDirectory(flags)
	if err != nil {
		return err
	}

	interactive, err := flags.GetBool("interactive")
	if err != nil {
		return err
	}

	go func() {
		defer close(errChan)

		files, err := find(dir, func(info fs.FileInfo) bool {
			if info.IsDir() {
				return true
			}

			name := info.Name()
			ext := filepath.Ext(name)
			_, ok := typeMap[ext]
			typeB := ok || len(types) == 0

			filterB := regexpFilter(name)

			return typeB && filterB
		})
		if err != nil {
			errChan <- err
			return
		}

		totalSize := int64(0)
		for _, f := range files {
			totalSize += f.size
		}

		val, reduceUnit := getReduce(unit, totalSize)
		header := fmt.Sprintf("Total: %0.3f%s\t%s", val, reduceUnit, color.HiGreenString(dir))
		colorPrintln(header)
		colorPrintln(strings.Repeat("─", len(header)+2))

		l := list.NewWriter()
		l.SetStyle(list.StyleConnectedLight)

		markPrint(files, limit, all, directory)
		infoFiles := buildInfoFile(l, files, 0, depth, unit, totalSize, recursion)
		maxLen := 0
		for _, info := range infoFiles {
			if maxLen < info.strLen {
				maxLen = info.strLen
			}
		}
		printTree(l.Render(), infoFiles, maxLen)

		rendering(interactive, out.String())
		errChan <- nil
	}()

	if err := <-errChan; err != nil {
		return err
	}

	return nil
}

func setWorker(flags *flag.FlagSet) error {
	workerNum, err := flags.GetInt("worker")
	if err != nil {
		return err
	}

	w = worker.New(workerNum)

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
		return unit, nil
	default:
		return "", errors.New("invalid unit:" + unit)
	}
}

func getDirectory(flags *flag.FlagSet) (bool, error) {
	directory, err := flags.GetBool("directory")
	if err != nil {
		return false, err
	}

	return directory, nil
}

func find(dir string, filter func(info fs.FileInfo) bool) ([]*file, error) {
	if !sysFilter(dir) {
		return nil, nil
	}

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("no such directory")
		}

		return nil, nil
	}

	var wg = sync.WaitGroup{}
	fileChan := make(chan *file, len(dirEntries))
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
			fileChan <- &file{
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
				return
			}

			totalSize := int64(0)
			for _, subFile := range subFiles {
				totalSize += subFile.size
			}

			fileChan <- &file{
				sub:   subFiles,
				name:  entry.Name(),
				isDir: true,
				size:  totalSize,
			}
		}
		w.Run(do)
	}
	wg.Wait()
	w.Close()
	close(fileChan)

	files := make([]*file, 0, len(dirEntries))
	for f := range fileChan {
		files = append(files, f)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].size > files[j].size })

	return files, nil
}

func buildInfoFile(l list.Writer, files []*file, n, depth int64, unit string, totalSize int64, recursion bool) []fileInfo {
	if n == depth && !recursion {
		return nil
	}

	var infoFiles []fileInfo
	for _, f := range files {
		if !f.print {
			continue
		}

		val, reduceUnit := getReduce(unit, f.size)
		infoFiles = append(infoFiles, fileInfo{
			size:      val,
			uint:      reduceUnit,
			usageRate: float64(f.size) / float64(totalSize) * 100,
			strLen:    len(fmt.Sprintf("%0.1f", val)),
			isDir:     f.isDir,
		})

		name := f.name
		if f.isDir {
			name = color.HiGreenString(name)
		}
		l.AppendItem(name)

		if f.isDir {
			l.Indent()
			subUsageSizes := buildInfoFile(l, f.sub, n+1, depth, unit, totalSize, recursion)
			infoFiles = append(infoFiles, subUsageSizes...)
			l.UnIndent()
		}
	}

	return infoFiles
}

// markPrint returns fileInfo with print flags.
func markPrint(files []*file, limit int64, all bool, directory bool) []fileInfo {
	cl := clist.New()
	pushList(cl, files)

	for cl.Len() != 0 {
		if limit <= 0 {
			break
		}

		element := cl.Back()
		cl.Remove(element)

		f := element.Value.(*file)
		if f.isDir && f.size == 0 && !all {
			continue
		}

		if !f.isDir && directory {
			// only display directory.
			continue
		}

		limit--
		f.print = true

		pushList(cl, f.sub)
	}

	return nil
}

func pushList(cl *clist.List, files []*file) {
	for _, f := range files {
		cl.PushFront(f)
	}
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
	_, _ = fmt.Fprintln(out, a...)
}

func printTree(content string, infoFiles []fileInfo, maxLen int) {
	size := len(infoFiles)
	format := " %" + strconv.Itoa(maxLen) + ".1f%s %5.1f%%"
	for i, line := range strings.Split(content, "\n") {
		if i >= size {
			continue
		}

		info := infoFiles[i]
		str := fmt.Sprintf(format, info.size, info.uint, info.usageRate)
		if info.isDir {
			str = color.HiRedString(str)
		}

		colorPrintln(str, line)
	}
	colorPrintln()
}

func genRegexpFilter(filter string) (func(str string) bool, error) {
	if filter == "" {
		return func(str string) bool {
			return true
		}, nil
	}

	compile, err := regexp.Compile(filter)
	if err != nil {
		return nil, err
	}

	return compile.MatchString, nil
}

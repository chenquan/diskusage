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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	flag "github.com/spf13/pflag"

	"github.com/spf13/cobra"
)

type (
	file struct {
		sub   []file
		name  string
		isDir bool
		size  int64
		fileTimeInfo
		mode fs.FileMode
	}
	fileTimeInfo struct {
		createTime time.Time
		modifyTime time.Time
	}
)

const (
	Bytes = 1
	KB    = 1024
	MB    = KB * 1024
	GB    = MB * 1024
	TB    = GB * 1024
)

var errChan = make(chan error)

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

	go func() {
		files, err := find(dir)
		if err != nil {
			errChan <- err
		}

		totalSize := int64(0)
		for _, f := range files {
			totalSize += f.size
		}

		header := fmt.Sprintf("total size:%s\tdir:%s", getReduce(unit, totalSize), color.GreenString(dir))
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

func find(dir string) ([]file, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var wg = sync.WaitGroup{}
	fileChan := make(chan file, len(dirEntries))
	for _, entry := range dirEntries {
		entry := entry
		fileInfo, err := entry.Info()
		timeInfo := getFileTimeInfo(fileInfo)
		if !entry.IsDir() {
			if err != nil {
				return nil, err
			}
			fileChan <- file{
				name:         entry.Name(),
				size:         fileInfo.Size(),
				fileTimeInfo: timeInfo,
				mode:         fileInfo.Mode(),
			}
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			subFiles, err := find(path.Join(dir, entry.Name()))
			if err != nil {
				errChan <- err
			}

			totalSize := int64(0)
			for _, subFile := range subFiles {
				totalSize += subFile.size
			}

			fileChan <- file{
				sub:          subFiles,
				name:         entry.Name(),
				isDir:        true,
				size:         totalSize,
				fileTimeInfo: timeInfo,
				mode:         fileInfo.Mode(),
			}
		}()
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
		s := fmt.Sprintf("%s%s\t%s\t%s\t%s", bar, f.modifyTime.Format("20060102 15:04:05"), f.mode, getReduce(unit, f.size), color.GreenString(f.name))
		if f.isDir {
			colorPrintln(color.BlueString(s))
		} else {
			colorPrintln(s)
		}

		if f.isDir {
			printFiles(f.sub, n+1, depth, unit)
		}
	}
}

var units = []int{Bytes, KB, MB, GB, TB}
var unitStrings = []string{"B", "K", "M", "G", "T"}

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

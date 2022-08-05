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
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	flag "github.com/spf13/pflag"

	"github.com/spf13/cobra"
)

type file struct {
	sub   []file
	name  string
	isDir bool
	size  int64
}

const (
	Bytes = 1
	KB    = 1024
	MB    = KB * 1024
	GB    = MB * 1024
	TB    = GB * 1024
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

	reduce := getReduce(unit)

	files, err := find(dir)
	if err != nil {
		return err
	}

	totalSize := int64(0)
	for _, f := range files {
		totalSize += f.size
	}

	header := fmt.Sprintf("total size:%0.3f%s\tdir:%s", float64(totalSize)/float64(reduce), unit, color.GreenString(dir))
	colorPrintln(header)
	colorPrintln(strings.Repeat("-", len(header)+2))

	printFiles(files, 0, depth, unit)
	return nil
}

func getUnit(flags *flag.FlagSet) (string, error) {
	unit, err := flags.GetString("unit")
	if err != nil {
		return "", err
	}

	switch unit {
	case "B":
		return "Bytes", nil
	case "K":
		return "KB", nil
	case "M":
		return "MB", nil
	case "G":
		return "GB", nil
	case "T":
		return "TB", nil
	}

	return unit, nil
}

func find(dir string) ([]file, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]file, 0, len(dirEntries))
	for _, entry := range dirEntries {
		if !entry.IsDir() {
			fileInfo, err := entry.Info()
			if err != nil {
				return nil, err
			}

			files = append(files, file{
				name:  entry.Name(),
				isDir: false,
				size:  fileInfo.Size(),
			})
			continue
		}

		subFiles, err := find(path.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}

		totalSize := int64(0)
		for _, subFile := range subFiles {
			totalSize += subFile.size
		}

		files = append(files, file{
			sub:   subFiles,
			name:  entry.Name(),
			isDir: true,
			size:  totalSize,
		})

	}

	return files, nil
}

func printFiles(files []file, n, depth int, unit string) {
	if n == depth {
		return
	}

	reduce := getReduce(unit)
	bar := strings.Repeat("    ", n) + "|---"

	for _, f := range files {
		typ := "file"
		if f.isDir {
			typ = "dir"
		}
		s := fmt.Sprintf("%stype:%s\tsize:%.3f%s\t%s", bar, typ, float64(f.size)/float64(reduce), unit, color.GreenString(f.name))
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

func getReduce(unit string) int {
	reduce := 1
	switch unit {
	case "Bytes":
		reduce = Bytes
	case "KB":
		reduce = KB
	case "MB":
		reduce = MB
	case "GB":
		reduce = GB
	case "TB":
		reduce = TB
	}

	return reduce
}
func colorPrintln(a ...any) {
	_, _ = fmt.Fprintln(color.Output, a...)
}

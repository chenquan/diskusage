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

package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/chenquan/diskusage/internal"
	"github.com/spf13/cobra"
)

const BuildVersion = "0.8.0"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "diskusage",
	Short: "A tool for showing disk usage.",
	Example: `1. The maximum display unit is GB: diskusage -u G 
2. Only files named Doc or docx are counted:
  a. diskusage -t doc,docx 
  b. diskusage -f ".+\.(doc|docx)$"
3. Supports color output to pipeline: 
  a. diskusage -c always | less -R 
  b. diskusage -c always | more
4. Displays a 2-level tree structure: diskusage -d 2
5. Specify the directory /usr: diskusage --dir /usr
6. Export disk usage to file: diskusage > diskusage.txt`,
	Long: `A tool for showing disk usage.

GitHub: https://github.com/chenquan/diskusage
Issues: https://github.com/chenquan/diskusage/issues`,
	RunE: internal.Stat,
	Version: fmt.Sprintf(
		"%s %s/%s %s", BuildVersion,
		runtime.GOOS, runtime.GOARCH, runtime.Version()),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("unit", "u", "M", "displayed units. optional: B(Bytes), K(KB), M(MB), G(GB), T(TB)")
	rootCmd.Flags().String("dir", "./", "directory path")
	rootCmd.Flags().IntP("depth", "d", 1, "shows the depth of the tree directory structure")
	rootCmd.Flags().StringSliceP("type", "t", []string{}, "only count certain types of files  (default all)")
	rootCmd.Flags().StringP("filter", "f", "", "regular expressions are used to filter files")
	rootCmd.Flags().BoolP("all", "a", false, "display all directories, otherwise only display folders whose usage size is not 0")
	rootCmd.Flags().StringP("color", "c", "auto", "set color output mode. optional: auto, always, ignore")
	rootCmd.Flags().IntP("worker", "w", 5120, "number of workers searching the directory")
}

# diskusage
<a href="https://www.buymeacoffee.com/chenquan"><img src="https://img.buymeacoffee.com/button-api/?text=Buy me a coffee&emoji=&slug=chenquan&button_colour=FFDD00&font_colour=000000&font_family=Poppins&outline_colour=000000&coffee_colour=ffffff" /></a>

[![Release](https://img.shields.io/github/v/release/chenquan/diskusage.svg?style=flat-square)](https://github.com/chenquan/diskusage)
[![Download](https://goproxy.cn/stats/github.com/chenquan/diskusage/badges/download-count.svg)](https://github.com/chenquan/diskusage)
[![GitHub](https://img.shields.io/github/license/chenquan/diskusage)](LICENSE)

English | [简体中文](README-CN.md)


💥A tool for showing disk usage. (Linux, MacOS and Windows)

![](image/linux-pipe-more.png)
![](image/only-dir.png)
![](image/interactive.png)

## 😜installation

```shell
go install github.com/chenquan/diskusage@latest
```

or [download](https://github.com/chenquan/diskusage/releases).

## 👏how to use

```
$ diskusage -h
A tool for showing disk usage.

GitHub: https://github.com/chenquan/diskusage
Issues: https://github.com/chenquan/diskusage/issues

Usage:
  diskusage [flags]

Examples:
1.The maximum display unit is GB: diskusage -u G
2.Only files named doc or docx are counted:
  a.diskusage -t doc,docx
  b.diskusage -f ".+\.(doc|docx)$"
3.Supports color output to pipeline:
  a.diskusage -c always | less -R
  b.diskusage -c always | more
4.Displays a 2-level tree structure: diskusage -d 2
5.Specify the directory /usr: diskusage --dir /usr
6.Export disk usage to file: diskusage > diskusage.txt
7.Enable interactive: diskusage -i
8.Exclude directories/files by regex: diskusage -e node_modules -e "\.log$"
9.Export as JSON: diskusage --json
10.Sort by name: diskusage --sort name

Flags:
  -a, --all             display all directories, otherwise only display folders whose usage size is not 0
  -c, --color string    set color output mode. optional: auto, always, ignore (default "auto")
  -d, --depth int       shows the depth of the tree directory structure (default 1)
      --dir string      directory path (default "./")
  -D, --directory       only display directory
  -e, --exclude strings regular expressions to exclude files and directories from the scan
  -f, --filter string   regular expressions are used to filter files
  -h, --help            help for diskusage
  -i, --interactive     enable interactive
      --json            output the result as JSON (sizes in bytes)
  -l, --limit int       limit the number of files and directories displayed (default 9223372036854775807)
      --progress        show scan progress on stderr (terminal only) (default true)
  -r, --recursion       automatically calculate directory depth, for recursively traversing all sub directories
      --sort string     sort entries by size (descending) or name (ascending) (default "size")
  -t, --type strings    only count certain types of files  (default all)
  -u, --unit string     displayed units. optional: B(Bytes), K(KB), M(MB), G(GB), T(TB) (default "M")
  -v, --version         version for diskusage
  -w, --worker int      number of workers searching the directory (default 32)
```

## 👀example

1. Only files named doc or docx are counted: `diskusage -t doc,docx` or `diskusage -f ".+\.(doc|docx)$"`
2. The maximum display unit is GB: `diskusage -u G`
3. Supports color output to pipeline: `diskusage -c always | less -R` or `diskusage -c always | more`

If you like or are using this project to learn or start your solution, please give it a star⭐. Thanks!

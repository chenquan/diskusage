# diskusage

[![Release](https://img.shields.io/github/v/release/chenquan/diskusage.svg?style=flat-square)](https://github.com/chenquan/diskusage)
[![Download](https://goproxy.cn/stats/github.com/chenquan/diskusage/badges/download-count.svg)](https://github.com/chenquan/diskusage)
[![GitHub](https://img.shields.io/github/license/chenquan/diskusage)](LICENSE)

[English](README.md) | 简体中文

一个显示磁盘使用情况的工具。 （Linux、MacOS 和 Windows）

![](image/linux-pipe-more.png)

## 😜安装

```shell
go install github.com/chenquan/diskusage@latest
```

或者 [下载](https://github.com/chenquan/diskusage/releases).

## 👏如何使用

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

Flags:
  -a, --all             display all directories, otherwise only display folders whose usage size is not 0
  -c, --color string    set color output mode. optional: auto, always, ignore (default "auto")
  -d, --depth int       shows the depth of the tree directory structure (default 1)
      --dir string      directory path (default "./")
  -f, --filter string   regular expressions are used to filter files
  -h, --help            help for diskusage
  -l, --limit int       limit the number of files and directories displayed (default 9223372036854775807)
  -t, --type strings    only count certain types of files  (default all)
  -u, --unit string     displayed units. optional: B(Bytes), K(KB), M(MB), G(GB), T(TB) (default "M")
  -v, --version         version for diskusage
  -w, --worker int      number of workers searching the directory (default 32)
```

## 👀案例

1. 只统计名为 doc 或 docx 的文件: `diskusage -t doc,docx` or `diskusage -f ".+\.(doc|docx)$"`
2. 最大显示单位GB: `diskusage -u G`
3. 支持颜色输出到管道: `diskusage -c always | less -R` or `diskusage -c always | more`

如果你喜欢或正在使用这个项目来学习或开始你的解决方案，请给它一个star⭐。谢谢！

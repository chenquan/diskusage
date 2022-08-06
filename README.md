# diskusage

> A tool for showing disk usage.

## installation

```shell
go install github.com/chenquan/diskusage@latest
```

## use

```
$ diskusage -h
A tool for showing disk usage.

Usage:
  diskusage [flags]

Flags:
  -d, --depth int       Shows the depth of the tree directory structure (default 1)
      --dir string      Dir path (default "./")
  -f, --filter string   Make regular expression filter (default ".")
  -h, --help            help for diskusage
  -t, --type strings    Only count certain types of files  (default all)
  -u, --unit string     Displayed units. optional: B(Bytes), K(KB), M(MB), G(GB), T(TB) (default "M")
  ```

![](image/cmd.png)
# diskusage

> A tool for showing disk usage.

## installation

```shell
go install github.com/chenquan/diskusage@latest
```

## how to use

```
$ diskusage -h
A tool for showing disk usage.

Usage:
  diskusage [flags]

Flags:
  -d, --depth int       shows the depth of the tree directory structure (default 1)
      --dir string      dir path (default "./")
  -f, --filter string   regular expression filter (default ".+")
  -h, --help            help for diskusage
  -t, --type strings    only count certain types of files  (default all)
  -u, --unit string     displayed units. optional: B(Bytes), K(KB), M(MB), G(GB), T(TB) (default "M")
```

![](image/cmd.png)

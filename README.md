# go-dir

纯标准库实现的目录小工具，零第三方依赖。整理磁盘、找大文件、看类型分布时用得着。

## 功能

- 找大文件：按体积排序列出最大的 N 个，可按 `-min` 限定最小体积
- 类型统计：按扩展名统计各类文件数量
- 目录树：像 `tree` 命令那样打印目录结构，可用 `-d` 限制深度

## 用法

```bash
# 最大的 10 个文件
go run . big .

# 只关心大于 1MB 的，取前 5
go run . big -top 5 -min 1024 .

# 扩展名分布
go run . types .

# 打印目录树，深度 2
go run . tree -d 2 .
```

## 编译

```bash
go build -o go-dir .
```

依赖：仅 Go 标准库（`os`、`path/filepath`、`sort`、`strings`、`fmt`）。

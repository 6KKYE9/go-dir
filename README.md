# go-dir

系统里那些零碎查询（文件大小、环境变量、随机串），终端里顺手就查了。

```bash
# 最大的 10 个文件
go run . big .

# 只关心大于 2MB 的，取前 5
go run . big -top 5 -min 2M .

# 扩展名分布
go run . types .

# 打印目录树，深度 2
go run . tree -d 2 .

# 找出空目录
go run . empty .
```

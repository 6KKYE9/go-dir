// go-dir 是一个纯标准库实现的目录小工具。
// 找大文件、按扩展名归类统计、打印目录树——这些在整理磁盘时很常用。
// 只用到 os、path/filepath、sort，不依赖任何第三方包。
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// humanSize 把字节数转成人类友好的 KB/MB/GB。
func humanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGT"[exp])
}

// cmdBig 找大文件：go-dir big [-top N] [-min 大小] <目录>
func cmdBig(args []string) {
	top := 10
	minSize := int64(0)
	dir := "."
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-top":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &top)
				i++
			}
		case "-min":
			if i+1 < len(args) {
				var m int64
				fmt.Sscanf(args[i+1], "%d", &m)
				minSize = m * 1024 // 参数以 KB 计
				i++
			}
		default:
			dir = args[i]
		}
	}
	type entry struct {
		path string
		size int64
	}
	var files []entry
	err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过无法访问的
		}
		if info.Mode().IsRegular() && info.Size() >= minSize {
			files = append(files, entry{p, info.Size()})
		}
		return nil
	})
	if err != nil {
		fmt.Println("遍历失败:", err)
		return
	}
	sort.Slice(files, func(i, j int) bool { return files[i].size > files[j].size })
	if top > len(files) {
		top = len(files)
	}
	fmt.Printf("目录 %s 下最大的 %d 个文件：\n", dir, top)
	for _, f := range files[:top] {
		fmt.Printf("%10s  %s\n", humanSize(f.size), f.path)
	}
}

// cmdTypes 按扩展名归类统计：go-dir types <目录>
func cmdTypes(args []string) {
	dir := "."
	if len(args) >= 1 {
		dir = args[0]
	}
	counts := map[string]int64{}
	err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || !info.Mode().IsRegular() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if ext == "" {
			ext = "(无扩展名)"
		}
		counts[ext]++
		return nil
	})
	if err != nil {
		fmt.Println("遍历失败:", err)
		return
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Printf("目录 %s 的文件类型分布：\n", dir)
	for _, k := range keys {
		fmt.Printf("%-14s %d 个\n", k, counts[k])
	}
}

// cmdTree 打印目录树：go-dir tree [-d 深度] <目录>
func cmdTree(args []string) {
	depth := 3
	dir := "."
	for i := 0; i < len(args); i++ {
		if args[i] == "-d" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &depth)
			i++
		} else {
			dir = args[i]
		}
	}
	fmt.Println(dir)
	printTree(dir, "", 0, depth)
}

// printTree 递归打印，prefix 是前缀缩进字符串。
func printTree(dir, prefix string, level, maxDepth int) {
	if level >= maxDepth {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	// 目录在前、文件在后，看起来更顺眼。
	sort.Slice(entries, func(i, j int) bool {
		di, dj := entries[i].IsDir(), entries[j].IsDir()
		if di != dj {
			return di
		}
		return entries[i].Name() < entries[j].Name()
	})
	for i, e := range entries {
		last := i == len(entries)-1
		branch := "├── "
		nextPrefix := prefix + "│   "
		if last {
			branch = "└── "
			nextPrefix = prefix + "    "
		}
		fmt.Printf("%s%s%s\n", prefix, branch, e.Name())
		if e.IsDir() {
			printTree(filepath.Join(dir, e.Name()), nextPrefix, level+1, maxDepth)
		}
	}
}

func usage() {
	fmt.Print(`go-dir 目录小工具

用法:
  go-dir big   [-top N] [-min KB] <目录>   找最大的文件
  go-dir types <目录>                      按扩展名统计数量
  go-dir tree  [-d 深度] <目录>            打印目录树
`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	switch os.Args[1] {
	case "big":
		cmdBig(os.Args[2:])
	case "types":
		cmdTypes(os.Args[2:])
	case "tree":
		cmdTree(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Println("未知命令:", os.Args[1])
		usage()
	}
}

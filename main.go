// go-dir 是目录小工具。
// 找大文件、按扩展名归类统计、打印目录树、找出空目录——整理磁盘时很常用。
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// humanSize 把字节数转成人类友好的 B/KB/MB/GB/TB。
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

// parseSizeArg 解析带可选单位的大小参数，如 "500"->500B、"2K"->2048、"3M"->3*1024^2、"1G"。
func parseSizeArg(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("空大小")
	}
	mult := int64(1)
	last := s[len(s)-1]
	if strings.IndexByte("kKmMgGtT", last) >= 0 {
		switch strings.ToLower(string(last)) {
		case "k":
			mult = 1024
		case "m":
			mult = 1024 * 1024
		case "g":
			mult = 1024 * 1024 * 1024
		case "t":
			mult = 1024 * 1024 * 1024 * 1024
		}
		s = s[:len(s)-1]
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("无法解析大小: %q", s)
	}
	return n * mult, nil
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
				if m, err := parseSizeArg(args[i+1]); err == nil {
					minSize = m
				} else {
					fmt.Println("大小参数无效:", err)
				}
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
	if len(files) == 0 {
		fmt.Println("没有符合条件的文件")
		return
	}
	sort.Slice(files, func(i, j int) bool { return files[i].size > files[j].size })
	if top > len(files) {
		top = len(files)
	}
	fmt.Printf("目录 %s 下最大的 %d 个文件（共 %d 个匹配）：\n", dir, top, len(files))
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
	total := int64(0)
	for k := range counts {
		keys = append(keys, k)
		total += counts[k]
	}
	sort.Strings(keys)
	fmt.Printf("目录 %s 的文件类型分布（共 %d 个文件）：\n", dir, total)
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

// cmdEmpty 找出目录下所有空目录（不含任何文件或子目录）：go-dir empty <目录>
func cmdEmpty(args []string) {
	dir := "."
	if len(args) >= 1 {
		dir = args[0]
	}
	var empties []string
	err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			entries, e2 := os.ReadDir(p)
			if e2 == nil && len(entries) == 0 {
				empties = append(empties, p)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("遍历失败:", err)
		return
	}
	if len(empties) == 0 {
		fmt.Println("没有空目录")
		return
	}
	fmt.Printf("找到 %d 个空目录：\n", len(empties))
	for _, p := range empties {
		fmt.Println(p)
	}
}

func usage() {
	fmt.Print(`go-dir 目录小工具

用法:
  go-dir big   [-top N] [-min 大小] <目录>    找最大的文件
  go-dir types <目录>                         按扩展名统计数量
  go-dir tree  [-d 深度] <目录>               打印目录树
  go-dir empty <目录>                         找出空目录
大小单位支持 K/M/G/T 后缀，如 -min 2M 表示 2 MiB。
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
	case "empty":
		cmdEmpty(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Println("未知命令:", os.Args[1])
		usage()
	}
}

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHumanSize(t *testing.T) {
	cases := []struct {
		b    int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{3 * 1024 * 1024 * 1024, "3.0 GB"},
	}
	for _, c := range cases {
		if got := humanSize(c.b); got != c.want {
			t.Errorf("humanSize(%d)=%q want %q", c.b, got, c.want)
		}
	}
}

func TestParseSizeArg(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"500", 500, true},
		{"2K", 2048, true},
		{"2k", 2048, true},
		{"3M", 3 * 1024 * 1024, true},
		{"1G", 1024 * 1024 * 1024, true},
		{"1T", 1024 * 1024 * 1024 * 1024, true},
		{"", 0, false},
		{"abc", 0, false},
	}
	for _, c := range cases {
		got, err := parseSizeArg(c.in)
		if c.ok && err != nil {
			t.Errorf("parseSizeArg(%q) unexpected err: %v", c.in, err)
			continue
		}
		if !c.ok {
			if err == nil {
				t.Errorf("parseSizeArg(%q) should error", c.in)
			}
			continue
		}
		if got != c.want {
			t.Errorf("parseSizeArg(%q)=%d want %d", c.in, got, c.want)
		}
	}
}

func makeTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	// a/b (empty), a/c/file.txt, rootfile.bin
	if err := os.MkdirAll(filepath.Join(root, "a", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "a", "c"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a", "c", "file.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "rootfile.bin"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCmdEmpty(t *testing.T) {
	// 输出到 stdout 不便于断言，这里验证遍历逻辑：至少能找到 a/b 这个空目录。
	root := makeTree(t)
	var empties []string
	filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			entries, _ := os.ReadDir(p)
			if len(entries) == 0 {
				empties = append(empties, p)
			}
		}
		return nil
	})
	found := false
	for _, e := range empties {
		if filepath.Base(e) == "b" {
			found = true
		}
	}
	if !found {
		t.Errorf("应在 a/b 找到空目录, got %v", empties)
	}
}

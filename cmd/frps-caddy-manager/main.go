package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"frps-caddy-manager/internal/menu"
	"frps-caddy-manager/internal/service"
	"frps-caddy-manager/internal/util"
)

func main() {
	flagRoot := flag.String("root", "", "工作目录，默认为当前目录")
	flag.Parse()

	util.Must(util.EnsureRoot())
	util.Must(util.EnsureLinux())

	svcManager, err := service.Detect()
	util.Must(err)

	rootDir := *flagRoot
	if rootDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取当前目录失败: %v\n", err)
			os.Exit(1)
		}
		rootDir = wd
	} else {
		abs, err := filepath.Abs(rootDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "解析工作目录失败: %v\n", err)
			os.Exit(1)
		}
		rootDir = abs
	}

	manager := menu.NewManager(rootDir, svcManager)
	manager.Run()
}

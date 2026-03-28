package util

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var (
	errNotRoot  = errors.New("请使用 sudo 权限运行该程序")
	errNotLinux = errors.New("当前程序仅支持在 Linux (systemd) 环境运行")
)

// EnsureRoot 检查当前进程是否以 root 权限运行。
func EnsureRoot() error {
	if os.Geteuid() != 0 {
		return errNotRoot
	}
	return nil
}

// EnsureLinux 检查操作系统是否为 Linux。
func EnsureLinux() error {
	if runtime.GOOS != "linux" {
		return errNotLinux
	}
	return nil
}

// EnsureCommands 检查关键外部命令是否存在。
func EnsureCommands(cmds ...string) error {
	missing := make([]string, 0, len(cmds))
	for _, name := range cmds {
		if _, err := exec.LookPath(name); err != nil {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("缺少依赖: %v", missing)
	}
	return nil
}

// Must 是一个辅助函数，用于在 fatal 条件下终止程序。
func Must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

package install

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ConflictError 用于表示目标路径已经存在内置二进制。
type ConflictError struct {
	Message string
}

func (e ConflictError) Error() string { return e.Message }

// Installer 负责将内置二进制写入目标目录。
type Installer struct {
	RootDir     string
	FRPSDir     string
	CaddyDir    string
	FRPSBinary  string
	CaddyBinary string
}

// NewInstaller 创建内置安装器。
func NewInstaller(root string) *Installer {
	return &Installer{
		RootDir:     root,
		FRPSDir:     filepath.Join(root, "frps"),
		CaddyDir:    filepath.Join(root, "caddy"),
		FRPSBinary:  filepath.Join(root, "frps", "frps"),
		CaddyBinary: filepath.Join(root, "caddy", "caddy"),
	}
}

// DetectArch 校验当前架构。
func DetectArch() (string, error) {
	if runtime.GOOS != "linux" {
		return "", errors.New("仅支持 Linux 系统")
	}
	switch runtime.GOARCH {
	case "amd64":
		return "amd64", nil
	default:
		return "", fmt.Errorf("暂未内置 %s 架构，请在 AMD64 环境运行", runtime.GOARCH)
	}
}

// InstallEmbedded 写入内置二进制。
func (i *Installer) InstallEmbedded() error {
	if err := i.ensureDirs(); err != nil {
		return err
	}

	var conflicts []string

	if _, err := os.Stat(i.FRPSBinary); err == nil {
		conflicts = append(conflicts, fmt.Sprintf("检测到 frps 已存在: %s\n如需重新安装，请先移除该文件", i.FRPSBinary))
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("检查 frps 文件状态失败: %w", err)
	}

	if _, err := os.Stat(i.CaddyBinary); err == nil {
		conflicts = append(conflicts, fmt.Sprintf("检测到 caddy 已存在: %s\n如需重新安装，请先移除该文件", i.CaddyBinary))
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("检查 caddy 文件状态失败: %w", err)
	}

	if len(conflicts) > 0 {
		return ConflictError{Message: strings.Join(conflicts, "\n")}
	}

	if err := writeBinary(frpsBinary, i.FRPSBinary); err != nil {
		return fmt.Errorf("写入 frps 失败: %w", err)
	}

	if err := writeBinary(caddyBinary, i.CaddyBinary); err != nil {
		return fmt.Errorf("写入 caddy 失败: %w", err)
	}

	fmt.Printf("frps 已安装 -> %s\n", i.FRPSBinary)
	fmt.Printf("caddy 已安装 -> %s\n", i.CaddyBinary)
	return nil
}

func (i *Installer) ensureDirs() error {
	for _, dir := range []string{i.FRPSDir, i.CaddyDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
		}
	}
	return nil
}

func writeBinary(data []byte, dest string) error {
	tmp, err := os.CreateTemp(filepath.Dir(dest), ".tmp")
	if err != nil {
		return err
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	if err := tmp.Chmod(0o755); err != nil {
		return fmt.Errorf("设置执行权限失败: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmp.Name(), dest); err != nil {
		return fmt.Errorf("移动文件失败: %w", err)
	}

	return nil
}

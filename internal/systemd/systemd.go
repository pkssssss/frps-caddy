package systemd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	frpsServicePath  = "/etc/systemd/system/frps.service"
	caddyServicePath = "/etc/systemd/system/caddy.service"
)

// CreateUnitFiles 写入 systemd service 文件并执行 daemon-reload。
func CreateUnitFiles(workingDir, frpsBinary, frpsConfig, caddyBinary, caddyConfig string) error {
	frpsUnit := fmt.Sprintf(`[Unit]
Description=frps Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=%s
ExecStart=%s -c %s
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
`, workingDir, frpsBinary, frpsConfig)

	caddyUnit := fmt.Sprintf(`[Unit]
Description=Caddy Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=%s
ExecStart=%s run --config %s
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
`, workingDir, caddyBinary, caddyConfig)

	if err := os.WriteFile(frpsServicePath, []byte(frpsUnit), 0o644); err != nil {
		return fmt.Errorf("写入 %s 失败: %w", frpsServicePath, err)
	}
	if err := os.WriteFile(caddyServicePath, []byte(caddyUnit), 0o644); err != nil {
		return fmt.Errorf("写入 %s 失败: %w", caddyServicePath, err)
	}

	if err := runSystemctl("daemon-reload"); err != nil {
		return fmt.Errorf("systemctl daemon-reload 失败: %w", err)
	}

	fmt.Println("已更新", frpsServicePath)
	fmt.Println("已更新", caddyServicePath)
	return nil
}

// Control 执行 systemctl 指令。
func Control(service, action string) error {
	return runSystemctl(action, service)
}

// Status 调用 systemctl status。
func Status(service string) error {
	cmd := exec.Command("systemctl", "status", service, "--no-pager")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// IsEnabled 返回 systemctl is-enabled 的输出。
func IsEnabled(service string) (string, error) {
	cmd := exec.Command("systemctl", "is-enabled", service)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return strings.TrimSpace(buf.String()), err
	}
	return strings.TrimSpace(buf.String()), nil
}

// IsActive 返回 systemctl is-active 的输出。
func IsActive(service string) (string, error) {
	cmd := exec.Command("systemctl", "is-active", service)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return strings.TrimSpace(buf.String()), err
	}
	return strings.TrimSpace(buf.String()), nil
}

func runSystemctl(args ...string) error {
	cmd := exec.Command("systemctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ServicePaths 返回两个服务文件路径。
func ServicePaths() (string, string) {
	return frpsServicePath, caddyServicePath
}

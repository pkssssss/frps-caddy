package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	openrcFRPSServicePath  = "/etc/init.d/frps"
	openrcCaddyServicePath = "/etc/init.d/caddy"
)

type openRCManager struct{}

func (openRCManager) CreateUnits(workingDir, frpsBinary, frpsConfig, caddyBinary, caddyConfig string) error {
	frpsScript := buildOpenRCRunScript("frps", workingDir, frpsBinary, "-c "+frpsConfig, "/run/frps.pid")
	if err := writeOpenRCRunScript(openrcFRPSServicePath, frpsScript); err != nil {
		return err
	}

	caddyScript := buildOpenRCRunScript("caddy", workingDir, caddyBinary, "run --config "+caddyConfig, "/run/caddy.pid")
	if err := writeOpenRCRunScript(openrcCaddyServicePath, caddyScript); err != nil {
		return err
	}

	fmt.Println("已更新", openrcFRPSServicePath)
	fmt.Println("已更新", openrcCaddyServicePath)
	return nil
}

func (openRCManager) Control(service, action string) error {
	switch action {
	case "enable":
		cmd := exec.Command("rc-update", "add", service, "default")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "disable":
		cmd := exec.Command("rc-update", "del", service, "default")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	cmd := exec.Command("rc-service", service, action)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (openRCManager) Status(service string) error {
	cmd := exec.Command("rc-service", service, "status")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (openRCManager) IsEnabled(service string) (string, error) {
	cmd := exec.Command("rc-update", "show", "default")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	output := buf.String()
	if err != nil {
		return strings.TrimSpace(output), err
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), service) {
			return "enabled", nil
		}
	}

	return "disabled", nil
}

func (openRCManager) IsActive(service string) (string, error) {
	cmd := exec.Command("rc-service", service, "status")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	output := strings.TrimSpace(buf.String())
	if err == nil {
		return "active", nil
	}

	if strings.Contains(strings.ToLower(output), "stopped") {
		return "inactive", err
	}
	return output, err
}

func (openRCManager) ServicePaths() (string, string) {
	return openrcFRPSServicePath, openrcCaddyServicePath
}

func buildOpenRCRunScript(name, workingDir, command, commandArgs, pidfile string) string {
	return fmt.Sprintf(`#!/sbin/openrc-run
description="%s service"

directory="%s"
command="%s"
command_args="%s"
command_user="root"
pidfile="%s"
command_background="yes"
start_stop_daemon_args="--make-pidfile --pidfile %s"

depend() {
    need net
}
`, name, workingDir, command, commandArgs, pidfile, pidfile)
}

func writeOpenRCRunScript(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		return fmt.Errorf("写入 %s 失败: %w", path, err)
	}

	return nil
}

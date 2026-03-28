package service

import (
	"errors"

	"frps-caddy-manager/internal/util"
)

// Manager 抽象了不同 init 系统的服务管理能力。
type Manager interface {
	CreateUnits(workingDir, frpsBinary, frpsConfig, caddyBinary, caddyConfig string) error
	Control(service, action string) error
	Status(service string) error
	IsEnabled(service string) (string, error)
	IsActive(service string) (string, error)
	ServicePaths() (string, string)
}

// Detect 返回可用的服务管理实现，优先 systemd，其次 OpenRC；都不存在则报错。
func Detect() (Manager, error) {
	if err := util.EnsureCommands("systemctl"); err == nil {
		return systemdManager{}, nil
	}

	if err := util.EnsureCommands("rc-service", "rc-update"); err == nil {
		return openRCManager{}, nil
	}

	return nil, errors.New("当前环境未检测到 systemd 或 OpenRC，不支持运行")
}

package service

import (
	"strings"

	"frps-caddy-manager/internal/systemd"
)

type systemdManager struct{}

func (systemdManager) CreateUnits(workingDir, frpsBinary, frpsConfig, caddyBinary, caddyConfig string) error {
	return systemd.CreateUnitFiles(workingDir, frpsBinary, frpsConfig, caddyBinary, caddyConfig)
}

func (systemdManager) Control(service, action string) error {
	return systemd.Control(normalizeSystemdServiceName(service), action)
}

func (systemdManager) Status(service string) error {
	return systemd.Status(normalizeSystemdServiceName(service))
}

func (systemdManager) IsEnabled(service string) (string, error) {
	return systemd.IsEnabled(normalizeSystemdServiceName(service))
}

func (systemdManager) IsActive(service string) (string, error) {
	return systemd.IsActive(normalizeSystemdServiceName(service))
}

func (systemdManager) ServicePaths() (string, string) {
	return systemd.ServicePaths()
}

func normalizeSystemdServiceName(service string) string {
	if strings.HasSuffix(service, ".service") {
		return service
	}
	return service + ".service"
}

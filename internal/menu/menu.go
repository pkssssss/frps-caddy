package menu

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"frps-caddy-manager/internal/config"
	"frps-caddy-manager/internal/install"
	"frps-caddy-manager/internal/service"
)

const bannerWidth = 60

const (
	frpsServiceName  = "frps"
	caddyServiceName = "caddy"
)

// Manager 负责处理终端菜单交互。
type Manager struct {
	RootDir   string
	Installer *install.Installer
	Config    *config.Manager
	Service   service.Manager
	Reader    *bufio.Reader
}

// NewManager 创建菜单管理器。
func NewManager(root string, svc service.Manager) *Manager {
	return &Manager{
		RootDir:   root,
		Installer: install.NewInstaller(root),
		Config:    config.NewManager(root),
		Service:   svc,
		Reader:    bufio.NewReader(os.Stdin),
	}
}

// Run 主菜单循环。
func (m *Manager) Run() {
	for {
		m.printStatusSummary()
		fmt.Println(makeBanner(" frps + caddy 管理 "))
		fmt.Println("1) 立即安装")
		fmt.Println("2) 配置设置")
		fmt.Println("3) 启动管理")
		fmt.Println("4) 开机自启")
		fmt.Println("5) 退出菜单")
		choice := m.prompt("请选择: ")

		switch choice {
		case "1":
			m.handleInstall()
		case "2":
			m.handleConfig()
		case "3":
			m.handleStartStop()
		case "4":
			m.handleAutoStart()
		case "5":
			fmt.Println("已退出。")
			return
		default:
			fmt.Println("无效选择。")
		}
	}
}

func (m *Manager) handleInstall() {
	arch, err := install.DetectArch()
	if err != nil {
		fmt.Println("检测架构失败:", err)
		return
	}
	fmt.Printf("=== 开始安装 frps 与 caddy (架构: %s) ===\n", arch)
	if err := m.Installer.InstallEmbedded(); err != nil {
		var conflict install.ConflictError
		if errors.As(err, &conflict) {
			fmt.Println(conflict.Message)
		} else {
			fmt.Println("安装失败:", err)
		}
		return
	}
	if err := m.Config.EnsureDefaults(); err != nil {
		fmt.Println("生成默认配置失败:", err)
		return
	}

	if err := m.Service.CreateUnits(m.RootDir, m.Installer.FRPSBinary, m.Config.FRPSConfigPath, m.Installer.CaddyBinary, m.Config.CaddyConfigPath); err != nil {
		fmt.Println("生成服务配置失败:", err)
		return
	}

	m.pause()
}

func (m *Manager) handleConfig() {
	for {
		m.printStatusSummary()
		fmt.Println(makeBanner(" 配置设置 "))
		fmt.Println("1) 管理 frps 配置")
		fmt.Println("2) 管理 Caddy 配置")
		fmt.Println("3) 返回主菜单")
		choice := m.prompt("请选择: ")

		switch choice {
		case "1":
			m.manageSingleConfig("frps")
		case "2":
			m.manageSingleConfig("caddy")
		case "3":
			return
		default:
			fmt.Println("无效选择。")
		}
	}
}

func (m *Manager) manageSingleConfig(which string) {
	for {
		m.printStatusSummary()
		if which == "frps" {
			fmt.Println(makeBanner(" frps 配置 "))
		} else {
			fmt.Println(makeBanner(" Caddy 配置 "))
		}
		if which == "frps" {
			fmt.Println("1) 恢复默认配置")
		} else {
			fmt.Println("1) 下载远程配置")
		}
		fmt.Println("2) 编辑配置")
		fmt.Println("3) 查看配置")
		fmt.Println("4) 返回上一层")
		choice := m.prompt("请选择: ")

		switch choice {
		case "1":
			if which == "frps" {
				if err := m.Config.OverwriteDefault(which); err != nil {
					fmt.Println("恢复失败:", err)
				} else {
					fmt.Println("已恢复默认配置，并重新生成 auth.token。")
				}
				m.pause()
			} else {
				url := m.prompt("请输入远程配置文件 URL: ")
				if url == "" {
					fmt.Println("URL 不能为空")
					m.pause()
				} else if err := m.Config.DownloadFromRemote(which, url); err != nil {
					fmt.Println("下载失败:", err)
					m.pause()
				} else {
					m.pause()
				}
			}
		case "2":
			if err := m.Config.Edit(which); err != nil {
				fmt.Println("编辑失败:", err)
			}
		case "3":
			if err := m.Config.View(which); err != nil {
				fmt.Println("查看失败:", err)
			}
			m.pause()
		case "4":
			return
		default:
			fmt.Println("无效选择。")
		}
	}
}

func (m *Manager) handleStartStop() {
	for {
		m.printStatusSummary()
		fmt.Println(makeBanner(" 启动管理 "))
		fmt.Println("1) 启动 frps + caddy")
		fmt.Println("2) 停止 frps + caddy")
		fmt.Println("3) 重启 frps + caddy")
		fmt.Println("4) 查看 frps 状态")
		fmt.Println("5) 查看 caddy 状态")
		fmt.Println("6) 返回主菜单")
		choice := m.prompt("请选择: ")

		switch choice {
		case "1":
			frpsOK := m.runControl(frpsServiceName, "start")
			caddyOK := m.runControl(caddyServiceName, "start")
			showResult("启动", frpsOK, caddyOK)
			m.pause()
		case "2":
			frpsOK := m.runControl(frpsServiceName, "stop")
			caddyOK := m.runControl(caddyServiceName, "stop")
			showResult("停止", frpsOK, caddyOK)
			m.pause()
		case "3":
			frpsOK := m.runControl(frpsServiceName, "restart")
			caddyOK := m.runControl(caddyServiceName, "restart")
			showResult("重启", frpsOK, caddyOK)
			m.pause()
		case "4":
			if err := m.Service.Status(frpsServiceName); err != nil {
				fmt.Println("查看状态失败:", err)
			}
			m.pause()
		case "5":
			if err := m.Service.Status(caddyServiceName); err != nil {
				fmt.Println("查看状态失败:", err)
			}
			m.pause()
		case "6":
			return
		default:
			fmt.Println("无效选择。")
		}
	}
}

func (m *Manager) handleAutoStart() {
	for {
		m.printStatusSummary()
		fmt.Println(makeBanner(" 开机自启 "))
		fmt.Println("1) 启用 frps + caddy 开机自启")
		fmt.Println("2) 取消 frps + caddy 开机自启")
		fmt.Println("3) 查看 frps 自启状态")
		fmt.Println("4) 查看 caddy 自启状态")
		fmt.Println("5) 返回主菜单")
		choice := m.prompt("请选择: ")

		switch choice {
		case "1":
			frpsOK := m.runControl(frpsServiceName, "enable")
			caddyOK := m.runControl(caddyServiceName, "enable")
			showAutoStartResult("启用", frpsOK, caddyOK)
			m.pause()
		case "2":
			frpsOK := m.runControl(frpsServiceName, "disable")
			caddyOK := m.runControl(caddyServiceName, "disable")
			showAutoStartResult("取消", frpsOK, caddyOK)
			m.pause()
		case "3":
			status, err := m.Service.IsEnabled(frpsServiceName)
			if err != nil {
				fmt.Println("frps 自启状态: ", status, " (", err, ")")
			} else {
				fmt.Println("frps 自启状态: ", status)
			}
			m.pause()
		case "4":
			status, err := m.Service.IsEnabled(caddyServiceName)
			if err != nil {
				fmt.Println("caddy 自启状态: ", status, " (", err, ")")
			} else {
				fmt.Println("caddy 自启状态: ", status)
			}
			m.pause()
		case "5":
			return
		default:
			fmt.Println("无效选择。")
		}
	}
}

func (m *Manager) prompt(msg string) string {
	fmt.Print(msg)
	line, _ := m.Reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func (m *Manager) pause() {
	fmt.Println()
	fmt.Println("按 Enter 返回上一层菜单...")
	m.Reader.ReadString('\n')
}

func (m *Manager) printStatusSummary() {
	frpsActive, frpsActiveErr := m.Service.IsActive(frpsServiceName)
	frpsEnabled, frpsEnabledErr := m.Service.IsEnabled(frpsServiceName)
	caddyActive, caddyActiveErr := m.Service.IsActive(caddyServiceName)
	caddyEnabled, caddyEnabledErr := m.Service.IsEnabled(caddyServiceName)
	fmt.Println(makeBanner(" 服务状态 "))
	fmt.Printf("frps  : %s | 自启: %s\n", renderActiveStatus(frpsActive, frpsActiveErr), renderEnabledStatus(frpsEnabled, frpsEnabledErr))
	fmt.Printf("caddy : %s | 自启: %s\n", renderActiveStatus(caddyActive, caddyActiveErr), renderEnabledStatus(caddyEnabled, caddyEnabledErr))
}

func showResult(action string, frpsOK, caddyOK bool) {
	if frpsOK && caddyOK {
		fmt.Printf("frps 和 caddy 已%s。\n", action)
		return
	}
	if frpsOK {
		fmt.Printf("frps %s成功。\n", action)
	} else {
		fmt.Printf("frps %s失败。\n", action)
	}
	if caddyOK {
		fmt.Printf("caddy %s成功。\n", action)
	} else {
		fmt.Printf("caddy %s失败。\n", action)
	}
}

func showAutoStartResult(action string, frpsOK, caddyOK bool) {
	if frpsOK && caddyOK {
		fmt.Printf("frps 和 caddy 开机自启已%s。\n", action)
		return
	}
	if frpsOK {
		fmt.Printf("frps 开机自启%s成功。\n", action)
	} else {
		fmt.Printf("frps 开机自启%s失败。\n", action)
	}
	if caddyOK {
		fmt.Printf("caddy 开机自启%s成功。\n", action)
	} else {
		fmt.Printf("caddy 开机自启%s失败。\n", action)
	}
}

func (m *Manager) runControl(serviceName, action string) bool {
	if err := m.Service.Control(serviceName, action); err != nil {
		fmt.Printf("执行 %s %s 失败: %v\n", action, serviceName, err)
		return false
	}
	return true
}

const (
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
)

func colorize(text, color string) string {
	return color + text + colorReset
}

func renderActiveStatus(status string, err error) string {
	if err != nil {
		msg := strings.TrimSpace(status)
		if msg == "" {
			msg = err.Error()
		}
		return colorize(fmt.Sprintf("未知(%s)", msg), colorYellow)
	}

	switch status {
	case "active":
		return colorize("运行中", colorGreen)
	case "inactive":
		return colorize("已停止", colorRed)
	case "failed":
		return colorize("失败", colorRed)
	case "activating":
		return colorize("启动中", colorYellow)
	case "deactivating":
		return colorize("停止中", colorYellow)
	default:
		if status == "" {
			return colorize("未知", colorYellow)
		}
		return colorize(status, colorYellow)
	}
}

func renderEnabledStatus(status string, err error) string {
	if err != nil {
		msg := strings.TrimSpace(status)
		if msg == "" {
			msg = err.Error()
		}
		return colorize(fmt.Sprintf("未知(%s)", msg), colorYellow)
	}

	switch status {
	case "enabled":
		return colorize("已启用", colorGreen)
	case "disabled":
		return colorize("未启用", colorRed)
	case "static":
		return colorize("静态", colorYellow)
	case "indirect":
		return colorize("间接", colorYellow)
	default:
		if status == "" {
			return colorize("未知", colorYellow)
		}
		return colorize(status, colorYellow)
	}
}

func makeBanner(title string) string {
	width := bannerWidth
	titleWidth := visualWidth(title)
	if titleWidth >= width {
		return title
	}
	totalPadding := width - titleWidth
	left := totalPadding / 2
	right := totalPadding - left
	return strings.Repeat("=", left) + title + strings.Repeat("=", right)
}

// visualWidth 粗略估算标题在终端的显示宽度，ASCII 计 1，其他字符计 2。
func visualWidth(s string) int {
	width := 0
	for _, r := range s {
		if r <= 0x7F {
			width += 1
		} else {
			width += 2
		}
	}
	return width
}

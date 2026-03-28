package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	frpsTemplate = `# frps.toml 默认示例 (TOML 格式)，请根据实际需求调整
bindAddr = "0.0.0.0"
bindPort = 7000

auth.method = "token"
auth.token = "%s"

# http/https
vhostHTTPPort = 7070

# tls
transport.tls.force = true
`

	caddyDefault = `# Caddyfile 默认示例，请按需修改
:80 {
    respond "Caddy is running"
}
`
)

const frpsTokenByteLength = 32

// Manager 负责处理配置文件操作。
type Manager struct {
	FRPSConfigPath  string
	CaddyConfigPath string
}

// NewManager 创建配置管理器。
func NewManager(root string) *Manager {
	return &Manager{
		FRPSConfigPath:  filepath.Join(root, "frps", "frps.toml"),
		CaddyConfigPath: filepath.Join(root, "caddy", "Caddyfile"),
	}
}

// EnsureDefaults 在配置不存在时写入默认模板。
func (m *Manager) EnsureDefaults() error {
	if err := ensureFile(m.FRPSConfigPath, buildFRPSDefault); err != nil {
		return err
	}
	if err := ensureFile(m.CaddyConfigPath, func() (string, error) { return caddyDefault, nil }); err != nil {
		return err
	}
	return nil
}

// OverwriteDefault 覆盖写入默认模板。
func (m *Manager) OverwriteDefault(which string) error {
	var target, content string
	switch strings.ToLower(which) {
	case "frps":
		target = m.FRPSConfigPath
		generated, err := buildFRPSDefault()
		if err != nil {
			return err
		}
		content = generated
	case "caddy":
		target = m.CaddyConfigPath
		content = caddyDefault
	default:
		return fmt.Errorf("未知配置类型: %s", which)
	}
	return os.WriteFile(target, []byte(content), 0o600)
}

// Edit 打开配置文件供编辑。
func (m *Manager) Edit(which string) error {
	var target string
	switch strings.ToLower(which) {
	case "frps":
		target = m.FRPSConfigPath
	case "caddy":
		target = m.CaddyConfigPath
	default:
		return fmt.Errorf("未知配置类型: %s", which)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		if _, err := exec.LookPath("nano"); err == nil {
			editor = "nano"
		} else if _, err := exec.LookPath("vi"); err == nil {
			editor = "vi"
		} else {
			return errors.New("未检测到可用编辑器，请设置 EDITOR 环境变量")
		}
	}

	cmd := exec.Command(editor, target)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// View 将配置内容输出到 stdout。
func (m *Manager) View(which string) error {
	var target string
	switch strings.ToLower(which) {
	case "frps":
		target = m.FRPSConfigPath
	case "caddy":
		target = m.CaddyConfigPath
	default:
		return fmt.Errorf("未知配置类型: %s", which)
	}

	data, err := os.ReadFile(target)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("配置不存在: %s", target)
	}
	if err != nil {
		return err
	}

	fmt.Println("-----", target, "-----")
	fmt.Println(string(data))
	fmt.Println("------------------------------------")
	return nil
}

func ensureFile(path string, builder func() (string, error)) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("检查 %s 失败: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	content, err := builder()
	if err != nil {
		return fmt.Errorf("生成默认配置失败: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("写入默认配置失败: %w", err)
	}

	fmt.Printf("已生成默认配置 -> %s\n", path)
	return nil
}

func buildFRPSDefault() (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(frpsTemplate, token), nil
}

// DownloadFromRemote 从远程 URL 下载配置文件。
func (m *Manager) DownloadFromRemote(which string, remoteURL string) error {
	parsedURL, err := url.Parse(remoteURL)
	if err != nil {
		return fmt.Errorf("无效的 URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("仅支持 http 和 https 协议")
	}

	var target string
	switch strings.ToLower(which) {
	case "caddy":
		target = m.CaddyConfigPath
	default:
		return fmt.Errorf("未知配置类型: %s", which)
	}

	transport := &http.Transport{
		Proxy:             http.ProxyFromEnvironment,
		ForceAttemptHTTP2: false,
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	req, err := http.NewRequest(http.MethodGet, remoteURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	// 某些 CDN/网关对 Go 默认 UA 或 HTTP/2 兼容性较差，这里显式使用 HTTP/1.1 且自定义 UA。
	req.Header.Set("User-Agent", "frps-caddy-manager/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，服务器返回状态码: %d", resp.StatusCode)
	}

	const maxContentSize = 1024 * 1024
	content, err := io.ReadAll(io.LimitReader(resp.Body, maxContentSize))
	if err != nil {
		return fmt.Errorf("读取内容失败: %w", err)
	}

	if len(content) == maxContentSize {
		return errors.New("配置文件过大，超过 1MB 限制")
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(target, content, 0o600); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("已从 %s 下载配置到 %s\n", remoteURL, target)
	return nil
}

func generateToken() (string, error) {
	buf := make([]byte, frpsTokenByteLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个用 Go 语言编写的 frps（frp 服务端）和 Caddy 管理工具，专为 Linux systemd 环境设计。项目提供交互式菜单界面，用于配置、安装和管理 frps 与 Caddy 服务。

## 构建与开发命令

### 基本构建与运行
- `go build ./cmd/frps-caddy-manager`：编译可执行文件到当前目录
- `go run ./cmd/frps-caddy-manager`：直接运行 CLI，便于交互式验证菜单
- `go run ./cmd/frps-caddy-manager -root /path/to/workdir`：指定工作目录运行

### 测试命令
- `go test ./...`：执行全部单元测试，提交前确保无失败
- `GOFLAGS="-race" go test ./...`：检查并发安全，建议在变更配置流程时使用

### 代码格式化
- `gofmt -w .`：格式化所有 Go 文件
- `goimports -w .`：整理导入并格式化（推荐）

## 项目架构

### 目录结构与职责
- `cmd/frps-caddy-manager/`：CLI 入口点，负责环境校验和菜单协调
- `internal/config/`：配置管理核心，处理 frps 与 Caddy 配置的生成、覆盖和远程下载
- `internal/install/`：二进制文件安装和初始化逻辑
- `internal/systemd/`：systemd 服务单元文件的生成和管理
- `internal/menu/`：交互式终端菜单系统
- `internal/util/`：环境校验和辅助工具函数

### 核心组件交互
1. **main.go**：启动入口，执行环境校验（root权限、Linux系统、systemctl命令）
2. **menu.Manager**：主菜单控制器，协调所有子系统
3. **config.Manager**：配置文件管理器，维护 frps.toml 和 Caddyfile
4. **install.Installer**：处理二进制文件下载、安装和初始化
5. **systemd**：负责服务单元文件的创建和服务管理

### 配置文件路径约定
- 工作目录下的 `frps/frps.toml`：frps 服务配置
- 工作目录下的 `caddy/Caddyfile`：Caddy 反向代理配置
- 系统服务路径：`/etc/systemd/system/frps.service` 和 `/etc/systemd/system/caddy.service`

## 开发规范

### 代码风格
- 使用 Go 1.22+ 特性
- 导出类型与函数采用 PascalCase，包内部工具保持 camelCase
- 错误信息保持中文提示，与现有实现一致
- 避免未使用代码，践行 YAGNI 原则

### 测试要求
- 采用 Go testing 框架，推荐表驱动测试
- 测试函数命名遵循 `Test<模块><场景>` 规范
- 目标覆盖率 ≥ 80%，覆盖 happy path 与错误处理路径
- 新增行为必须包含相应测试用例

### 提交规范
- 推荐采用 Conventional Commits（如 `feat: add caddy overwrite command`）
- PR 描述需列出变更要点、测试方式及配置影响
- 确保代码通过全部测试后再合并

## 安全与环境约束

### 权限要求
- 大部分操作需要 root 权限
- 使用 `util.EnsureRoot()` 和 `util.EnsureCommands()` 进行环境校验
- 仅支持 Linux systemd 环境

### 依赖检查
程序启动前会自动检查：
- 当前用户是否为 root
- 操作系统是否为 Linux
- systemctl 命令是否可用

## 配置模板

项目内置了 frps 和 Caddy 的默认配置模板：
- **frps.toml**：包含基本的端口绑定、Token 认证和 TLS 设置
- **Caddyfile**：简单的 HTTP 响应配置，可根据需要扩展

配置管理器支持：
- 生成随机 Token
- 覆盖现有配置
- 从远程 URL 下载配置
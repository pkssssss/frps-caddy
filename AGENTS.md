# Repository Guidelines

## 项目结构与模块组织
- `cmd/frps-caddy-manager/`：Go CLI 入口，负责解析菜单并协调内部模块。
- `internal/config/`：处理 frps 与 Caddy 配置的生成、覆盖和远程下载逻辑。
- `internal/install/` 与 `internal/systemd/`：落地二进制、生成 systemd 单元并完成服务安装。
- `internal/menu/`、`internal/util/`：终端菜单与环境校验工具，确保在 Linux systemd + root 环境运行。
- 默认配置会写入 `~/frps/` 与 `~/caddy/` 子目录，可用 `config.Manager` 的方法生成或覆盖。

## 构建、测试与开发命令
- `go build ./cmd/frps-caddy-manager`：编译可执行文件到当前目录。
- `go run ./cmd/frps-caddy-manager`：直接运行 CLI，便于交互式验证菜单。
- `go test ./...`：执行全部单元测试，提交前确保无失败。
- 建议在变更配置流程时使用 `GOFLAGS="-race" go test ./...` 检查并发安全。

## 代码风格与命名约定
- 使用 Go 1.22+，统一通过 `gofmt` / `goimports` 自动格式化，提交前运行 `gofmt -w`。
- 导出类型与函数采用 PascalCase，包内部工具保持 camelCase；避免出现未使用代码以践行 YAGNI。
- 错误信息保持中文提示，与现有实现一致；常量及模板置于对应模块内，拒绝重复定义。

## 测试指南
- 采用 Go testing 框架，推荐使用表驱动测试覆盖配置分支与系统调用失败路径。
- 测试函数命名遵循 `Test<模块><场景>`，必要时补充 `*_testdata` 辅助文件。
- 新增行为需覆盖 happy path 与错误处理，目标覆盖率 ≥ 80%；若暂不可测需在 PR 中说明原因。

## 提交与合并请求规范
- 推荐采用 Conventional Commits（如 `feat: add caddy overwrite command`），并在正文描述动机与影响范围。
- PR 描述需列出变更要点、测试方式及配置影响；若涉及 systemd 或安装脚本，请附终端录屏或命令输出。
- 确保代码通过 `go test ./...`，并在合并前获取至少一名维护者审阅。

## 安全与配置提示
- 大部分操作需 root 权限；调用 `util.EnsureRoot` / `util.EnsureCommands` 校验环境后再执行写入。
- 避免在非 Linux systemd 环境部署，如需兼容请在 issue 中说明约束与计划。

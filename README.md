# frps-caddy-manager

Linux systemd 环境下的 `frps` + `Caddy` 本地管理工具。

```bash
go build ./cmd/frps-caddy-manager
sudo ./frps-caddy-manager
```

指定工作目录：

```bash
sudo ./frps-caddy-manager -root /path/to/workdir
```

要求：`Linux`、`root`、`systemd`。

默认会在工作目录下写入 `frps/` 和 `caddy/`。

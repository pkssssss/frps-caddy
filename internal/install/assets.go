package install

import _ "embed"

//go:embed assets/frps
var frpsBinary []byte

//go:embed assets/caddy
var caddyBinary []byte

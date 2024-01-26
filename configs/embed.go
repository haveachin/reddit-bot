package configs

import _ "embed"

//go:embed config.yml
var DefaultConfig []byte

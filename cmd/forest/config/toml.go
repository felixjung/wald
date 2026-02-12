package config

import (
	"bytes"
	"fmt"

	"github.com/BurntSushi/toml"
	internalconfig "github.com/felixjung/forest/internal/config"
)

func marshalConfig(cfg *internalconfig.Config) ([]byte, error) {
	var buffer bytes.Buffer
	if err := toml.NewEncoder(&buffer).Encode(cfg); err != nil {
		return nil, fmt.Errorf("render config: %w", err)
	}
	return buffer.Bytes(), nil
}

package config

import (
	"github.com/lynx-go/lynx"
	"github.com/spf13/pflag"
)

const EnvPrefix = "LYNX"

var envBoundKeys = []string{
	"security.jwt.secret",
	"security.jwt.refresh_token_secret",
	"data.database.source",
	"data.redis.password",
	"file.buckets.default.access_key_id",
	"file.buckets.default.access_key_secret",
}

// ConfigureViper keeps Lynx file loading behavior and enables env overrides.
//
// viper's AutomaticEnv already normalizes dotted keys to upper-snake env names
// (e.g. data.database.source -> LYNX_DATA_DATABASE_SOURCE), so the env bound
// keys below are documented for clarity; AutomaticEnv covers them.
func ConfigureViper(f *pflag.FlagSet, c lynx.ConfigSource, extraPaths ...string) error {
	if err := lynx.DefaultBindConfigFunc(f, c); err != nil {
		return err
	}

	for _, path := range extraPaths {
		if path == "" {
			continue
		}
		c.AddSearchPath(path)
	}

	c.SetEnvPrefix(EnvPrefix)
	c.AutomaticEnv()

	return nil
}

func NewBindConfigFunc(extraPaths ...string) lynx.BindConfigFunc {
	return func(f *pflag.FlagSet, c lynx.ConfigSource) error {
		return ConfigureViper(f, c, extraPaths...)
	}
}

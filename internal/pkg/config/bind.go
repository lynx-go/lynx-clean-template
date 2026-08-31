package config

import (
	"strings"

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
// lynx v0.8+ passes a ConfigSource (viper-backed) instead of *viper.Viper;
// sensitive keys in envBoundKeys are bound explicitly for reliable overrides.
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

	for _, key := range envBoundKeys {
		if err := c.BindEnv(key, envName(key)); err != nil {
			return err
		}
	}

	return nil
}

// envName converts a dotted config path to its environment variable name,
// e.g. "data.database.source" -> "LYNX_DATA_DATABASE_SOURCE".
func envName(key string) string {
	return EnvPrefix + "_" + strings.ToUpper(strings.NewReplacer(".", "_", "-", "_").Replace(key))
}

func NewBindConfigFunc(extraPaths ...string) lynx.BindConfigFunc {
	return func(f *pflag.FlagSet, c lynx.ConfigSource) error {
		return ConfigureViper(f, c, extraPaths...)
	}
}

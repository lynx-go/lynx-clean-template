package config

import (
	"strings"

	"github.com/lynx-go/lynx"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const EnvPrefix = "lynx"

var envBoundKeys = []string{
	"security.jwt.secret",
	"security.jwt.refresh_token_secret",
	"data.database.source",
	"data.redis.password",
	"file.buckets.default.access_key_id",
	"file.buckets.default.access_key_secret",
}

// ConfigureViper keeps Lynx file loading behavior and enables env overrides.
func ConfigureViper(f *pflag.FlagSet, v *viper.Viper, extraPaths ...string) error {
	if err := lynx.DefaultBindConfigFunc(f, v); err != nil {
		return err
	}

	for _, path := range extraPaths {
		if path == "" {
			continue
		}
		v.AddConfigPath(path)
	}

	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	for _, key := range envBoundKeys {
		if err := v.BindEnv(key); err != nil {
			return err
		}
	}

	return nil
}

func NewBindConfigFunc(extraPaths ...string) lynx.BindConfigFunc {
	return func(f *pflag.FlagSet, v *viper.Viper) error {
		return ConfigureViper(f, v, extraPaths...)
	}
}

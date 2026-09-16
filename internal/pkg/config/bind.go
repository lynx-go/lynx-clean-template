package config

import (
	"strings"

	"github.com/lynx-go/lynx"
	"github.com/spf13/pflag"
)

const EnvPrefix = "LYNX"

// envBoundKeys 列出必须可靠支持环境变量覆盖的敏感键：这些键往往从配置
// 文件中省略（只通过环境注入），需要显式 BindEnv 才能进入 viper 的
// Unmarshal 视野；其余键依赖 SetEnvKeyReplacer + AutomaticEnv 即可覆盖。
var envBoundKeys = []string{
	"security.jwt.secret",
	"security.jwt.refresh_token_secret",
	"data.database.source",
	"data.redis.password",
	"file.buckets.default.access_key_id",
	"file.buckets.default.access_key_secret",
}

// ConfigureViper keeps Lynx file loading behavior and enables env overrides.
// lynx v1.8.0 起 ConfigSource 支持 SetEnvKeyReplacer：设置前缀与
// "." → "_" 替换规则并启用 AutomaticEnv 后，任意点分配置键都可以被
// LYNX_<DOTS_TO_UNDERSCORES> 形式的环境变量覆盖。
// 敏感键（含 map 内嵌套键，如 file.buckets.*）仍逐键 BindEnv：viper 的
// Unmarshal 基于 AllSettings，环境变量独有（配置文件缺失）的深层键需要
// 显式绑定才参与解码。
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
	c.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
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

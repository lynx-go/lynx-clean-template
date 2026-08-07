package config

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
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
// env keys are bound explicitly since the underlying viper instance is not
// exposed.
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

// UnmarshalConfig decodes the application config into out using json tags
// (the tags protoc-gen-go emits), keeping compatibility with proto-defined
// config structs. lynx v0.8+ removed the TagNameJSON option from
// Config.Unmarshal (which now always uses mapstructure tags), so the config
// tree is read field-by-field from the Config interface and decoded with
// mapstructure(json).
//
// Values are read at leaf paths (e.g. "data.database.source"): viper's env
// binding only applies to leaf lookups, so reading whole subtrees would drop
// environment overrides.
func UnmarshalConfig(c lynx.Config, out any) error {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("config: out must be a non-nil pointer")
	}
	root := map[string]any{}
	collect(c, rv.Elem(), "", root)
	if len(root) == 0 {
		return nil
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "json",
		Result:           out,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	})
	if err != nil {
		return err
	}
	return decoder.Decode(root)
}

// isStructKind reports whether t is a struct or a pointer to a struct.
func isStructKind(t reflect.Type) bool {
	if t.Kind() == reflect.Struct {
		return true
	}
	return t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Struct
}

// mapKeys returns the keys of a config subtree map (viper stores nested
// sections as map[string]any).
func mapKeys(raw any) []string {
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// collect walks the struct tree of the target and reads each leaf value from
// the Config interface. Nested struct fields and struct-typed map entries
// (e.g. file.buckets.<name>) are recursed path-wise so that leaf lookups hit
// env bindings; the rebuilt tree uses the same snake_case keys as the json
// tags, which the decoder matches recursively.
func collect(c lynx.Config, rv reflect.Value, prefix string, target map[string]any) {
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		tag, _, _ := strings.Cut(sf.Tag.Get("json"), ",")
		if tag == "" || tag == "-" || sf.PkgPath != "" {
			continue
		}
		path := tag
		if prefix != "" {
			path = prefix + "." + tag
		}

		switch {
		case isStructKind(sf.Type):
			fv := rv.Field(i)
			if fv.Kind() == reflect.Pointer {
				// proto 生成的嵌套字段是指针；配置中存在该段时也需要
				// 递归（即使目标结构体的指针尚未初始化）。
				fv = reflect.New(fv.Type().Elem())
			}
			fv = fv.Elem()
			sub := map[string]any{}
			collect(c, fv, path, sub)
			if len(sub) > 0 {
				target[tag] = sub
			}
		case sf.Type.Kind() == reflect.Map:
			keys := mapKeys(c.Get(path))
			elemType := sf.Type.Elem()
			elemValue := elemType
			if elemValue.Kind() == reflect.Pointer {
				elemValue = elemValue.Elem()
			}
			if len(keys) > 0 && isStructKind(elemType) {
				sub := map[string]any{}
				for _, k := range keys {
					inner := map[string]any{}
					collect(c, reflect.New(elemValue).Elem(), path+"."+k, inner)
					sub[k] = inner
				}
				target[tag] = sub
			} else if raw := c.Get(path); raw != nil {
				target[tag] = raw
			}
		default:
			if raw := c.Get(path); raw != nil {
				target[tag] = raw
			}
		}
	}
}

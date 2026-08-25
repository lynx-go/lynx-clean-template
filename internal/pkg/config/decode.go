package config

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/lynx-go/lynx"
)

// DecodeLynxConfig decodes the Lynx application config into out using the
// struct `json` tags. This preserves the behavior of the old
// `lynx.TagNameJSON` decoder, which the v1.5.1 Config.Unmarshal no longer
// accepts as an argument (viper would otherwise match by Go field name and
// miss snake_case keys such as `access_key_id`).
func DecodeLynxConfig(app lynx.App, out any) error {
	return DecodeLynxConfigFromSource(app.Config(), out)
}

// DecodeLynxConfigFromSource decodes a ConfigSource into out using `json` tags.
func DecodeLynxConfigFromSource(c lynx.Config, out any) error {
	var raw map[string]any
	if err := c.Unmarshal(&raw); err != nil {
		return err
	}
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "json",
		Result:           out,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return err
	}
	return dec.Decode(raw)
}

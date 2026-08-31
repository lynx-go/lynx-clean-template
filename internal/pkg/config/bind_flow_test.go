package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lynx-go/lynx"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// TestConfigureViperFullFlow 复刻 lynx initConfigure 的顺序：
// SetFlagsFunc/Parse -> BindConfigFunc(ConfigureViper) -> ReadInConfig -> BindPFlags，
// 验证 env 绑定在完整流程下仍生效。
func TestConfigureViperFullFlow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "data:\n  database:\n    source: \"\"\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_ = f.String("config-dir", "", "config dir")

	v := viper.New()
	c := lynx.NewViperConfig(v)
	if err := ConfigureViper(f, c, dir); err != nil {
		t.Fatal(err)
	}
	if err := v.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	if err := v.BindPFlags(f); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LYNX_DATA_DATABASE_SOURCE", "postgres://env:pass@host/db")

	got := v.GetString("data.database.source")
	t.Logf("got %q", got)
	if got != "postgres://env:pass@host/db" {
		t.Fatalf("env override failed, got %q", got)
	}
}

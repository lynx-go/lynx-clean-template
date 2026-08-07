package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lynx-go/lynx"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestUnmarshalConfigWithViper(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
server:
  http:
    addr: ":9099"
    timeout: "30s"
    cors:
      allow_origins: ["*"]
  grpc:
    addr: ":8088"
security:
  jwt:
    secret: "test-secret"
    token_expiry_sec: 3600
data:
  database:
    driver: pgx
    dialect: postgres
    source: "postgres://user:pass@localhost:5432/db"
pubsub:
  kafka:
    hello:
      brokers: ["127.0.0.1:9092"]
      topic: topic_hello
      consumer:
        group_id: consumer_hello
        instances: 2
file:
  buckets:
    default:
      region_name: auto
app:
  user:
    default_avatar_urls: ["a.png", "b.png"]
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	c := lynx.NewViperConfig(v)

	var cfg AppConfig
	if err := UnmarshalConfig(c, &cfg); err != nil {
		t.Fatal(err)
	}

	if got := cfg.GetServer().GetHttp().GetAddr(); got != ":9099" {
		t.Errorf("server.http.addr = %q, want %q", got, ":9099")
	}
	if got := cfg.GetServer().GetHttp().GetCors().GetAllowOrigins(); len(got) != 1 || got[0] != "*" {
		t.Errorf("cors allow_origins = %v", got)
	}
	if got := cfg.GetServer().GetGrpc().GetAddr(); got != ":8088" {
		t.Errorf("server.grpc.addr = %q", got)
	}
	if got := cfg.GetSecurity().GetJwt().GetSecret(); got != "test-secret" {
		t.Errorf("security.jwt.secret = %q", got)
	}
	if got := cfg.GetSecurity().GetJwt().GetTokenExpirySec(); got != 3600 {
		t.Errorf("security.jwt.token_expiry_sec = %d", got)
	}
	if got := cfg.GetData().GetDatabase().GetSource(); got != "postgres://user:pass@localhost:5432/db" {
		t.Errorf("data.database.source = %q", got)
	}
	kafka := cfg.GetPubsub().GetKafka()
	if got := kafka["hello"].GetTopic(); got != "topic_hello" {
		t.Errorf("kafka hello topic = %q", got)
	}
	if got := kafka["hello"].GetConsumer().GetInstances(); got != 2 {
		t.Errorf("kafka consumer instances = %d", got)
	}
	if got := cfg.GetFile().GetBuckets()["default"].GetRegionName(); got != "auto" {
		t.Errorf("file bucket region = %q", got)
	}
	if got := cfg.GetApp().GetUser().GetDefaultAvatarUrls(); len(got) != 2 {
		t.Errorf("default_avatar_urls = %v", got)
	}
}

func TestConfigureViperEnvBinding(t *testing.T) {
	v := viper.New()
	c := lynx.NewViperConfig(v)
	if err := ConfigureViper(pflag.NewFlagSet("test", pflag.ContinueOnError), c); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LYNX_DATA_DATABASE_SOURCE", "postgres://env:pass@host/db")
	if got := c.GetString("data.database.source"); got != "postgres://env:pass@host/db" {
		t.Errorf("env override = %q", got)
	}
}

func TestEnvName(t *testing.T) {
	cases := map[string]string{
		"data.database.source":               "LYNX_DATA_DATABASE_SOURCE",
		"file.buckets.default.access_key_id": "LYNX_FILE_BUCKETS_DEFAULT_ACCESS_KEY_ID",
		"security.jwt.refresh_token_secret":  "LYNX_SECURITY_JWT_REFRESH_TOKEN_SECRET",
	}
	for k, want := range cases {
		if got := envName(k); got != want {
			t.Errorf("envName(%q) = %q, want %q", k, got, want)
		}
	}
}

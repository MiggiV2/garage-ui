package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "garage.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func envFrom(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestLoadPrefersExplicitAdminToken(t *testing.T) {
	cfg, err := Load(envFrom(map[string]string{
		"GARAGE_ADMIN_TOKEN": "explicit",
		"GARAGE_CONFIG_FILE": writeFile(t, "[admin]\nadmin_token = \"from-file\"\n"),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AdminToken != "explicit" {
		t.Fatalf("expected explicit token, got %q", cfg.AdminToken)
	}
}

func TestLoadReadsAdminTokenFromGarageConfig(t *testing.T) {
	path := writeFile(t, `metadata_dir = "/var/lib/garage/meta"

[s3_api]
api_bind_addr = "[::]:3900"
s3_region = "garage"

[admin]
api_bind_addr = "[::]:3903"
admin_token = "from-file"
metrics_token = "metrics"
`)
	cfg, err := Load(envFrom(map[string]string{"GARAGE_CONFIG_FILE": path}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AdminToken != "from-file" {
		t.Fatalf("expected token from file, got %q", cfg.AdminToken)
	}
}

func TestLoadReadsS3RegionFromGarageConfigAndEndpointFromEnv(t *testing.T) {
	path := writeFile(t, "[s3_api]\ns3_region = \"home\"\n\n[admin]\nadmin_token = \"t\"\n")
	cfg, err := Load(envFrom(map[string]string{"GARAGE_CONFIG_FILE": path, "S3_ENDPOINT": "https://s3.example.test"}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.S3Region != "home" {
		t.Fatalf("expected region from file, got %q", cfg.S3Region)
	}
	if cfg.S3Endpoint != "https://s3.example.test" {
		t.Fatalf("expected endpoint from env, got %q", cfg.S3Endpoint)
	}
}

func TestLoadS3RegionEnvOverridesFileAndDefaultsToGarage(t *testing.T) {
	path := writeFile(t, "[s3_api]\ns3_region = \"home\"\n\n[admin]\nadmin_token = \"t\"\n")
	cfg, err := Load(envFrom(map[string]string{"GARAGE_CONFIG_FILE": path, "S3_REGION": "eu"}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.S3Region != "eu" {
		t.Fatalf("expected env region, got %q", cfg.S3Region)
	}

	cfg, err = Load(envFrom(map[string]string{"GARAGE_ADMIN_TOKEN": "t"}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.S3Region != "garage" {
		t.Fatalf("expected default region garage, got %q", cfg.S3Region)
	}
}

func TestLoadFailsWithoutToken(t *testing.T) {
	if _, err := Load(envFrom(map[string]string{})); err == nil {
		t.Fatal("expected error when no admin token is configured")
	}
}

func TestLoadFailsWhenConfigFileHasNoToken(t *testing.T) {
	path := writeFile(t, "[admin]\napi_bind_addr = \"[::]:3903\"\n")
	if _, err := Load(envFrom(map[string]string{"GARAGE_CONFIG_FILE": path})); err == nil {
		t.Fatal("expected error when config file has no admin_token")
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(envFrom(map[string]string{"GARAGE_ADMIN_TOKEN": "t"}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AdminURL != "http://garage:3903" {
		t.Fatalf("unexpected default admin url %q", cfg.AdminURL)
	}
	if cfg.ListenAddr != ":8080" {
		t.Fatalf("unexpected default listen addr %q", cfg.ListenAddr)
	}
	if cfg.BasicAuthEnabled() {
		t.Fatal("basic auth must be disabled by default")
	}
}

func TestLoadBasicAuthRequiresBothValues(t *testing.T) {
	_, err := Load(envFrom(map[string]string{"GARAGE_ADMIN_TOKEN": "t", "UI_USERNAME": "admin"}))
	if err == nil {
		t.Fatal("expected error when only UI_USERNAME is set")
	}

	cfg, err := Load(envFrom(map[string]string{"GARAGE_ADMIN_TOKEN": "t", "UI_USERNAME": "admin", "UI_PASSWORD": "pw"}))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.BasicAuthEnabled() {
		t.Fatal("expected basic auth to be enabled")
	}
}

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"recipe-image/internal/config"
)

func TestLoadRequiresWSS(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	content := `
server:
  ws_url: "ws://insecure.example/ws"
  token: "secret"
oss:
  endpoint: "oss-cn-test.aliyuncs.com"
  bucket: "b"
`
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.Load(p); err == nil {
		t.Fatal("expected wss validation error")
	}
}

func TestLoadOK(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	content := `
server:
  ws_url: "wss://example.com/api/ws/image-worker"
  token: "secret"
oss:
  endpoint: "oss-cn-test.aliyuncs.com"
  access_key_id: "a"
  access_key_secret: "b"
  bucket: "c"
`
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.Load(p); err != nil {
		t.Fatal(err)
	}
}

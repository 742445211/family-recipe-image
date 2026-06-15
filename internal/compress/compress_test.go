package compress_test

import (
	"os"
	"path/filepath"
	"testing"

	"recipe-image/internal/compress"
	"recipe-image/internal/config"
)

func TestDetectFormat(t *testing.T) {
	if got := compress.DetectFormat(".jpg", ""); got != "jpeg" {
		t.Fatalf("jpg: %s", got)
	}
	if got := compress.DetectFormat(".gif", ""); got != "gif" {
		t.Fatalf("gif: %s", got)
	}
}

func TestReplaceKeyExtension(t *testing.T) {
	got := compress.ReplaceKeyExtension("recipe/123.webp", ".png")
	if got != "recipe/123.png" {
		t.Fatalf("got %s", got)
	}
}

func TestProcessGIFSkipped(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.gif")
	if err := os.WriteFile(p, []byte("GIF89a"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.CompressConfig{}
	res, err := compress.Process(cfg, p, "recipe/a.gif", "gif")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Skipped || res.Reason != "gif_not_supported" {
		t.Fatalf("expected gif skip: %+v", res)
	}
}

package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	OSS        OSSConfig        `yaml:"oss"`
	Worker     WorkerConfig     `yaml:"worker"`
	Compress   CompressConfig   `yaml:"compress"`
	Firdgemate FirdgemateConfig `yaml:"firdgemate"`
}

type ServerConfig struct {
	WsURL           string `yaml:"ws_url"`
	Token           string `yaml:"token"`
	WorkerID        string `yaml:"worker_id"`
	ReconnectSec    int    `yaml:"reconnect_sec"`
	PingIntervalSec int    `yaml:"ping_interval_sec"`
}

type OSSConfig struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	Bucket          string `yaml:"bucket"`
	CustomDomain    string `yaml:"custom_domain"`
}

type WorkerConfig struct {
	MaxConcurrent int    `yaml:"max_concurrent"`
	TempDir       string `yaml:"temp_dir"`
}

type CompressConfig struct {
	OxipngPath    string   `yaml:"oxipng_path"`
	OxipngFlags   []string `yaml:"oxipng_flags"`
	OxipngThreads int      `yaml:"oxipng_threads"`
	CjpegPath     string   `yaml:"cjpeg_path"`
	DjpegPath     string   `yaml:"djpeg_path"`
	CjpegQuality  int      `yaml:"cjpeg_quality"`
	MagickPath    string   `yaml:"magick_path"`
	FfmpegPath    string   `yaml:"ffmpeg_path"`
}

type FirdgemateConfig struct {
	ModelPath       string  `yaml:"model_path"`
	OnnxLibPath     string  `yaml:"onnx_lib_path"`
	NumClasses      int     `yaml:"num_classes"`
	ConfThreshold   float32 `yaml:"conf_threshold"`
	IouThreshold    float32 `yaml:"iou_threshold"`
	InputSize       int     `yaml:"input_size"`
	IntraOpThreads  int     `yaml:"intra_op_threads"`
}

var AppConfig *Config

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	applyDefaults(cfg)
	if err := cfg.Validate(); err != nil {
		return err
	}
	AppConfig = cfg
	return nil
}

func applyDefaults(c *Config) {
	if c.Server.ReconnectSec == 0 {
		c.Server.ReconnectSec = 5
	}
	if c.Server.PingIntervalSec == 0 {
		c.Server.PingIntervalSec = 30
	}
	if c.Server.WorkerID == "" {
		c.Server.WorkerID = "pi-gateway"
	}
	if c.Worker.MaxConcurrent == 0 {
		c.Worker.MaxConcurrent = 2
	}
	if c.Worker.TempDir == "" {
		c.Worker.TempDir = "/tmp/recipe-image"
	}
	if c.Compress.OxipngPath == "" {
		c.Compress.OxipngPath = "/usr/local/bin/oxipng"
	}
	if len(c.Compress.OxipngFlags) == 0 {
		c.Compress.OxipngFlags = []string{"-o", "4", "--strip", "safe"}
	}
	if c.Compress.OxipngThreads == 0 {
		c.Compress.OxipngThreads = 4
	}
	if c.Compress.CjpegPath == "" {
		c.Compress.CjpegPath = "/usr/local/bin/cjpeg"
	}
	if c.Compress.DjpegPath == "" {
		c.Compress.DjpegPath = "/usr/local/bin/djpeg"
	}
	if c.Compress.CjpegQuality == 0 {
		c.Compress.CjpegQuality = 85
	}
	if c.Compress.MagickPath == "" {
		c.Compress.MagickPath = "/usr/bin/magick"
	}
	if c.Compress.FfmpegPath == "" {
		c.Compress.FfmpegPath = "/usr/bin/ffmpeg"
	}
	if c.Firdgemate.ConfThreshold == 0 {
		c.Firdgemate.ConfThreshold = 0.45
	}
	if c.Firdgemate.IouThreshold == 0 {
		c.Firdgemate.IouThreshold = 0.5
	}
	if c.Firdgemate.InputSize == 0 {
		c.Firdgemate.InputSize = 640
	}
	if c.Firdgemate.IntraOpThreads == 0 {
		c.Firdgemate.IntraOpThreads = 4
	}
	if c.Firdgemate.NumClasses == 0 {
		c.Firdgemate.NumClasses = 47
	}
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	if !strings.HasPrefix(c.Server.WsURL, "wss://") {
		return fmt.Errorf("server.ws_url must use wss:// for TLS")
	}
	if strings.TrimSpace(c.Server.Token) == "" {
		return fmt.Errorf("server.token is required")
	}
	if strings.TrimSpace(c.OSS.Endpoint) == "" || strings.TrimSpace(c.OSS.Bucket) == "" {
		return fmt.Errorf("oss endpoint and bucket are required")
	}
	return nil
}

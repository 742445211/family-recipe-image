package ossclient

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"recipe-image/internal/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Client struct {
	cfg    config.OSSConfig
	bucket *oss.Bucket
}

func NewClient(cfg config.OSSConfig) (*Client, error) {
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("oss connect: %w", err)
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("oss bucket: %w", err)
	}
	return &Client{cfg: cfg, bucket: bucket}, nil
}

func (c *Client) Download(key, destPath string) (int64, string, error) {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return 0, "", err
	}
	body, err := c.bucket.GetObject(key)
	if err != nil {
		return 0, "", fmt.Errorf("get object: %w", err)
	}
	defer body.Close()

	f, err := os.Create(destPath)
	if err != nil {
		return 0, "", err
	}
	defer f.Close()

	n, err := io.Copy(f, body)
	if err != nil {
		return 0, "", err
	}

	meta, err := c.bucket.GetObjectDetailedMeta(key)
	contentType := ""
	if err == nil {
		contentType = meta.Get("Content-Type")
	}
	return n, contentType, nil
}

func (c *Client) UploadOverwrite(key, filePath, contentType string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	opts := []oss.Option{}
	if contentType != "" {
		opts = append(opts, oss.ContentType(contentType))
	}
	return c.bucket.PutObjectFromFile(key, filePath, opts...)
}

func (c *Client) DeleteObject(key string) error {
	return c.bucket.DeleteObject(key)
}

func (c *Client) BuildURL(key string) string {
	if c.cfg.CustomDomain != "" {
		return fmt.Sprintf("%s/%s", c.cfg.CustomDomain, key)
	}
	return fmt.Sprintf("https://%s.%s/%s", c.cfg.Bucket, c.cfg.Endpoint, key)
}

func (c *Client) ObjectExists(key string) (bool, error) {
	return c.bucket.IsObjectExist(key)
}

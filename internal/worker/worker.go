package worker

import (
	"fmt"
	"os"
	"path/filepath"

	"recipe-image/internal/compress"
	"recipe-image/internal/config"
	"recipe-image/internal/firdgemate"
	ossclient "recipe-image/internal/oss"
	"recipe-image/internal/protocol"
)

type CompressWorker struct {
	cfg    *config.Config
	oss    *ossclient.Client
	tempDir string
}

func NewCompressWorker(cfg *config.Config, oss *ossclient.Client) *CompressWorker {
	return &CompressWorker{cfg: cfg, oss: oss, tempDir: cfg.Worker.TempDir}
}

func (w *CompressWorker) Run(task *protocol.TaskMessage) (*protocol.CompressDetail, error) {
	if err := os.MkdirAll(w.tempDir, 0o755); err != nil {
		return nil, err
	}
	ext := filepath.Ext(task.OssKey)
	localPath := filepath.Join(w.tempDir, task.TaskID+ext)
	origSize, contentType, err := w.oss.Download(task.OssKey, localPath)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer os.Remove(localPath)

	format := compress.DetectFormat(ext, contentType)
	result, err := compress.Process(w.cfg.Compress, localPath, task.OssKey, format)
	if err != nil {
		return nil, err
	}
	defer os.Remove(result.OutputPath)
	if result.OutputPath != localPath {
		// converted temp
	}

	detail := &protocol.CompressDetail{
		Skipped:         result.Skipped,
		Reason:          result.Reason,
		Format:          result.Format,
		OutputFormat:    result.OutputFormat,
		NewOssKey:       result.NewKey,
		OriginalBytes:   result.OriginalBytes,
		CompressedBytes: result.CompressedBytes,
	}
	if origSize > 0 && detail.OriginalBytes == 0 {
		detail.OriginalBytes = origSize
	}

	if result.Skipped {
		return detail, nil
	}

	uploadKey := result.NewKey
	uploadCT := contentTypeForFormat(result.OutputFormat)
	if err := w.oss.UploadOverwrite(uploadKey, result.OutputPath, uploadCT); err != nil {
		return nil, fmt.Errorf("upload: %w", err)
	}
	if uploadKey != task.OssKey {
		detail.NewOssKey = uploadKey
	}
	return detail, nil
}

func contentTypeForFormat(format string) string {
	switch format {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	default:
		return "image/png"
	}
}

type RecognizeWorker struct {
	cfg      *config.Config
	oss      *ossclient.Client
	detector *firdgemate.Detector
	tempDir  string
}

func NewRecognizeWorker(cfg *config.Config, oss *ossclient.Client, detector *firdgemate.Detector) *RecognizeWorker {
	return &RecognizeWorker{cfg: cfg, oss: oss, detector: detector, tempDir: cfg.Worker.TempDir}
}

func (w *RecognizeWorker) Run(task *protocol.TaskMessage) (*protocol.RecognizeDetail, error) {
	if w.detector == nil {
		return nil, fmt.Errorf("detector not initialized")
	}
	if err := os.MkdirAll(w.tempDir, 0o755); err != nil {
		return nil, err
	}
	ext := filepath.Ext(task.OssKey)
	localPath := filepath.Join(w.tempDir, task.TaskID+ext)
	_, _, err := w.oss.Download(task.OssKey, localPath)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer os.Remove(localPath)

	ingredients, err := w.detector.Detect(localPath)
	if err != nil {
		return nil, err
	}
	return protocol.NewRecognizeDetail(ingredients), nil
}

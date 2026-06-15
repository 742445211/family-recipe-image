package compress

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"recipe-image/internal/config"
)

type Result struct {
	OutputPath      string
	OutputFormat    string
	NewKey          string
	OriginalBytes   int64
	CompressedBytes int64
	Skipped         bool
	Reason          string
	Format          string
}

func DetectFormat(ext, contentType string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg":
		return "jpeg"
	case ".png":
		return "png"
	case ".gif":
		return "gif"
	case ".webp":
		return "webp"
	case ".bmp":
		return "bmp"
	case ".tif", ".tiff":
		return "tiff"
	case ".heic", ".heif":
		return "heic"
	}
	ct := strings.ToLower(contentType)
	switch {
	case strings.Contains(ct, "jpeg"):
		return "jpeg"
	case strings.Contains(ct, "png"):
		return "png"
	case strings.Contains(ct, "gif"):
		return "gif"
	case strings.Contains(ct, "webp"):
		return "webp"
	case strings.Contains(ct, "bmp"):
		return "bmp"
	case strings.Contains(ct, "tiff"):
		return "tiff"
	}
	return "unknown"
}

func ReplaceKeyExtension(key, newExt string) string {
	base := strings.TrimSuffix(key, filepath.Ext(key))
	if !strings.HasPrefix(newExt, ".") {
		newExt = "." + newExt
	}
	return base + newExt
}

func Process(cfg config.CompressConfig, inputPath, ossKey, format string) (*Result, error) {
	info, err := os.Stat(inputPath)
	if err != nil {
		return nil, err
	}
	origSize := info.Size()

	switch format {
	case "gif":
		return &Result{
			OutputPath:    inputPath,
			Format:        format,
			OriginalBytes: origSize,
			Skipped:       true,
			Reason:        "gif_not_supported",
		}, nil
	case "jpeg":
		out, err := optimizeJPEG(cfg, inputPath)
		if err != nil {
			return nil, err
		}
		return finalizeSameKey(out, ossKey, format, origSize)
	case "png":
		out, err := optimizePNG(cfg, inputPath)
		if err != nil {
			return nil, err
		}
		return finalizeSameKey(out, ossKey, format, origSize)
	case "webp", "bmp", "tiff", "heic", "unknown":
		pngPath, err := convertToPNG(cfg, inputPath, format)
		if err != nil {
			return nil, err
		}
		out, err := optimizePNG(cfg, pngPath)
		if err != nil {
			return nil, err
		}
		newKey := ReplaceKeyExtension(ossKey, ".png")
		st, err := os.Stat(out)
		if err != nil {
			return nil, err
		}
		compSize := st.Size()
		if compSize >= origSize {
			return &Result{
				OutputPath:      inputPath,
				OutputFormat:    format,
				NewKey:          ossKey,
				OriginalBytes:   origSize,
				CompressedBytes: origSize,
				Skipped:         true,
				Reason:          "no_size_reduction",
				Format:          format,
			}, nil
		}
		return &Result{
			OutputPath:      out,
			OutputFormat:    "png",
			NewKey:          newKey,
			OriginalBytes:   origSize,
			CompressedBytes: compSize,
			Format:          format,
		}, nil
	default:
		return &Result{
			OutputPath:    inputPath,
			Format:        format,
			OriginalBytes: origSize,
			Skipped:       true,
			Reason:        "unsupported_format",
		}, nil
	}
}

func finalizeSameKey(outPath, ossKey, format string, origSize int64) (*Result, error) {
	st, err := os.Stat(outPath)
	if err != nil {
		return nil, err
	}
	compSize := st.Size()
	if compSize >= origSize {
		return &Result{
			OutputPath:      outPath,
			OutputFormat:    format,
			NewKey:          ossKey,
			OriginalBytes:   origSize,
			CompressedBytes: origSize,
			Skipped:         true,
			Reason:          "no_size_reduction",
			Format:          format,
		}, nil
	}
	return &Result{
		OutputPath:      outPath,
		OutputFormat:    format,
		NewKey:          ossKey,
		OriginalBytes:   origSize,
		CompressedBytes: compSize,
		Format:          format,
	}, nil
}

func optimizeJPEG(cfg config.CompressConfig, inputPath string) (string, error) {
	outPath := inputPath + ".opt.jpg"
	if _, err := os.Stat(cfg.CjpegPath); err == nil {
		args := []string{
			"-optimize", "-progressive",
			"-quality", fmt.Sprintf("%d", cfg.CjpegQuality),
			"-outfile", outPath,
		}
		djpeg := exec.Command(cfg.DjpegPath, inputPath)
		cjpeg := exec.Command(cfg.CjpegPath, args...)
		var decodeOut bytes.Buffer
		djpeg.Stdout = &decodeOut
		if err := djpeg.Run(); err == nil {
			cjpeg.Stdin = &decodeOut
			if err := cjpeg.Run(); err == nil {
				return outPath, nil
			}
		}
	}
	// fallback: ffmpeg re-encode when mozjpeg unavailable
	qscale := 4
	if cfg.CjpegQuality >= 90 {
		qscale = 2
	} else if cfg.CjpegQuality >= 80 {
		qscale = 4
	} else {
		qscale = 6
	}
	cmd := exec.Command(cfg.FfmpegPath, "-y", "-i", inputPath, "-q:v", fmt.Sprintf("%d", qscale), outPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("jpeg optimize: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return outPath, nil
}

func optimizePNG(cfg config.CompressConfig, inputPath string) (string, error) {
	args := append([]string{}, cfg.OxipngFlags...)
	args = append(args, "-t", fmt.Sprintf("%d", cfg.OxipngThreads), inputPath)
	cmd := exec.Command(cfg.OxipngPath, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("oxipng: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return inputPath, nil
}

func convertToPNG(cfg config.CompressConfig, inputPath, format string) (string, error) {
	outPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".converted.png"
	if format == "webp" {
		if err := exec.Command("dwebp", inputPath, "-o", outPath).Run(); err == nil {
			return outPath, nil
		}
	}
	cmd := exec.Command(cfg.MagickPath, inputPath, "-strip", outPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		cmd2 := exec.Command(cfg.FfmpegPath, "-y", "-i", inputPath, outPath)
		if out2, err2 := cmd2.CombinedOutput(); err2 != nil {
			return "", fmt.Errorf("convert to png: magick=%v ffmpeg=%v: %s / %s", err, err2, out, out2)
		}
	}
	return outPath, nil
}

//go:build ignore

// 从 OSS 下载并识别。用法 go run -tags cgo ./tools/recognize_oss_key.go <oss_key>
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"recipe-image/internal/config"
	"recipe-image/internal/firdgemate"
	ossclient "recipe-image/internal/oss"
	"recipe-image/internal/protocol"
	"recipe-image/internal/worker"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: recognize_oss_key <oss_key>")
		os.Exit(1)
	}
	key := os.Args[1]
	if err := config.Load("config.yaml"); err != nil {
		panic(err)
	}
	cfg := config.AppConfig
	oss, err := ossclient.NewClient(cfg.OSS)
	if err != nil {
		panic(err)
	}
	det, err := firdgemate.NewDetector(cfg.Firdgemate)
	if err != nil {
		panic(err)
	}
	defer det.Close()
	rw := worker.NewRecognizeWorker(cfg, oss, det)
	task := &protocol.TaskMessage{
		TaskID: "local-test",
		Action: protocol.ActionRecognize,
		OssKey: key,
	}
	detail, err := rw.Run(task)
	if err != nil {
		panic(err)
	}
	fmt.Printf("key=%s\ningredients=%v\nitems=%d\n", key, detail.Ingredients, len(detail.Items))
	_ = filepath.Base(key)
}

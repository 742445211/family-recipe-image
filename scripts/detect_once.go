//go:build ignore

// 测试：本地图片食材识别。用法 go run -tags cgo ./scripts/detect_once.go <image-path>
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"recipe-image/internal/config"
	"recipe-image/internal/firdgemate"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: detect_once <image-path>")
		os.Exit(1)
	}
	if err := config.Load("config.yaml"); err != nil {
		panic(err)
	}
	d, err := firdgemate.NewDetector(config.AppConfig.Firdgemate)
	if err != nil {
		panic(err)
	}
	defer d.Close()
	ingredients, err := d.Detect(os.Args[1])
	if err != nil {
		panic(err)
	}
	out, _ := json.MarshalIndent(ingredients, "", "  ")
	fmt.Println(string(out))
}

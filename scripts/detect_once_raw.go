//go:build ignore

// 测试：输出 YOLO 原始检测框。用法 go run -tags cgo ./scripts/detect_once_raw.go <image-path>
package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sort"

	"recipe-image/internal/config"
	"recipe-image/internal/firdgemate"

	"github.com/yalue/onnxruntime_go"
)

type rawDetection struct {
	Index      int     `json:"index"`
	ClassID    int     `json:"class_id"`
	LabelEN    string  `json:"label_en"`
	LabelZH    string  `json:"label_zh"`
	Confidence float32 `json:"confidence"`
	CX         float32 `json:"cx"`
	CY         float32 `json:"cy"`
	W          float32 `json:"w"`
	H          float32 `json:"h"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: detect_once_raw <image-path>")
		os.Exit(1)
	}
	if err := config.Load("config.yaml"); err != nil {
		panic(err)
	}
	cfg := config.AppConfig.Firdgemate

	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	b := img.Bounds()
	origW, origH := b.Dx(), b.Dy()

	if cfg.OnnxLibPath != "" {
		onnxruntime_go.SetSharedLibraryPath(cfg.OnnxLibPath)
	}
	if err := onnxruntime_go.InitializeEnvironment(); err != nil {
		panic(err)
	}
	session, err := onnxruntime_go.NewDynamicAdvancedSession(cfg.ModelPath, []string{"images"}, []string{"output0"}, nil)
	if err != nil {
		panic(err)
	}
	defer session.Destroy()

	input, _, _ := prepareInput(img, cfg.InputSize)
	numClasses := cfg.NumClasses
	if numClasses <= 0 {
		numClasses = 47
	}
	channels := 4 + numClasses
	numBoxes := 8400

	inTensor, err := onnxruntime_go.NewTensor(onnxruntime_go.NewShape(1, 3, int64(cfg.InputSize), int64(cfg.InputSize)), input)
	if err != nil {
		panic(err)
	}
	defer inTensor.Destroy()
	outTensor, err := onnxruntime_go.NewEmptyTensor[float32](onnxruntime_go.NewShape(1, int64(channels), int64(numBoxes)))
	if err != nil {
		panic(err)
	}
	defer outTensor.Destroy()
	if err := session.Run([]onnxruntime_go.Value{inTensor}, []onnxruntime_go.Value{outTensor}); err != nil {
		panic(err)
	}
	output := outTensor.GetData()
	labels := firdgemate.DefaultLabels(numClasses)

	var dets []rawDetection
	for i := 0; i < numBoxes; i++ {
		bestClass := -1
		bestScore := float32(0)
		for c := 0; c < numClasses; c++ {
			score := output[(4+c)*numBoxes+i]
			if score > bestScore {
				bestScore = score
				bestClass = c
			}
		}
		if bestScore < cfg.ConfThreshold {
			continue
		}
		en := firdgemate.ClassName(bestClass, labels)
		dets = append(dets, rawDetection{
			Index:      i,
			ClassID:    bestClass,
			LabelEN:    en,
			LabelZH:    firdgemate.ToChineseIngredient(en),
			Confidence: bestScore,
			CX:         output[0*numBoxes+i],
			CY:         output[1*numBoxes+i],
			W:          output[2*numBoxes+i],
			H:          output[3*numBoxes+i],
		})
	}
	sort.Slice(dets, func(i, j int) bool {
		return dets[i].Confidence > dets[j].Confidence
	})

	result := map[string]any{
		"image": map[string]any{
			"path":        os.Args[1],
			"width":       origW,
			"height":      origH,
			"input_size":  cfg.InputSize,
			"conf_thresh": cfg.ConfThreshold,
			"iou_thresh":  cfg.IouThreshold,
		},
		"total_above_threshold": len(dets),
		"detections":            dets,
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
}

func prepareInput(img image.Image, size int) ([]float32, int, int) {
	b := img.Bounds()
	origW, origH := b.Dx(), b.Dy()
	resized := resizeNearest(img, size, size)
	data := make([]float32, 3*size*size)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			r, g, bl, _ := resized.At(x, y).RGBA()
			i := y*size + x
			data[i] = float32(r>>8) / 255.0
			data[size*size+i] = float32(g>>8) / 255.0
			data[2*size*size+i] = float32(bl>>8) / 255.0
		}
	}
	return data, origW, origH
}

func resizeNearest(img image.Image, w, h int) image.Image {
	b := img.Bounds()
	srcW, srcH := b.Dx(), b.Dy()
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sx := b.Min.X + x*srcW/w
			sy := b.Min.Y + y*srcH/h
			out.Set(x, y, img.At(sx, sy))
		}
	}
	return out
}

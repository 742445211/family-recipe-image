//go:build ignore

// 测试：对比各类别原始得分。用法 go run -tags cgo ./scripts/detect_class_scores.go <image-path>
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

type classScore struct {
	ClassID int     `json:"class_id"`
	LabelEN string  `json:"label_en"`
	LabelZH string  `json:"label_zh"`
	Score   float32 `json:"score"`
}

type anchorTopK struct {
	Index int          `json:"index"`
	CX    float32      `json:"cx"`
	CY    float32      `json:"cy"`
	Top   []classScore `json:"top_classes"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: detect_class_scores <image-path>")
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

	type scoredAnchor struct {
		index int
		best  float32
	}
	var anchors []scoredAnchor
	for i := 0; i < numBoxes; i++ {
		best := float32(0)
		for c := 0; c < numClasses; c++ {
			if s := output[(4+c)*numBoxes+i]; s > best {
				best = s
			}
		}
		if best >= 0.3 {
			anchors = append(anchors, scoredAnchor{index: i, best: best})
		}
	}
	sort.Slice(anchors, func(i, j int) bool { return anchors[i].best > anchors[j].best })

	var topAnchors []anchorTopK
	for _, a := range anchors[:min(5, len(anchors))] {
		scores := make([]classScore, numClasses)
		for c := 0; c < numClasses; c++ {
			en := firdgemate.ClassName(c, labels)
			scores[c] = classScore{
				ClassID: c,
				LabelEN: en,
				LabelZH: firdgemate.ToChineseIngredient(en),
				Score:   output[(4+c)*numBoxes+a.index],
			}
		}
		sort.Slice(scores, func(i, j int) bool { return scores[i].Score > scores[j].Score })
		topAnchors = append(topAnchors, anchorTopK{
			Index: a.index,
			CX:    output[0*numBoxes+a.index],
			CY:    output[1*numBoxes+a.index],
			Top:   scores[:5],
		})
	}

	eggIdx := 19
	var eggScores []float32
	for i := 0; i < numBoxes; i++ {
		s := output[(4+eggIdx)*numBoxes+i]
		if s >= 0.1 {
			eggScores = append(eggScores, s)
		}
	}
	sort.Slice(eggScores, func(i, j int) bool { return eggScores[i] > eggScores[j] })

	result := map[string]any{
		"top_anchors_top5_classes": topAnchors,
		"egg_class_best_scores":    eggScores[:min(10, len(eggScores))],
		"egg_class_max":            maxFloat(eggScores),
		"orange_class_max":         bestClassMax(output, numBoxes, numClasses, 33),
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
}

func bestClassMax(output []float32, numBoxes, numClasses, classID int) float32 {
	max := float32(0)
	for i := 0; i < numBoxes; i++ {
		if s := output[(4+classID)*numBoxes+i]; s > max {
			max = s
		}
	}
	return max
}

func maxFloat(v []float32) float32 {
	if len(v) == 0 {
		return 0
	}
	m := v[0]
	for _, x := range v[1:] {
		if x > m {
			m = x
		}
	}
	return m
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

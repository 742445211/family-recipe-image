//go:build cgo

package firdgemate

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sort"
	"sync"

	"recipe-image/internal/config"

	"github.com/yalue/onnxruntime_go"
)

type Detector struct {
	cfg     config.FirdgemateConfig
	session *onnxruntime_go.DynamicAdvancedSession
	mu      sync.Mutex
}

var ortOnce sync.Once
var ortInitErr error

func NewDetector(cfg config.FirdgemateConfig) (*Detector, error) {
	if _, err := os.Stat(cfg.ModelPath); err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}
	ortOnce.Do(func() {
		if cfg.OnnxLibPath != "" {
			onnxruntime_go.SetSharedLibraryPath(cfg.OnnxLibPath)
		}
		ortInitErr = onnxruntime_go.InitializeEnvironment()
	})
	if ortInitErr != nil {
		return nil, fmt.Errorf("onnx init: %w", ortInitErr)
	}

	inputNames := []string{"images"}
	outputNames := []string{"output0"}
	session, err := onnxruntime_go.NewDynamicAdvancedSession(cfg.ModelPath, inputNames, outputNames, nil)
	if err != nil {
		return nil, fmt.Errorf("onnx session: %w", err)
	}
	return &Detector{cfg: cfg, session: session}, nil
}

func (d *Detector) Close() {
	if d.session != nil {
		d.session.Destroy()
	}
}

func (d *Detector) Detect(imagePath string) ([]string, error) {
	f, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	input, _, _ := prepareInput(img, d.cfg.InputSize)
	numClasses := d.cfg.NumClasses
	if numClasses <= 0 {
		numClasses = 47
	}
	channels := 4 + numClasses
	outputSize := channels * 8400
	output := make([]float32, outputSize)
	inTensor, err := onnxruntime_go.NewTensor(onnxruntime_go.NewShape(1, 3, int64(d.cfg.InputSize), int64(d.cfg.InputSize)), input)
	if err != nil {
		return nil, err
	}
	defer inTensor.Destroy()
	outTensor, err := onnxruntime_go.NewEmptyTensor[float32](onnxruntime_go.NewShape(1, int64(channels), 8400))
	if err != nil {
		return nil, err
	}
	defer outTensor.Destroy()

	d.mu.Lock()
	err = d.session.Run([]onnxruntime_go.Value{inTensor}, []onnxruntime_go.Value{outTensor})
	d.mu.Unlock()
	if err != nil {
		return nil, err
	}
	copy(output, outTensor.GetData())

	labels := DefaultLabels(numClasses)
	boxes := decodeYOLOv8(output, numClasses, d.cfg.ConfThreshold)
	seen := map[string]bool{}
	var ingredients []string
	for _, b := range boxes {
		name := ToChineseIngredient(ClassName(b.ClassID, labels))
		if !seen[name] {
			seen[name] = true
			ingredients = append(ingredients, name)
		}
	}
	sort.Strings(ingredients)
	return ingredients, nil
}

type box struct {
	ClassID    int
	Confidence float32
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

func decodeYOLOv8(output []float32, numClasses int, confThresh float32) []box {
	channels := 4 + numClasses
	numBoxes := 8400
	if len(output) < channels*numBoxes {
		numBoxes = len(output) / channels
	}
	var candidates []box
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
		if bestScore >= confThresh {
			candidates = append(candidates, box{ClassID: bestClass, Confidence: bestScore})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Confidence > candidates[j].Confidence
	})
	if len(candidates) > 20 {
		candidates = candidates[:20]
	}
	return candidates
}

package firdgemate

import (
	"image"
	"math"
	"sort"
)

const letterboxPadValue = 114.0 / 255.0

type detectionBox struct {
	ClassID    int
	Confidence float32
	CX, CY     float32
	W, H       float32
}

func prepareLetterboxInput(img image.Image, size int) []float32 {
	b := img.Bounds()
	origW, origH := b.Dx(), b.Dy()
	data := make([]float32, 3*size*size)
	for i := range data {
		data[i] = letterboxPadValue
	}
	if origW <= 0 || origH <= 0 {
		return data
	}

	scale := math.Min(float64(size)/float64(origW), float64(size)/float64(origH))
	newW := int(math.Round(float64(origW) * scale))
	newH := int(math.Round(float64(origH) * scale))
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}
	padX := (size - newW) / 2
	padY := (size - newH) / 2

	resized := resizeNearest(img, newW, newH)
	rb := resized.Bounds()
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			r, g, bl, _ := resized.At(rb.Min.X+x, rb.Min.Y+y).RGBA()
			dstX := padX + x
			dstY := padY + y
			i := dstY*size + dstX
			data[i] = float32(r>>8) / 255.0
			data[size*size+i] = float32(g>>8) / 255.0
			data[2*size*size+i] = float32(bl>>8) / 255.0
		}
	}
	return data
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

func decodeYOLOv8(output []float32, numClasses int, confThresh, iouThresh float32, maxDet int) []detectionBox {
	channels := 4 + numClasses
	numBoxes := 8400
	if len(output) < channels*numBoxes {
		numBoxes = len(output) / channels
	}
	if maxDet <= 0 {
		maxDet = 20
	}

	var candidates []detectionBox
	for i := 0; i < numBoxes; i++ {
		cx := output[0*numBoxes+i]
		cy := output[1*numBoxes+i]
		w := output[2*numBoxes+i]
		h := output[3*numBoxes+i]
		if w <= 0 || h <= 0 {
			continue
		}

		bestClass := -1
		bestScore := float32(0)
		for c := 0; c < numClasses; c++ {
			score := output[(4+c)*numBoxes+i]
			if score > bestScore {
				bestScore = score
				bestClass = c
			}
		}
		if bestClass < 0 || bestScore < confThresh {
			continue
		}
		candidates = append(candidates, detectionBox{
			ClassID:    bestClass,
			Confidence: bestScore,
			CX:         cx,
			CY:         cy,
			W:          w,
			H:          h,
		})
	}

	kept := nms(candidates, iouThresh)
	sort.Slice(kept, func(i, j int) bool {
		return kept[i].Confidence > kept[j].Confidence
	})
	if len(kept) > maxDet {
		kept = kept[:maxDet]
	}
	return kept
}

func nms(boxes []detectionBox, iouThresh float32) []detectionBox {
	if len(boxes) == 0 {
		return nil
	}
	sorted := append([]detectionBox(nil), boxes...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Confidence > sorted[j].Confidence
	})

	kept := make([]detectionBox, 0, len(sorted))
	suppressed := make([]bool, len(sorted))
	for i := 0; i < len(sorted); i++ {
		if suppressed[i] {
			continue
		}
		kept = append(kept, sorted[i])
		for j := i + 1; j < len(sorted); j++ {
			if suppressed[j] {
				continue
			}
			if sorted[i].ClassID != sorted[j].ClassID {
				continue
			}
			if boxIoU(sorted[i], sorted[j]) > iouThresh {
				suppressed[j] = true
			}
		}
	}
	return kept
}

func boxIoU(a, b detectionBox) float32 {
	ax1 := a.CX - a.W/2
	ay1 := a.CY - a.H/2
	ax2 := a.CX + a.W/2
	ay2 := a.CY + a.H/2
	bx1 := b.CX - b.W/2
	by1 := b.CY - b.H/2
	bx2 := b.CX + b.W/2
	by2 := b.CY + b.H/2

	interX1 := max(ax1, bx1)
	interY1 := max(ay1, by1)
	interX2 := min(ax2, bx2)
	interY2 := min(ay2, by2)
	interW := interX2 - interX1
	interH := interY2 - interY1
	if interW <= 0 || interH <= 0 {
		return 0
	}
	interArea := interW * interH
	areaA := a.W * a.H
	areaB := b.W * b.H
	union := areaA + areaB - interArea
	if union <= 0 {
		return 0
	}
	return interArea / union
}

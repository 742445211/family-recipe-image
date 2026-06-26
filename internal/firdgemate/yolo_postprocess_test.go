package firdgemate

import (
	"image"
	"image/color"
	"testing"
)

func TestPrepareLetterboxInputPreservesAspectRatio(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}

	data := prepareLetterboxInput(img, 640)
	if len(data) != 3*640*640 {
		t.Fatalf("unexpected tensor size: %d", len(data))
	}

	// Top/bottom padding rows should remain gray.
	for x := 0; x < 640; x++ {
		top := data[x]
		bottom := data[639*640+x]
		if top != letterboxPadValue || bottom != letterboxPadValue {
			t.Fatalf("expected gray padding, got top=%v bottom=%v", top, bottom)
		}
	}
}

func TestNMSKeepsBestPerClass(t *testing.T) {
	boxes := []detectionBox{
		{ClassID: 19, Confidence: 0.9, CX: 100, CY: 100, W: 40, H: 40},
		{ClassID: 19, Confidence: 0.8, CX: 105, CY: 105, W: 40, H: 40},
		{ClassID: 33, Confidence: 0.85, CX: 300, CY: 300, W: 40, H: 40},
	}
	kept := nms(boxes, 0.5)
	if len(kept) != 2 {
		t.Fatalf("expected 2 boxes after NMS, got %d", len(kept))
	}
}

func TestDecodeYOLOv8FindsHighConfidenceBox(t *testing.T) {
	numClasses := 3
	numBoxes := 8400
	channels := 4 + numClasses
	output := make([]float32, channels*numBoxes)
	idx := 100
	output[0*numBoxes+idx] = 320
	output[1*numBoxes+idx] = 320
	output[2*numBoxes+idx] = 80
	output[3*numBoxes+idx] = 80
	output[(4+1)*numBoxes+idx] = 0.92

	boxes := decodeYOLOv8(output, numClasses, 0.25, 0.5, 10)
	if len(boxes) != 1 {
		t.Fatalf("expected 1 detection, got %d", len(boxes))
	}
	if boxes[0].ClassID != 1 {
		t.Fatalf("expected class 1, got %d", boxes[0].ClassID)
	}
}

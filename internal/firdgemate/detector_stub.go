//go:build !cgo

package firdgemate

import (
	"fmt"

	"recipe-image/internal/config"
)

type Detector struct{}

func NewDetector(cfg config.FirdgemateConfig) (*Detector, error) {
	return nil, fmt.Errorf("firdgemate requires CGO and ONNX runtime (build with CGO_ENABLED=1 on Pi)")
}

func (d *Detector) Close() {}

func (d *Detector) Detect(imagePath string) ([]string, error) {
	return nil, fmt.Errorf("firdgemate not available without CGO")
}

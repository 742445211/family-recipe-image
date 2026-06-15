package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"recipe-image/internal/config"
	"recipe-image/internal/dispatcher"
	"recipe-image/internal/firdgemate"
	ossclient "recipe-image/internal/oss"
	"recipe-image/internal/protocol"
	"recipe-image/internal/worker"
	"recipe-image/internal/ws"
)

func main() {
	cfgPath := "config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}
	if err := config.Load(cfgPath); err != nil {
		log.Fatalf("config: %v", err)
	}
	cfg := config.AppConfig

	oss, err := ossclient.NewClient(cfg.OSS)
	if err != nil {
		log.Fatalf("oss: %v", err)
	}

	var detector *firdgemate.Detector
	if d, err := firdgemate.NewDetector(cfg.Firdgemate); err != nil {
		log.Printf("warning: firdgemate disabled: %v", err)
	} else {
		detector = d
		defer detector.Close()
	}

	compressW := worker.NewCompressWorker(cfg, oss)
	recognizeW := worker.NewRecognizeWorker(cfg, oss, detector)

	sender := &resultSender{}
	disp := dispatcher.New(compressW, recognizeW, cfg.Worker.MaxConcurrent, sender.send)
	wsClient := ws.NewClient(cfg.Server, disp)
	sender.client = wsClient

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Printf("recipe-gateway starting worker_id=%s", cfg.Server.WorkerID)
	if err := wsClient.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("gateway: %v", err)
	}
	disp.Wait()
	log.Println("recipe-gateway stopped")
}

type resultSender struct {
	client *ws.Client
}

func (r *resultSender) send(msg *protocol.TaskResultMessage) {
	if r.client != nil {
		r.client.SendResult(msg)
	}
}

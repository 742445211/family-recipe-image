package dispatcher

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"recipe-image/internal/protocol"
	"recipe-image/internal/worker"
)

type ResultSender func(*protocol.TaskResultMessage)

type Dispatcher struct {
	compress  *worker.CompressWorker
	recognize *worker.RecognizeWorker
	sem       chan struct{}
	send      ResultSender
	wg        sync.WaitGroup
}

func New(compress *worker.CompressWorker, recognize *worker.RecognizeWorker, maxConcurrent int, send ResultSender) *Dispatcher {
	if maxConcurrent <= 0 {
		maxConcurrent = 2
	}
	return &Dispatcher{
		compress:  compress,
		recognize: recognize,
		sem:       make(chan struct{}, maxConcurrent),
		send:      send,
	}
}

func (d *Dispatcher) Submit(ctx context.Context, task *protocol.TaskMessage) {
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		select {
		case d.sem <- struct{}{}:
			defer func() { <-d.sem }()
		case <-ctx.Done():
			d.sendError(task, "dispatcher shutdown")
			return
		}
		d.handle(task)
	}()
}

func (d *Dispatcher) Wait() {
	d.wg.Wait()
}

func (d *Dispatcher) handle(task *protocol.TaskMessage) {
	log.Printf("[worker] start task=%s action=%s key=%s scope=%s scan_id=%d",
		task.TaskID, task.Action, task.OssKey, task.Parsed.Scope, task.Parsed.ScanID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	done := make(chan struct{})
	var (
		result *protocol.TaskResultMessage
	)
	go func() {
		defer close(done)
		switch task.Action {
		case protocol.ActionCompress:
			detail, err := d.compress.Run(task)
			if err != nil {
				result = d.errorResult(task, err.Error())
				return
			}
			raw, _ := protocol.MarshalDetail(detail)
			result = &protocol.TaskResultMessage{
				Type:   protocol.TypeTaskResult,
				TaskID: task.TaskID,
				Status: protocol.StatusOK,
				Action: task.Action,
				OssKey: task.OssKey,
				Detail: raw,
				Meta:   task.Meta,
			}
		case protocol.ActionRecognize:
			detail, err := d.recognize.Run(task)
			if err != nil {
				result = d.errorResult(task, err.Error())
				return
			}
			raw, _ := protocol.MarshalDetail(detail)
			result = &protocol.TaskResultMessage{
				Type:   protocol.TypeTaskResult,
				TaskID: task.TaskID,
				Status: protocol.StatusOK,
				Action: task.Action,
				OssKey: task.OssKey,
				Detail: raw,
				Meta:   task.Meta,
			}
		default:
			result = d.errorResult(task, "unknown action: "+task.Action)
		}
	}()

	select {
	case <-ctx.Done():
		d.sendError(task, "task timeout")
	case <-done:
		if result != nil {
			log.Printf("[worker] done task=%s status=%s action=%s", task.TaskID, result.Status, result.Action)
			d.send(result)
		}
	}
}

func (d *Dispatcher) errorResult(task *protocol.TaskMessage, msg string) *protocol.TaskResultMessage {
	return &protocol.TaskResultMessage{
		Type:     protocol.TypeTaskResult,
		TaskID:   task.TaskID,
		Status:   protocol.StatusError,
		Action:   task.Action,
		OssKey:   task.OssKey,
		ErrorMsg: msg,
		Meta:     task.Meta,
	}
}

func (d *Dispatcher) sendError(task *protocol.TaskMessage, msg string) {
	log.Printf("[worker] task %s error: %s", task.TaskID, msg)
	d.send(d.errorResult(task, msg))
}

func MustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

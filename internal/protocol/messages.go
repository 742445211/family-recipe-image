package protocol

import "encoding/json"

const (
	TypeTask       = "task"
	TypeTaskResult = "task_result"
	TypePing       = "ping"
	TypePong       = "pong"
	TypeRegistered = "registered"

	ActionCompress  = "compress"
	ActionRecognize = "recognize"

	StatusOK    = "ok"
	StatusError = "error"

	ScopeFridge = "fridge"
)

// TaskMeta 任务元数据；冰箱识别使用 scope=fridge + scan_id。
type TaskMeta struct {
	RecipeID uint64 `json:"recipe_id,omitempty"`
	Scope    string `json:"scope,omitempty"`
	ScanID   uint64 `json:"scan_id,omitempty"`
}

type TaskMessage struct {
	Type    string          `json:"type"`
	TaskID  string          `json:"task_id"`
	Action  string          `json:"action"`
	OssKey  string          `json:"oss_key"`
	OssURL  string          `json:"oss_url,omitempty"`
	Meta    json.RawMessage `json:"meta,omitempty"`
	Parsed  TaskMeta        `json:"-"`
}

type CompressDetail struct {
	Skipped         bool   `json:"skipped,omitempty"`
	Reason          string `json:"reason,omitempty"`
	Format          string `json:"format,omitempty"`
	OutputFormat    string `json:"output_format,omitempty"`
	NewOssKey       string `json:"new_oss_key,omitempty"`
	OriginalBytes   int64  `json:"original_bytes,omitempty"`
	CompressedBytes int64  `json:"compressed_bytes,omitempty"`
}

type RecognizeItem struct {
	Name   string `json:"name"`
	Amount string `json:"amount,omitempty"`
}

type RecognizeDetail struct {
	Ingredients []string        `json:"ingredients,omitempty"`
	Items       []RecognizeItem `json:"items,omitempty"`
}

type TaskResultMessage struct {
	Type     string          `json:"type"`
	TaskID   string          `json:"task_id"`
	Status   string          `json:"status"`
	Action   string          `json:"action"`
	OssKey   string          `json:"oss_key"`
	ErrorMsg string          `json:"error_msg,omitempty"`
	Detail   json.RawMessage `json:"detail,omitempty"`
	Meta     json.RawMessage `json:"meta,omitempty"`
}

type RegisteredMessage struct {
	Type     string `json:"type"`
	WorkerID string `json:"worker_id"`
}

type PingMessage struct {
	Type string `json:"type"`
}

type PongMessage struct {
	Type string `json:"type"`
}

func MarshalDetail(v any) (json.RawMessage, error) {
	return json.Marshal(v)
}

func ParseTaskMeta(raw json.RawMessage) TaskMeta {
	var m TaskMeta
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &m)
	}
	return m
}

func ParseTask(data []byte) (*TaskMessage, error) {
	var msg TaskMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	msg.Parsed = ParseTaskMeta(msg.Meta)
	return &msg, nil
}

func NewRecognizeDetail(names []string) *RecognizeDetail {
	items := make([]RecognizeItem, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		items = append(items, RecognizeItem{Name: name})
	}
	return &RecognizeDetail{
		Ingredients: names,
		Items:       items,
	}
}

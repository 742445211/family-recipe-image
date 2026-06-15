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
)

type TaskMeta struct {
	RecipeID uint64 `json:"recipe_id,omitempty"`
}

type TaskMessage struct {
	Type    string          `json:"type"`
	TaskID  string          `json:"task_id"`
	Action  string          `json:"action"`
	OssKey  string          `json:"oss_key"`
	OssURL  string          `json:"oss_url,omitempty"`
	Meta    TaskMeta        `json:"meta,omitempty"`
	RawMeta json.RawMessage `json:"-"`
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

type RecognizeDetail struct {
	Ingredients []string `json:"ingredients,omitempty"`
}

type TaskResultMessage struct {
	Type     string          `json:"type"`
	TaskID   string          `json:"task_id"`
	Status   string          `json:"status"`
	Action   string          `json:"action"`
	OssKey   string          `json:"oss_key"`
	ErrorMsg string          `json:"error_msg,omitempty"`
	Detail   json.RawMessage `json:"detail,omitempty"`
	Meta     TaskMeta        `json:"meta,omitempty"`
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

func ParseTask(data []byte) (*TaskMessage, error) {
	var msg TaskMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

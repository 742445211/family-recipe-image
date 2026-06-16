package protocol

import (
	"encoding/json"
	"testing"
)

func TestParseTaskFridgeMeta(t *testing.T) {
	raw := []byte(`{
		"type":"task","task_id":"abc","action":"recognize",
		"oss_key":"fridge/x.jpg","oss_url":"https://cdn/x.jpg",
		"meta":{"scope":"fridge","scan_id":3}
	}`)
	task, err := ParseTask(raw)
	if err != nil {
		t.Fatal(err)
	}
	if task.Parsed.Scope != ScopeFridge || task.Parsed.ScanID != 3 {
		t.Fatalf("meta=%+v", task.Parsed)
	}
	if string(task.Meta) == "" {
		t.Fatal("raw meta lost")
	}
}

func TestNewRecognizeDetail(t *testing.T) {
	d := NewRecognizeDetail([]string{"苹果", "牛奶"})
	if len(d.Ingredients) != 2 || len(d.Items) != 2 || d.Items[0].Name != "苹果" {
		t.Fatalf("%+v", d)
	}
	b, err := MarshalDetail(d)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["items"]; !ok {
		t.Fatalf("missing items: %s", string(b))
	}
}

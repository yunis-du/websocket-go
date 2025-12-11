package websocket

import (
	"encoding/json"
	"net/http"
	"testing"
)

type TestPayload struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

func TestStructMessage_JSON(t *testing.T) {
	payload := TestPayload{Foo: "hello", Bar: 123}
	msg := NewStructMessage("test-ns", "test-type", payload)

	bytes, err := msg.Bytes(JSONProtocol)
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	var decoded struct {
		Namespace string      `json:"namespace"`
		Type      string      `json:"type"`
		Obj       TestPayload `json:"obj"`
	}
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if decoded.Namespace != "test-ns" {
		t.Errorf("Expected namespace 'test-ns', got '%s'", decoded.Namespace)
	}
	if decoded.Obj.Foo != "hello" {
		t.Errorf("Expected payload foo 'hello', got '%s'", decoded.Obj.Foo)
	}
}

func TestStructMessage_Reply(t *testing.T) {
	msg := NewStructMessage("test-ns", "request", nil, "original-uuid")
	replyPayload := map[string]string{"result": "ok"}
	reply := msg.Reply(replyPayload, "response", http.StatusOK)

	if reply.Namespace != "test-ns" {
		t.Errorf("Expected reply namespace 'test-ns', got '%s'", reply.Namespace)
	}
	if reply.Uuid != "original-uuid" {
		t.Errorf("Expected reply uuid 'original-uuid', got '%s'", reply.Uuid)
	}
	if reply.Type != "response" {
		t.Errorf("Expected reply type 'response', got '%s'", reply.Type)
	}

	// Verify payload
	if val, ok := reply.Value.(map[string]string); !ok || val["result"] != "ok" {
		t.Errorf("Expected reply payload {'result': 'ok'}, got %v", reply.Value)
	}
}

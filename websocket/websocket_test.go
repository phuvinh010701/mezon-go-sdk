package websocket_test

import (
	"encoding/json"
	"testing"

	"github.com/phuvinh010701/mezon-go-sdk/websocket"
)

func TestNew_EmptyToken(t *testing.T) {
	_, err := websocket.New("")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func TestNew_ValidToken(t *testing.T) {
	conn, err := websocket.New("test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil {
		t.Fatal("expected non-nil Conn")
	}
}

func TestConn_IsOpen_NotConnected(t *testing.T) {
	conn, _ := websocket.New("test-token")
	if conn.IsOpen() {
		t.Error("expected IsOpen() == false before Connect")
	}
}

func TestEnvelope_MarshalRoundtrip(t *testing.T) {
	env := websocket.Envelope{
		CID:     "42",
		Type:    "channel_message",
		Payload: json.RawMessage(`{"id":"msg-1"}`),
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got websocket.Envelope
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.CID != "42" {
		t.Errorf("CID: got %q, want %q", got.CID, "42")
	}
	if got.Type != "channel_message" {
		t.Errorf("Type: got %q", got.Type)
	}
}

func TestConn_OnEvent_OffEvent(t *testing.T) {
	conn, _ := websocket.New("test-token")

	called := 0
	handler := func(payload json.RawMessage) { called++ }

	conn.OnEvent("channel_message", handler)
	conn.OffEvent("channel_message")
	// After OffEvent no handler should be registered; no panic expected.
	_ = called // avoid unused variable error
}

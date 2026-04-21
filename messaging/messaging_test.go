// Package messaging_test contains unit tests for the messaging package.
package messaging_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/phuvinh010701/mezon-go-sdk/client"
	"github.com/phuvinh010701/mezon-go-sdk/messaging"
	"github.com/phuvinh010701/mezon-go-sdk/models"
)

// newTestService spins up an httptest.Server with the given handler and returns
// a messaging.Service backed by a client pointing to that server.
func newTestService(t *testing.T, handler http.HandlerFunc) (*messaging.Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c, err := client.New(
		client.WithAPIKey("test-key"),
		client.WithBaseURL(srv.URL),
		client.WithHTTPClient(srv.Client()),
	)
	if err != nil {
		srv.Close()
		t.Fatalf("client.New: %v", err)
	}
	return messaging.New(c), srv
}

// ---- SendMessage ----

func TestSendMessage_NilRequest(t *testing.T) {
	svc, srv := newTestService(t, http.NotFound)
	defer srv.Close()

	_, err := svc.SendMessage(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestSendMessage_MissingChannelID(t *testing.T) {
	svc, srv := newTestService(t, http.NotFound)
	defer srv.Close()

	_, err := svc.SendMessage(context.Background(), &models.SendMessageRequest{
		Content: &models.MessageContent{T: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for missing ChannelID")
	}
}

func TestSendMessage_MissingContent(t *testing.T) {
	svc, srv := newTestService(t, http.NotFound)
	defer srv.Close()

	_, err := svc.SendMessage(context.Background(), &models.SendMessageRequest{
		ChannelID: "ch-1",
	})
	if err == nil {
		t.Fatal("expected error for missing Content")
	}
}

func TestSendMessage_OK(t *testing.T) {
	want := models.ChannelMessageAck{MessageID: "msg-99", ChannelID: "ch-1"}
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: got %s, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := svc.SendMessage(context.Background(), &models.SendMessageRequest{
		ChannelID: "ch-1",
		Mode:      models.StreamModeChannel,
		Content:   &models.MessageContent{T: "hello"},
		Code:      models.MessageTypeChat,
	})
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if got.MessageID != want.MessageID {
		t.Errorf("MessageID: got %q, want %q", got.MessageID, want.MessageID)
	}
}

// ---- SendEphemeralMessage ----

func TestSendEphemeralMessage_MissingReceiverIDs(t *testing.T) {
	svc, srv := newTestService(t, http.NotFound)
	defer srv.Close()

	_, err := svc.SendEphemeralMessage(context.Background(), &models.SendEphemeralRequest{
		ChannelID: "ch-1",
		Content:   &models.MessageContent{T: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for missing ReceiverIDs")
	}
}

// ---- EditMessage ----

func TestEditMessage_OK(t *testing.T) {
	want := models.ChannelMessageAck{MessageID: "msg-1"}
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method: got %s, want PUT", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := svc.EditMessage(context.Background(), &models.EditMessageRequest{
		MessageID: "msg-1",
		ChannelID: "ch-1",
		Mode:      models.StreamModeChannel,
		Content:   &models.MessageContent{T: "updated"},
	})
	if err != nil {
		t.Fatalf("EditMessage: %v", err)
	}
	if got.MessageID != "msg-1" {
		t.Errorf("MessageID: got %q, want msg-1", got.MessageID)
	}
}

// ---- DeleteMessage ----

func TestDeleteMessage_OK(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method: got %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	err := svc.DeleteMessage(context.Background(), &models.DeleteMessageRequest{
		MessageID: "msg-1",
		ChannelID: "ch-1",
		Mode:      models.StreamModeChannel,
	})
	if err != nil {
		t.Fatalf("DeleteMessage: %v", err)
	}
}

// ---- AddReaction ----

func TestAddReaction_MissingEmoji(t *testing.T) {
	svc, srv := newTestService(t, http.NotFound)
	defer srv.Close()

	err := svc.AddReaction(context.Background(), &models.AddReactionRequest{
		MessageID: "msg-1",
		ChannelID: "ch-1",
	})
	if err == nil {
		t.Fatal("expected error for missing Emoji")
	}
}

// ---- RemoveReaction ----

func TestRemoveReaction_DoesNotMutateCaller(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	req := &models.AddReactionRequest{
		MessageID:    "msg-1",
		ChannelID:    "ch-1",
		Emoji:        "👍",
		ActionDelete: false, // caller sets false
	}

	_ = svc.RemoveReaction(context.Background(), req)

	// Caller's struct must not be mutated.
	if req.ActionDelete {
		t.Error("RemoveReaction mutated caller's ActionDelete field")
	}
}

// ---- GetChannel ----

func TestGetChannel_MissingID(t *testing.T) {
	svc, srv := newTestService(t, http.NotFound)
	defer srv.Close()

	_, err := svc.GetChannel(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty channelID")
	}
}

func TestGetChannel_OK(t *testing.T) {
	want := models.ChannelDetail{ChannelID: "ch-42", ChannelLabel: "general"}
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := svc.GetChannel(context.Background(), "ch-42")
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	if got.ChannelLabel != "general" {
		t.Errorf("ChannelLabel: got %q, want general", got.ChannelLabel)
	}
}

// ---- ListChannels ----

func TestListChannels_OK(t *testing.T) {
	want := models.ChannelDescList{
		ChannelDescs: []models.ChannelDescription{
			{ChannelID: "ch-1", ChannelLabel: "alpha"},
		},
	}
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := svc.ListChannels(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListChannels: %v", err)
	}
	if len(got.ChannelDescs) != 1 || got.ChannelDescs[0].ChannelID != "ch-1" {
		t.Errorf("unexpected channel list: %+v", got.ChannelDescs)
	}
}

// ---- CreateChannel ----

func TestCreateChannel_NilRequest(t *testing.T) {
	svc, srv := newTestService(t, http.NotFound)
	defer srv.Close()

	_, err := svc.CreateChannel(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}

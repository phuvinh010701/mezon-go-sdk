package models_test

import (
	"encoding/json"
	"testing"

	"github.com/phuvinh010701/mezon-go-sdk/models"
)

func TestChannelMessageMarshalRoundtrip(t *testing.T) {
	msg := models.ChannelMessage{
		ID:        "msg-1",
		ChannelID: "ch-1",
		ClanID:    "clan-1",
		SenderID:  "user-1",
		Content:   &models.MessageContent{T: "Hello, World!"},
		Code:      models.MessageTypeChat,
		Mode:      models.StreamModeChannel,
		Mentions: []models.MessageMention{
			{UserID: "user-2", Username: "alice"},
		},
		Attachments: []models.MessageAttachment{
			{URL: "https://example.com/img.png", Type: "image/png", Size: 2048, Name: "img.png"},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got models.ChannelMessage
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != msg.ID {
		t.Errorf("ID: got %q, want %q", got.ID, msg.ID)
	}
	if got.Content == nil || got.Content.T != "Hello, World!" {
		t.Errorf("Content.T: got %v, want %q", got.Content, "Hello, World!")
	}
	if len(got.Mentions) != 1 || got.Mentions[0].Username != "alice" {
		t.Errorf("Mentions: got %v", got.Mentions)
	}
}

func TestSendMessageRequestJSONTags(t *testing.T) {
	req := models.SendMessageRequest{
		ChannelID: "ch-42",
		Mode:      models.StreamModeDM,
		Content:   &models.MessageContent{T: "hi"},
		Code:      models.MessageTypeChat,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	if _, ok := m["channel_id"]; !ok {
		t.Error("expected json key channel_id")
	}
}

func TestMessageEmbedRoundtrip(t *testing.T) {
	color := 0xFF5500
	embed := models.MessageEmbed{
		Title:       "Test Embed",
		Description: "desc",
		Color:       &color,
		Fields: []models.EmbedField{
			{Name: "field1", Value: "val1", Inline: true},
		},
		Footer:    &models.EmbedFooter{Text: "footer text"},
		Thumbnail: &models.EmbedThumbnail{URL: "https://example.com/thumb.png"},
	}

	data, err := json.Marshal(embed)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got models.MessageEmbed
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Title != "Test Embed" {
		t.Errorf("Title: got %q", got.Title)
	}
	if got.Color == nil || *got.Color != color {
		t.Errorf("Color: got %v, want %d", got.Color, color)
	}
}

func TestChannelTypeConstants(t *testing.T) {
	tests := []struct {
		ct   models.ChannelType
		want int
	}{
		{models.ChannelTypeChannel, 1},
		{models.ChannelTypeDM, 3},
		{models.ChannelTypeMezonVoice, 10},
	}
	for _, tt := range tests {
		if int(tt.ct) != tt.want {
			t.Errorf("ChannelType %v: got %d, want %d", tt.ct, int(tt.ct), tt.want)
		}
	}
}

func TestEventConstants(t *testing.T) {
	if models.EventChannelMessage != "channel_message" {
		t.Errorf("EventChannelMessage: got %q", models.EventChannelMessage)
	}
	if models.EventMessageButtonClicked != "message_button_clicked" {
		t.Errorf("EventMessageButtonClicked: got %q", models.EventMessageButtonClicked)
	}
}

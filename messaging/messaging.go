// Package messaging provides methods for sending, editing, and deleting
// messages and managing channels in the Mezon API.
//
// Usage:
//
//	c, _ := client.New(client.WithAPIKey("..."))
//	svc := messaging.New(c)
//
//	ctx := context.Background()
//	ack, err := svc.SendMessage(ctx, &models.SendMessageRequest{
//	    ChannelID: "123",
//	    Mode:      models.StreamModeChannel,
//	    Content:   &models.MessageContent{T: "Hello!"},
//	    Code:      models.MessageTypeChat,
//	})
package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/phuvinh010701/mezon-go-sdk/client"
	"github.com/phuvinh010701/mezon-go-sdk/models"
)

// Service provides message and channel API operations.
type Service struct {
	c *client.Client
}

// New creates a new messaging Service backed by the given client.
func New(c *client.Client) *Service {
	return &Service{c: c}
}

// -------------------------------------------------------------------
// Message operations
// -------------------------------------------------------------------

// SendMessage sends a new message to a channel.
// Returns a ChannelMessageAck containing the new message ID.
func (s *Service) SendMessage(ctx context.Context, req *models.SendMessageRequest) (*models.ChannelMessageAck, error) {
	if req == nil {
		return nil, errNilRequest("SendMessageRequest")
	}
	if req.ChannelID == "" {
		return nil, errMissingField("ChannelID")
	}
	if req.Content == nil {
		return nil, errMissingField("Content")
	}

	body, err := jsonBody(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := s.c.HTTP().NewRequest(ctx, http.MethodPost, "/channels/messages", body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	var ack models.ChannelMessageAck
	if err := s.c.HTTP().Do(ctx, httpReq, &ack); err != nil {
		return nil, err
	}
	return &ack, nil
}

// SendEphemeralMessage sends a message visible only to specific users.
func (s *Service) SendEphemeralMessage(ctx context.Context, req *models.SendEphemeralRequest) (*models.ChannelMessageAck, error) {
	if req == nil {
		return nil, errNilRequest("SendEphemeralRequest")
	}
	if req.ChannelID == "" {
		return nil, errMissingField("ChannelID")
	}
	if len(req.ReceiverIDs) == 0 {
		return nil, errMissingField("ReceiverIDs")
	}
	if req.Content == nil {
		return nil, errMissingField("Content")
	}

	body, err := jsonBody(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := s.c.HTTP().NewRequest(ctx, http.MethodPost, "/channels/messages/ephemeral", body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	var ack models.ChannelMessageAck
	if err := s.c.HTTP().Do(ctx, httpReq, &ack); err != nil {
		return nil, err
	}
	return &ack, nil
}

// EditMessage updates the content of an existing message.
func (s *Service) EditMessage(ctx context.Context, req *models.EditMessageRequest) (*models.ChannelMessageAck, error) {
	if req == nil {
		return nil, errNilRequest("EditMessageRequest")
	}
	if req.MessageID == "" {
		return nil, errMissingField("MessageID")
	}
	if req.ChannelID == "" {
		return nil, errMissingField("ChannelID")
	}
	if req.Content == nil {
		return nil, errMissingField("Content")
	}

	body, err := jsonBody(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := s.c.HTTP().NewRequest(ctx, http.MethodPut, "/channels/messages", body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	var ack models.ChannelMessageAck
	if err := s.c.HTTP().Do(ctx, httpReq, &ack); err != nil {
		return nil, err
	}
	return &ack, nil
}

// DeleteMessage removes a message from a channel.
func (s *Service) DeleteMessage(ctx context.Context, req *models.DeleteMessageRequest) error {
	if req == nil {
		return errNilRequest("DeleteMessageRequest")
	}
	if req.MessageID == "" {
		return errMissingField("MessageID")
	}
	if req.ChannelID == "" {
		return errMissingField("ChannelID")
	}

	body, err := jsonBody(req)
	if err != nil {
		return err
	}
	httpReq, err := s.c.HTTP().NewRequest(ctx, http.MethodDelete, "/channels/messages", body)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	return s.c.HTTP().Do(ctx, httpReq, nil)
}

// AddReaction adds an emoji reaction to a message.
func (s *Service) AddReaction(ctx context.Context, req *models.AddReactionRequest) error {
	if req == nil {
		return errNilRequest("AddReactionRequest")
	}
	if req.MessageID == "" {
		return errMissingField("MessageID")
	}
	if req.ChannelID == "" {
		return errMissingField("ChannelID")
	}
	if req.Emoji == "" {
		return errMissingField("Emoji")
	}

	body, err := jsonBody(req)
	if err != nil {
		return err
	}
	httpReq, err := s.c.HTTP().NewRequest(ctx, http.MethodPost, "/channels/messages/emoji", body)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	return s.c.HTTP().Do(ctx, httpReq, nil)
}

// RemoveReaction removes an emoji reaction from a message.
func (s *Service) RemoveReaction(ctx context.Context, req *models.AddReactionRequest) error {
	if req == nil {
		return errNilRequest("AddReactionRequest")
	}
	// Copy to avoid mutating the caller's struct.
	r := *req
	r.ActionDelete = true
	return s.AddReaction(ctx, &r)
}

// -------------------------------------------------------------------
// Channel operations
// -------------------------------------------------------------------

// GetChannel fetches detailed information about a single channel.
func (s *Service) GetChannel(ctx context.Context, channelID string) (*models.ChannelDetail, error) {
	if channelID == "" {
		return nil, errMissingField("channelID")
	}

	httpReq, err := s.c.HTTP().NewRequest(ctx, http.MethodGet, "/channels/"+channelID, nil)
	if err != nil {
		return nil, err
	}

	var detail models.ChannelDetail
	if err := s.c.HTTP().Do(ctx, httpReq, &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

// ListChannels returns channels matching the given filter request.
func (s *Service) ListChannels(ctx context.Context, req *models.ListChannelsRequest) (*models.ChannelDescList, error) {
	if req == nil {
		req = &models.ListChannelsRequest{}
	}

	body, err := jsonBody(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := s.c.HTTP().NewRequest(ctx, http.MethodPost, "/channels/list", body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	var list models.ChannelDescList
	if err := s.c.HTTP().Do(ctx, httpReq, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// CreateChannel creates a new channel (or DM) and returns its description.
func (s *Service) CreateChannel(ctx context.Context, req *models.CreateChannelRequest) (*models.ChannelDescription, error) {
	if req == nil {
		return nil, errNilRequest("CreateChannelRequest")
	}

	body, err := jsonBody(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := s.c.HTTP().NewRequest(ctx, http.MethodPost, "/channels", body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	var desc models.ChannelDescription
	if err := s.c.HTTP().Do(ctx, httpReq, &desc); err != nil {
		return nil, err
	}
	return &desc, nil
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

// jsonBody marshals v and returns an io.Reader suitable for use as an HTTP request body.
func jsonBody(v any) (io.Reader, error) {
	buf, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}

func errNilRequest(name string) error {
	return fmt.Errorf("messaging: %s must not be nil", name)
}

func errMissingField(field string) error {
	return fmt.Errorf("messaging: %s is required", field)
}

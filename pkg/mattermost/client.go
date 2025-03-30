package mattermost

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"vk-test-assignment-mattermost-polls/pkg/config"
)

type Client struct {
	URL        string
	Token      string
	HTTPClient *http.Client
}

func NewClient(cfg config.MattermostConfig) *Client {
	return &Client{
		URL:   cfg.URL,
		Token: cfg.Token,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SendChannelMessage(channelID, message string) error {
	type postRequest struct {
		ChannelID string `json:"channel_id"`
		Message   string `json:"message"`
	}

	payload := postRequest{
		ChannelID: channelID,
		Message:   message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	url := fmt.Sprintf("%s/api/v4/posts", c.URL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to send message: status code %d", resp.StatusCode)
	}

	log.Debug().
		Str("channel_id", channelID).
		Str("message", message).
		Msg("Message sent to channel")

	return nil
}

func (c *Client) SendPollEndedNotification(channelID, pollID, question string) error {
	message := fmt.Sprintf("Poll time ended: \"%s\". Use `/poll results %s` to see the final results.", question, pollID)
	return c.SendChannelMessage(channelID, message)
}

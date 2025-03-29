package dto

const (
	ResponseTypeEphemeral = "ephemeral"
	ResponseTypeInChannel = "in_channel"
)

type MattermostResponse struct {
	ResponseType string      `json:"response_type"`
	Text         string      `json:"text"`
	Props        interface{} `json:"props,omitempty"`
}

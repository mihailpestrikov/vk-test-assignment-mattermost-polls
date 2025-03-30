package dto

type MattermostCommandRequest struct {
	Token       string `form:"token" validate:"required"`
	TeamID      string `form:"team_id" validate:"required"`
	TeamDomain  string `form:"team_domain"`
	ChannelID   string `form:"channel_id" validate:"required"`
	ChannelName string `form:"channel_name"`
	UserID      string `form:"user_id" validate:"required"`
	UserName    string `form:"user_name"`
	Command     string `form:"command" validate:"required"`
	Text        string `form:"text"`
	ResponseURL string `form:"response_url"`
	TriggerID   string `form:"trigger_id"`
}

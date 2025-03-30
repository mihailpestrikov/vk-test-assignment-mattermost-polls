package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog/log"

	mockservice "vk-test-assignment-mattermost-polls/internal/mocks/service"
	"vk-test-assignment-mattermost-polls/internal/model"
	"vk-test-assignment-mattermost-polls/internal/service"
	"vk-test-assignment-mattermost-polls/pkg/config"
	"vk-test-assignment-mattermost-polls/pkg/mattermost"
)

func createTestHandler(t *testing.T) (*Handler, *mockservice.MockIPollService, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockService := mockservice.NewMockIPollService(ctrl)

	cfg := config.MattermostConfig{
		WebhookSecret: "test_secret",
	}

	handler := &Handler{
		pollService:      &service.PollService{},
		mattermostCfg:    cfg,
		mattermostClient: mattermost.NewClient(cfg),
	}

	handler.pollService = mockService

	return handler, mockService, ctrl
}

func createFormRequest(values url.Values) *http.Request {
	req := httptest.NewRequest("POST", "/command", strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func TestHandler_handleCommand_InvalidToken(t *testing.T) {
	handler, _, ctrl := createTestHandler(t)
	defer ctrl.Finish()

	values := url.Values{}
	values.Add("token", "invalid_token")
	values.Add("team_id", "team1")
	values.Add("channel_id", "channel1")
	values.Add("user_id", "user1")
	values.Add("command", "/poll")
	values.Add("text", "create \"Test Question\" \"Option 1\" \"Option 2\"")

	w := httptest.NewRecorder()
	req := createFormRequest(values)

	handler.handleCommand(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestHandler_handleCommand_CreatePoll(t *testing.T) {
	handler, mockService, ctrl := createTestHandler(t)
	defer ctrl.Finish()

	poll := &model.Poll{
		ID:        "poll123",
		Question:  "Test Question",
		Options:   []string{"Option 1", "Option 2"},
		CreatedBy: "user1",
		ChannelID: "channel1",
		Status:    model.PollStatusActive,
	}

	mockService.EXPECT().
		CreatePoll("Test Question", []string{"Option 1", "Option 2"}, "user1", "channel1", 0).
		Return(poll, nil).
		Times(1)

	values := url.Values{}
	values.Add("token", "test_secret")
	values.Add("team_id", "team1")
	values.Add("channel_id", "channel1")
	values.Add("user_id", "user1")
	values.Add("command", "/poll")
	values.Add("text", "create \"Test Question\" \"Option 1\" \"Option 2\"")

	w := httptest.NewRecorder()
	req := createFormRequest(values)

	handler.handleCommand(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_handleCommand_Vote(t *testing.T) {
	handler, mockService, ctrl := createTestHandler(t)
	defer ctrl.Finish()

	poll := &model.Poll{
		ID:        "poll123",
		Question:  "Test Question",
		Options:   []string{"Option 1", "Option 2"},
		CreatedBy: "user1",
		ChannelID: "channel1",
		Status:    model.PollStatusActive,
	}

	mockService.EXPECT().
		GetPoll("poll123").
		Return(poll, nil).
		Times(1)

	mockService.EXPECT().
		Vote("poll123", "user1", 0).
		Return(nil).
		Times(1)

	values := url.Values{}
	values.Add("token", "test_secret")
	values.Add("team_id", "team1")
	values.Add("channel_id", "channel1")
	values.Add("user_id", "user1")
	values.Add("command", "/poll")
	values.Add("text", "vote poll123 1")

	w := httptest.NewRecorder()
	req := createFormRequest(values)

	handler.handleCommand(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_handleCommand_Results(t *testing.T) {
	handler, mockService, ctrl := createTestHandler(t)
	defer ctrl.Finish()

	results := &service.VoteResults{
		PollID:     "poll123",
		Question:   "Test Question",
		TotalVotes: 2,
		Results: []service.VoteCountResult{
			{OptionIndex: 0, OptionText: "Option 1", Count: 1},
			{OptionIndex: 1, OptionText: "Option 2", Count: 1},
		},
		IsActive: true,
	}

	poll := &model.Poll{
		ID:        "poll123",
		Question:  "Test Question",
		Options:   []string{"Option 1", "Option 2"},
		CreatedBy: "user2",
		ChannelID: "channel1",
		Status:    model.PollStatusActive,
	}

	mockService.EXPECT().
		GetResults("poll123").
		Return(results, nil).
		Times(1)

	mockService.EXPECT().
		GetPoll("poll123").
		Return(poll, nil).
		Times(1)

	values := url.Values{}
	values.Add("token", "test_secret")
	values.Add("team_id", "team1")
	values.Add("channel_id", "channel1")
	values.Add("user_id", "user1")
	values.Add("command", "/poll")
	values.Add("text", "results poll123")

	w := httptest.NewRecorder()
	req := createFormRequest(values)

	handler.handleCommand(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_handleCommand_EndPoll(t *testing.T) {
	handler, mockService, ctrl := createTestHandler(t)
	defer ctrl.Finish()

	results := &service.VoteResults{
		PollID:     "poll123",
		Question:   "Test Question",
		TotalVotes: 2,
		Results: []service.VoteCountResult{
			{OptionIndex: 0, OptionText: "Option 1", Count: 1},
			{OptionIndex: 1, OptionText: "Option 2", Count: 1},
		},
		IsActive: false,
	}

	mockService.EXPECT().
		EndPoll("poll123", "user1").
		Return(results, nil).
		Times(1)

	values := url.Values{}
	values.Add("token", "test_secret")
	values.Add("team_id", "team1")
	values.Add("channel_id", "channel1")
	values.Add("user_id", "user1")
	values.Add("command", "/poll")
	values.Add("text", "end poll123")

	w := httptest.NewRecorder()
	req := createFormRequest(values)

	handler.handleCommand(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_handleCommand_EndPollNotCreator(t *testing.T) {
	handler, mockService, ctrl := createTestHandler(t)
	defer ctrl.Finish()

	mockService.EXPECT().
		EndPoll("poll123", "user1").
		Return(nil, model.ErrNotPollCreator).
		Times(1)

	values := url.Values{}
	values.Add("token", "test_secret")
	values.Add("team_id", "team1")
	values.Add("channel_id", "channel1")
	values.Add("user_id", "user1")
	values.Add("command", "/poll")
	values.Add("text", "end poll123")

	w := httptest.NewRecorder()
	req := createFormRequest(values)

	handler.handleCommand(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_handleCommand_DeletePoll(t *testing.T) {
	handler, mockService, ctrl := createTestHandler(t)
	defer ctrl.Finish()

	mockService.EXPECT().
		DeletePoll("poll123", "user1").
		Return(nil).
		Times(1)

	values := url.Values{}
	values.Add("token", "test_secret")
	values.Add("team_id", "team1")
	values.Add("channel_id", "channel1")
	values.Add("user_id", "user1")
	values.Add("command", "/poll")
	values.Add("text", "delete poll123")

	w := httptest.NewRecorder()
	req := createFormRequest(values)

	handler.handleCommand(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_handleCommand_DeletePollNotCreator(t *testing.T) {
	handler, mockService, ctrl := createTestHandler(t)
	defer ctrl.Finish()

	mockService.EXPECT().
		DeletePoll("poll123", "user1").
		Return(model.ErrNotPollCreator).
		Times(1)

	values := url.Values{}
	values.Add("token", "test_secret")
	values.Add("team_id", "team1")
	values.Add("channel_id", "channel1")
	values.Add("user_id", "user1")
	values.Add("command", "/poll")
	values.Add("text", "delete poll123")

	w := httptest.NewRecorder()
	req := createFormRequest(values)

	handler.handleCommand(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_handleCommand_Info(t *testing.T) {
	handler, mockService, ctrl := createTestHandler(t)
	defer ctrl.Finish()

	poll := &model.Poll{
		ID:        "poll123",
		Question:  "Test Question",
		Options:   []string{"Option 1", "Option 2"},
		CreatedBy: "user1",
		ChannelID: "channel1",
		Status:    model.PollStatusActive,
	}

	mockService.EXPECT().
		GetPoll("poll123").
		Return(poll, nil).
		Times(1)

	values := url.Values{}
	values.Add("token", "test_secret")
	values.Add("team_id", "team1")
	values.Add("channel_id", "channel1")
	values.Add("user_id", "user1")
	values.Add("command", "/poll")
	values.Add("text", "info poll123")

	w := httptest.NewRecorder()
	req := createFormRequest(values)

	handler.handleCommand(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_handleCommand_Help(t *testing.T) {
	handler, _, ctrl := createTestHandler(t)
	defer ctrl.Finish()

	values := url.Values{}
	values.Add("token", "test_secret")
	values.Add("team_id", "team1")
	values.Add("channel_id", "channel1")
	values.Add("user_id", "user1")
	values.Add("command", "/poll")
	values.Add("text", "help")

	w := httptest.NewRecorder()
	req := createFormRequest(values)

	handler.handleCommand(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_handleCommand_InvalidCommand(t *testing.T) {
	handler, _, ctrl := createTestHandler(t)
	defer ctrl.Finish()

	values := url.Values{}
	values.Add("token", "test_secret")
	values.Add("team_id", "team1")
	values.Add("channel_id", "channel1")
	values.Add("user_id", "user1")
	values.Add("command", "/poll")
	values.Add("text", "invalid_command")

	w := httptest.NewRecorder()
	req := createFormRequest(values)

	handler.handleCommand(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_handleCommand_ParseError(t *testing.T) {
	handler, _, ctrl := createTestHandler(t)
	defer ctrl.Finish()

	log.Logger = log.Level(0)

	values := url.Values{}
	values.Add("token", "test_secret")
	values.Add("team_id", "team1")
	values.Add("channel_id", "channel1")
	values.Add("user_id", "user1")
	values.Add("command", "/poll")
	values.Add("text", "create \"Incomplete")

	w := httptest.NewRecorder()
	req := createFormRequest(values)

	handler.handleCommand(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

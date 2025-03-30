package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	mocks "vk-test-assignment-mattermost-polls/internal/mocks/repository"
	"vk-test-assignment-mattermost-polls/internal/model"
	"vk-test-assignment-mattermost-polls/pkg/config"
)

func TestNewPollService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pollConfig := config.PollConfig{
		DefaultDuration: 3600,
		MaxOptions:      10,
	}

	type args struct {
		repo       Repository
		pollConfig config.PollConfig
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Create new poll service",
			args: args{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPollService(tt.args.repo, tt.args.pollConfig)
			if got == nil {
				t.Errorf("NewPollService() returned nil")
			}
			if got.repo != tt.args.repo {
				t.Errorf("NewPollService().repo = %v, want %v", got.repo, tt.args.repo)
			}
			if !reflect.DeepEqual(got.pollConfig, tt.args.pollConfig) {
				t.Errorf("NewPollService().pollConfig = %v, want %v", got.pollConfig, tt.args.pollConfig)
			}
		})
	}
}

func TestPollService_DeletePoll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pollConfig := config.PollConfig{
		DefaultDuration: 3600,
		MaxOptions:      10,
	}

	now := time.Now().Unix()
	future := now + 3600

	activePoll := &model.Poll{
		ID:        "poll123",
		Question:  "Active Poll",
		Options:   []string{"Option 1", "Option 2"},
		CreatedBy: "user123",
		ChannelID: "channel456",
		CreatedAt: now,
		ExpiresAt: future,
		Status:    model.PollStatusActive,
	}

	mockRepo.EXPECT().
		GetPoll("poll123").
		Return(activePoll, nil).
		Times(2)

	mockRepo.EXPECT().
		GetPoll("not_found").
		Return(nil, model.ErrPollNotFound).
		Times(1)

	mockRepo.EXPECT().
		GetPoll("error_update").
		Return(activePoll, nil).
		Times(1)

	mockRepo.EXPECT().
		UpdatePollStatus("poll123", model.PollStatusDeleted).
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		UpdatePollStatus("error_update", model.PollStatusDeleted).
		Return(errors.New("update error")).
		Times(1)

	type fields struct {
		repo       Repository
		pollConfig config.PollConfig
	}
	type args struct {
		pollID string
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Delete poll by creator",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID: "poll123",
				userID: "user123",
			},
			wantErr: false,
		},
		{
			name: "Delete poll by non-creator",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID: "poll123",
				userID: "user456",
			},
			wantErr: true,
		},
		{
			name: "Delete non-existent poll",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID: "not_found",
				userID: "user123",
			},
			wantErr: true,
		},
		{
			name: "Error updating poll status",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID: "error_update",
				userID: "user123",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PollService{
				repo:       tt.fields.repo,
				pollConfig: tt.fields.pollConfig,
			}
			if err := s.DeletePoll(tt.args.pollID, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("DeletePoll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPollService_FinishExpiredPolls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pollConfig := config.PollConfig{
		DefaultDuration: 3600,
		MaxOptions:      10,
	}

	now := time.Now().Unix()
	past := now - 3600

	expiredPoll1 := &model.Poll{
		ID:        "expired1",
		Question:  "Expired Poll 1",
		Options:   []string{"Option 1", "Option 2"},
		CreatedBy: "user123",
		ChannelID: "channel456",
		CreatedAt: now - 7200,
		ExpiresAt: past,
		Status:    model.PollStatusActive,
	}

	expiredPoll2 := &model.Poll{
		ID:        "expired2",
		Question:  "Expired Poll 2",
		Options:   []string{"Option 1", "Option 2"},
		CreatedBy: "user123",
		ChannelID: "channel456",
		CreatedAt: now - 7200,
		ExpiresAt: past,
		Status:    model.PollStatusActive,
	}

	mockRepo.EXPECT().
		GetExpiredActivePolls().
		Return([]*model.Poll{expiredPoll1, expiredPoll2}, nil).
		Times(1)

	mockRepo.EXPECT().
		UpdatePollStatus("expired1", model.PollStatusClosed).
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		UpdatePollStatus("expired2", model.PollStatusClosed).
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		GetExpiredActivePolls().
		Return(nil, errors.New("database error")).
		Times(1)

	type fields struct {
		repo       Repository
		pollConfig config.PollConfig
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Successfully finish expired polls",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			wantErr: false,
		},
		{
			name: "Error getting expired polls",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PollService{
				repo:       tt.fields.repo,
				pollConfig: tt.fields.pollConfig,
			}
			if err := s.FinishExpiredPolls(); (err != nil) != tt.wantErr {
				t.Errorf("FinishExpiredPolls() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPollService_StartPollWatcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pollConfig := config.PollConfig{
		DefaultDuration: 3600,
		MaxOptions:      10,
	}

	type fields struct {
		repo       Repository
		pollConfig config.PollConfig
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Start poll watcher",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				ctx: context.Background(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PollService{
				repo:       tt.fields.repo,
				pollConfig: tt.fields.pollConfig,
			}

			ctx, cancel := context.WithCancel(tt.args.ctx)
			defer cancel()

			s.StartPollWatcher(ctx)

			time.Sleep(100 * time.Millisecond)

			cancel()
		})
	}
}

func TestPollService_StartPollCleaner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pollConfig := config.PollConfig{
		DefaultDuration: 3600,
		MaxOptions:      10,
	}

	type fields struct {
		repo       Repository
		pollConfig config.PollConfig
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Start poll cleaner",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				ctx: context.Background(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PollService{
				repo:       tt.fields.repo,
				pollConfig: tt.fields.pollConfig,
			}

			ctx, cancel := context.WithCancel(tt.args.ctx)
			defer cancel()

			s.StartPollCleaner(ctx)

			time.Sleep(100 * time.Millisecond)

			cancel()
		})
	}
}

func TestPollService_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pollConfig := config.PollConfig{
		DefaultDuration: 3600,
		MaxOptions:      10,
	}

	mockRepo.EXPECT().
		Close().
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		Close().
		Return(errors.New("close error")).
		Times(1)

	type fields struct {
		repo       Repository
		pollConfig config.PollConfig
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Close without error",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			wantErr: false,
		},
		{
			name: "Close with error",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PollService{
				repo:       tt.fields.repo,
				pollConfig: tt.fields.pollConfig,
			}
			if err := s.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPollService_CreatePoll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pollConfig := config.PollConfig{
		DefaultDuration: 3600,
		MaxOptions:      10,
	}

	mockRepo.EXPECT().
		CreatePoll(gomock.Any()).
		DoAndReturn(func(poll *model.Poll) error {
			if poll.Question != "Test Question" {
				t.Errorf("Expected poll question 'Test Question', got '%s'", poll.Question)
			}
			if !reflect.DeepEqual(poll.Options, []string{"Option 1", "Option 2"}) {
				t.Errorf("Expected poll options ['Option 1', 'Option 2'], got %v", poll.Options)
			}
			if poll.CreatedBy != "user123" {
				t.Errorf("Expected creator 'user123', got '%s'", poll.CreatedBy)
			}
			if poll.ChannelID != "channel456" {
				t.Errorf("Expected channel 'channel456', got '%s'", poll.ChannelID)
			}
			if poll.Status != model.PollStatusActive {
				t.Errorf("Expected status 'ACTIVE', got '%s'", poll.Status)
			}
			return nil
		}).Times(2)

	mockRepo.EXPECT().
		CreatePoll(gomock.Any()).
		Return(errors.New("repository error")).
		Times(1)

	type fields struct {
		repo       Repository
		pollConfig config.PollConfig
	}
	type args struct {
		question  string
		options   []string
		createdBy string
		channelID string
		duration  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Create poll with valid data",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				question:  "Test Question",
				options:   []string{"Option 1", "Option 2"},
				createdBy: "user123",
				channelID: "channel456",
				duration:  7200,
			},
			wantErr: false,
		},
		{
			name: "Create poll with default duration",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				question:  "Test Question",
				options:   []string{"Option 1", "Option 2"},
				createdBy: "user123",
				channelID: "channel456",
				duration:  0,
			},
			wantErr: false,
		},
		{
			name: "Create poll with empty question",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				question:  "",
				options:   []string{"Option 1", "Option 2"},
				createdBy: "user123",
				channelID: "channel456",
				duration:  3600,
			},
			wantErr: true,
		},
		{
			name: "Create poll with too few options",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				question:  "Test Question",
				options:   []string{"Option 1"},
				createdBy: "user123",
				channelID: "channel456",
				duration:  3600,
			},
			wantErr: true,
		},
		{
			name: "Create poll with too many options",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				question:  "Test Question",
				options:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
				createdBy: "user123",
				channelID: "channel456",
				duration:  3600,
			},
			wantErr: true,
		},
		{
			name: "Create poll with duplicate options",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				question:  "Test Question",
				options:   []string{"Option 1", "Option 1"},
				createdBy: "user123",
				channelID: "channel456",
				duration:  3600,
			},
			wantErr: true,
		},
		{
			name: "Repository error",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				question:  "Test Question",
				options:   []string{"Option 1", "Option 2"},
				createdBy: "user123",
				channelID: "channel456",
				duration:  3600,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PollService{
				repo:       tt.fields.repo,
				pollConfig: tt.fields.pollConfig,
			}
			_, err := s.CreatePoll(tt.args.question, tt.args.options, tt.args.createdBy, tt.args.channelID, tt.args.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePoll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPollService_GetPoll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pollConfig := config.PollConfig{
		DefaultDuration: 3600,
		MaxOptions:      10,
	}

	now := time.Now().Unix()
	future := now + 3600
	past := now - 3600

	activePoll := &model.Poll{
		ID:        "poll123",
		Question:  "Active Poll",
		Options:   []string{"Option 1", "Option 2"},
		CreatedBy: "user123",
		ChannelID: "channel456",
		CreatedAt: now,
		ExpiresAt: future,
		Status:    model.PollStatusActive,
	}

	expiredPoll := &model.Poll{
		ID:        "poll456",
		Question:  "Expired Poll",
		Options:   []string{"Option 1", "Option 2"},
		CreatedBy: "user123",
		ChannelID: "channel456",
		CreatedAt: now,
		ExpiresAt: past,
		Status:    model.PollStatusActive,
	}

	mockRepo.EXPECT().
		GetPoll("poll123").
		Return(activePoll, nil).
		Times(1)

	mockRepo.EXPECT().
		GetPoll("poll456").
		Return(expiredPoll, nil).
		Times(1)

	mockRepo.EXPECT().
		GetPoll("notfound").
		Return(nil, model.ErrPollNotFound).
		Times(1)

	mockRepo.EXPECT().
		UpdatePollStatus("poll456", model.PollStatusClosed).
		Return(nil).
		Times(1)

	type fields struct {
		repo       Repository
		pollConfig config.PollConfig
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Poll
		wantErr bool
	}{
		{
			name: "Get active poll",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				id: "poll123",
			},
			want:    activePoll,
			wantErr: false,
		},
		{
			name: "Get expired poll - should auto close",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				id: "poll456",
			},
			want: &model.Poll{
				ID:        "poll456",
				Question:  "Expired Poll",
				Options:   []string{"Option 1", "Option 2"},
				CreatedBy: "user123",
				ChannelID: "channel456",
				CreatedAt: now,
				ExpiresAt: past,
				Status:    model.PollStatusClosed,
			},
			wantErr: false,
		},
		{
			name: "Get non-existent poll",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				id: "notfound",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PollService{
				repo:       tt.fields.repo,
				pollConfig: tt.fields.pollConfig,
			}
			got, err := s.GetPoll(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPoll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.ID != tt.want.ID {
				t.Errorf("GetPoll().ID = %v, want %v", got.ID, tt.want.ID)
			}
			if tt.args.id == "poll456" {

				if got.Status != model.PollStatusClosed {
					t.Errorf("GetPoll().Status = %v, want %v", got.Status, model.PollStatusClosed)
				}
			} else {
				if got.Status != tt.want.Status {
					t.Errorf("GetPoll().Status = %v, want %v", got.Status, tt.want.Status)
				}
			}
		})
	}
}

func TestPollService_Vote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pollConfig := config.PollConfig{
		DefaultDuration: 3600,
		MaxOptions:      10,
	}

	now := time.Now().Unix()
	future := now + 3600

	activePoll := &model.Poll{
		ID:        "poll123",
		Question:  "Active Poll",
		Options:   []string{"Option 1", "Option 2", "Option 3"},
		CreatedBy: "user123",
		ChannelID: "channel456",
		CreatedAt: now,
		ExpiresAt: future,
		Status:    model.PollStatusActive,
	}

	closedPoll := &model.Poll{
		ID:        "poll456",
		Question:  "Closed Poll",
		Options:   []string{"Option 1", "Option 2"},
		CreatedBy: "user123",
		ChannelID: "channel456",
		CreatedAt: now,
		ExpiresAt: future,
		Status:    model.PollStatusClosed,
	}

	mockRepo.EXPECT().
		GetPoll("poll123").
		Return(activePoll, nil).
		Times(3)

	mockRepo.EXPECT().
		GetPoll("poll456").
		Return(closedPoll, nil).
		Times(1)

	mockRepo.EXPECT().
		GetPoll("notfound").
		Return(nil, model.ErrPollNotFound).
		Times(1)

	mockRepo.EXPECT().
		AddVote(gomock.Any()).
		DoAndReturn(func(vote *model.Vote) error {
			if vote.PollID != "poll123" || vote.UserID != "user789" || vote.OptionIdx != 1 {
				return errors.New("unexpected vote parameters")
			}
			return nil
		}).Times(1)

	mockRepo.EXPECT().
		AddVote(gomock.Any()).
		Return(model.ErrAlreadyVoted).
		Times(1)

	type fields struct {
		repo       Repository
		pollConfig config.PollConfig
	}
	type args struct {
		pollID    string
		userID    string
		optionIdx int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Valid vote",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID:    "poll123",
				userID:    "user789",
				optionIdx: 1,
			},
			wantErr: false,
		},
		{
			name: "Vote on closed poll",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID:    "poll456",
				userID:    "user789",
				optionIdx: 0,
			},
			wantErr: true,
		},
		{
			name: "Vote on non-existent poll",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID:    "notfound",
				userID:    "user789",
				optionIdx: 0,
			},
			wantErr: true,
		},
		{
			name: "Vote with invalid option index",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID:    "poll123",
				userID:    "user789",
				optionIdx: 5,
			},
			wantErr: true,
		},
		{
			name: "Already voted",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID:    "poll123",
				userID:    "existing",
				optionIdx: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PollService{
				repo:       tt.fields.repo,
				pollConfig: tt.fields.pollConfig,
			}
			if err := s.Vote(tt.args.pollID, tt.args.userID, tt.args.optionIdx); (err != nil) != tt.wantErr {
				t.Errorf("Vote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPollService_CalculateResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pollConfig := config.PollConfig{
		DefaultDuration: 3600,
		MaxOptions:      10,
	}

	now := time.Now().Unix()
	future := now + 3600

	poll := &model.Poll{
		ID:        "poll123",
		Question:  "Test Poll",
		Options:   []string{"Option 1", "Option 2", "Option 3"},
		CreatedBy: "user123",
		ChannelID: "channel456",
		CreatedAt: now,
		ExpiresAt: future,
		Status:    model.PollStatusActive,
	}

	votes := []*model.Vote{
		{ID: "vote1", PollID: "poll123", UserID: "user1", OptionIdx: 0},
		{ID: "vote2", PollID: "poll123", UserID: "user2", OptionIdx: 0},
		{ID: "vote3", PollID: "poll123", UserID: "user3", OptionIdx: 1},
		{ID: "vote4", PollID: "poll123", UserID: "user4", OptionIdx: 2},
		{ID: "vote5", PollID: "poll123", UserID: "user5", OptionIdx: 0},
	}

	mockRepo.EXPECT().
		GetVotesByPollID("poll123").
		Return(votes, nil).
		Times(1)

	mockRepo.EXPECT().
		GetVotesByPollID("empty").
		Return([]*model.Vote{}, nil).
		Times(1)

	mockRepo.EXPECT().
		GetVotesByPollID("error").
		Return(nil, errors.New("database error")).
		Times(1)

	type fields struct {
		repo       Repository
		pollConfig config.PollConfig
	}
	type args struct {
		poll *model.Poll
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *VoteResults
		wantErr bool
	}{
		{
			name: "Calculate results with votes",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				poll: poll,
			},
			want: &VoteResults{
				PollID:     "poll123",
				Question:   "Test Poll",
				TotalVotes: 5,
				Results: []VoteCountResult{
					{OptionIndex: 0, OptionText: "Option 1", Count: 3},
					{OptionIndex: 1, OptionText: "Option 2", Count: 1},
					{OptionIndex: 2, OptionText: "Option 3", Count: 1},
				},
				IsActive: true,
			},
			wantErr: false,
		},
		{
			name: "Calculate results with no votes",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				poll: &model.Poll{
					ID:        "empty",
					Question:  "Empty Poll",
					Options:   []string{"Option 1", "Option 2"},
					CreatedBy: "user123",
					ChannelID: "channel456",
					CreatedAt: now,
					ExpiresAt: future,
					Status:    model.PollStatusActive,
				},
			},
			want: &VoteResults{
				PollID:     "empty",
				Question:   "Empty Poll",
				TotalVotes: 0,
				Results: []VoteCountResult{
					{OptionIndex: 0, OptionText: "Option 1", Count: 0},
					{OptionIndex: 1, OptionText: "Option 2", Count: 0},
				},
				IsActive: true,
			},
			wantErr: false,
		},
		{
			name: "Error getting votes",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				poll: &model.Poll{
					ID:        "error",
					Question:  "Error Poll",
					Options:   []string{"Option 1", "Option 2"},
					CreatedBy: "user123",
					ChannelID: "channel456",
					CreatedAt: now,
					ExpiresAt: future,
					Status:    model.PollStatusActive,
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PollService{
				repo:       tt.fields.repo,
				pollConfig: tt.fields.pollConfig,
			}
			got, err := s.CalculateResults(tt.args.poll)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.PollID != tt.want.PollID {
				t.Errorf("CalculateResults().PollID = %v, want %v", got.PollID, tt.want.PollID)
			}
			if got.Question != tt.want.Question {
				t.Errorf("CalculateResults().Question = %v, want %v", got.Question, tt.want.Question)
			}
			if got.TotalVotes != tt.want.TotalVotes {
				t.Errorf("CalculateResults().TotalVotes = %v, want %v", got.TotalVotes, tt.want.TotalVotes)
			}
			if got.IsActive != tt.want.IsActive {
				t.Errorf("CalculateResults().IsActive = %v, want %v", got.IsActive, tt.want.IsActive)
			}

			if len(got.Results) != len(tt.want.Results) {
				t.Errorf("CalculateResults().Results length = %v, want %v", len(got.Results), len(tt.want.Results))
				return
			}

			for i, result := range got.Results {
				if result.OptionIndex != tt.want.Results[i].OptionIndex {
					t.Errorf("Results[%d].OptionIndex = %v, want %v", i, result.OptionIndex, tt.want.Results[i].OptionIndex)
				}
				if result.OptionText != tt.want.Results[i].OptionText {
					t.Errorf("Results[%d].OptionText = %v, want %v", i, result.OptionText, tt.want.Results[i].OptionText)
				}
				if result.Count != tt.want.Results[i].Count {
					t.Errorf("Results[%d].Count = %v, want %v", i, result.Count, tt.want.Results[i].Count)
				}
			}
		})
	}
}

func TestPollService_GetResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pollConfig := config.PollConfig{
		DefaultDuration: 3600,
		MaxOptions:      10,
	}

	now := time.Now().Unix()
	future := now + 3600

	activePoll := &model.Poll{
		ID:        "poll123",
		Question:  "Active Poll",
		Options:   []string{"Option 1", "Option 2", "Option 3"},
		CreatedBy: "user123",
		ChannelID: "channel456",
		CreatedAt: now,
		ExpiresAt: future,
		Status:    model.PollStatusActive,
	}

	votes := []*model.Vote{
		{ID: "vote1", PollID: "poll123", UserID: "user1", OptionIdx: 0},
		{ID: "vote2", PollID: "poll123", UserID: "user2", OptionIdx: 0},
		{ID: "vote3", PollID: "poll123", UserID: "user3", OptionIdx: 1},
	}

	mockRepo.EXPECT().
		GetPoll("poll123").
		Return(activePoll, nil).
		Times(1)

	mockRepo.EXPECT().
		GetPoll("notfound").
		Return(nil, model.ErrPollNotFound).
		Times(1)

	mockRepo.EXPECT().
		GetVotesByPollID("poll123").
		Return(votes, nil).
		Times(1)

	type fields struct {
		repo       Repository
		pollConfig config.PollConfig
	}
	type args struct {
		pollID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *VoteResults
		wantErr bool
	}{
		{
			name: "Get results for existing poll",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID: "poll123",
			},
			want: &VoteResults{
				PollID:     "poll123",
				Question:   "Active Poll",
				TotalVotes: 3,
				Results: []VoteCountResult{
					{OptionIndex: 0, OptionText: "Option 1", Count: 2},
					{OptionIndex: 1, OptionText: "Option 2", Count: 1},
					{OptionIndex: 2, OptionText: "Option 3", Count: 0},
				},
				IsActive: true,
			},
			wantErr: false,
		},
		{
			name: "Get results for non-existent poll",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID: "notfound",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PollService{
				repo:       tt.fields.repo,
				pollConfig: tt.fields.pollConfig,
			}
			got, err := s.GetResults(tt.args.pollID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.PollID != tt.want.PollID {
				t.Errorf("GetResults().PollID = %v, want %v", got.PollID, tt.want.PollID)
			}
			if got.Question != tt.want.Question {
				t.Errorf("GetResults().Question = %v, want %v", got.Question, tt.want.Question)
			}
			if got.TotalVotes != tt.want.TotalVotes {
				t.Errorf("GetResults().TotalVotes = %v, want %v", got.TotalVotes, tt.want.TotalVotes)
			}
			if got.IsActive != tt.want.IsActive {
				t.Errorf("GetResults().IsActive = %v, want %v", got.IsActive, tt.want.IsActive)
			}

			for i, result := range got.Results {
				if result.Count != tt.want.Results[i].Count {
					t.Errorf("Results[%d].Count = %v, want %v", i, result.Count, tt.want.Results[i].Count)
				}
			}
		})
	}
}

func TestPollService_EndPoll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	pollConfig := config.PollConfig{
		DefaultDuration: 3600,
		MaxOptions:      10,
	}

	now := time.Now().Unix()
	future := now + 3600

	activePoll := &model.Poll{
		ID:        "poll123",
		Question:  "Active Poll",
		Options:   []string{"Option 1", "Option 2", "Option 3"},
		CreatedBy: "user123",
		ChannelID: "channel456",
		CreatedAt: now,
		ExpiresAt: future,
		Status:    model.PollStatusActive,
	}

	activePollWithVotes := &model.Poll{
		ID:        "poll456",
		Question:  "Poll with votes",
		Options:   []string{"Option 1", "Option 2", "Option 3"},
		CreatedBy: "user123",
		ChannelID: "channel456",
		CreatedAt: now,
		ExpiresAt: future,
		Status:    model.PollStatusActive,
	}

	closedPoll := &model.Poll{
		ID:        "closedpoll",
		Question:  "Closed Poll",
		Options:   []string{"Option 1", "Option 2"},
		CreatedBy: "user123",
		ChannelID: "channel456",
		CreatedAt: now,
		ExpiresAt: future,
		Status:    model.PollStatusClosed,
	}

	votes := []*model.Vote{
		{ID: "vote1", PollID: "poll456", UserID: "user1", OptionIdx: 0},
		{ID: "vote2", PollID: "poll456", UserID: "user2", OptionIdx: 0},
		{ID: "vote3", PollID: "poll456", UserID: "user3", OptionIdx: 1},
	}

	mockRepo.EXPECT().
		GetPoll("poll123").
		Return(activePoll, nil).
		Times(1)

	mockRepo.EXPECT().
		GetPoll("poll456").
		Return(activePollWithVotes, nil).
		Times(1)

	mockRepo.EXPECT().
		GetPoll("closedpoll").
		Return(closedPoll, nil).
		Times(1)

	mockRepo.EXPECT().
		GetPoll("notfound").
		Return(nil, model.ErrPollNotFound).
		Times(1)

	mockRepo.EXPECT().
		UpdatePollStatus("poll123", model.PollStatusClosed).
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		UpdatePollStatus("poll456", model.PollStatusClosed).
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		GetVotesByPollID("poll123").
		Return([]*model.Vote{}, nil).
		Times(1)

	mockRepo.EXPECT().
		GetVotesByPollID("poll456").
		Return(votes, nil).
		Times(1)

	type fields struct {
		repo       Repository
		pollConfig config.PollConfig
	}
	type args struct {
		pollID string
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *VoteResults
		wantErr bool
	}{
		{
			name: "End active poll by creator",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID: "poll123",
				userID: "user123",
			},
			want: &VoteResults{
				PollID:     "poll123",
				Question:   "Active Poll",
				TotalVotes: 0,
				Results: []VoteCountResult{
					{OptionIndex: 0, OptionText: "Option 1", Count: 0},
					{OptionIndex: 1, OptionText: "Option 2", Count: 0},
					{OptionIndex: 2, OptionText: "Option 3", Count: 0},
				},
				IsActive: false,
			},
			wantErr: false,
		},
		{
			name: "End active poll with votes by creator",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID: "poll456",
				userID: "user123",
			},
			want: &VoteResults{
				PollID:     "poll456",
				Question:   "Poll with votes",
				TotalVotes: 3,
				Results: []VoteCountResult{
					{OptionIndex: 0, OptionText: "Option 1", Count: 2},
					{OptionIndex: 1, OptionText: "Option 2", Count: 1},
					{OptionIndex: 2, OptionText: "Option 3", Count: 0},
				},
				IsActive: false,
			},
			wantErr: false,
		},
		{
			name: "End already closed poll",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID: "closedpoll",
				userID: "user123",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "End non-existent poll",
			fields: fields{
				repo:       mockRepo,
				pollConfig: pollConfig,
			},
			args: args{
				pollID: "notfound",
				userID: "user123",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PollService{
				repo:       tt.fields.repo,
				pollConfig: tt.fields.pollConfig,
			}

			got, err := s.EndPoll(tt.args.pollID, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("EndPoll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if got != nil {
					t.Errorf("EndPoll() got = %v, want nil", got)
				}
				return
			}

			if got.PollID != tt.want.PollID {
				t.Errorf("EndPoll().PollID = %v, want %v", got.PollID, tt.want.PollID)
			}
			if got.Question != tt.want.Question {
				t.Errorf("EndPoll().Question = %v, want %v", got.Question, tt.want.Question)
			}
			if got.TotalVotes != tt.want.TotalVotes {
				t.Errorf("EndPoll().TotalVotes = %v, want %v", got.TotalVotes, tt.want.TotalVotes)
			}
			if got.IsActive != tt.want.IsActive {
				t.Errorf("EndPoll().IsActive = %v, want %v", got.IsActive, tt.want.IsActive)
			}

			if len(got.Results) != len(tt.want.Results) {
				t.Errorf("len(EndPoll().Results) = %v, want %v",
					len(got.Results), len(tt.want.Results))
				return
			}

			for i, result := range got.Results {
				if result.Count != tt.want.Results[i].Count {
					t.Errorf("Results[%d].Count = %v, want %v",
						i, result.Count, tt.want.Results[i].Count)
				}
			}
		})
	}
}

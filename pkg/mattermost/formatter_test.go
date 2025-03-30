package mattermost

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
	"vk-test-assignment-mattermost-polls/internal/api/dto"
	"vk-test-assignment-mattermost-polls/internal/model"
	"vk-test-assignment-mattermost-polls/internal/service"
)

func TestFormatError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want *dto.MattermostResponse
	}{
		{
			name: "Simple error message",
			args: args{
				err: errors.New("test error"),
			},
			want: &dto.MattermostResponse{
				ResponseType: "ephemeral",
				Text:         "Error: test error",
			},
		},
		{
			name: "Model error",
			args: args{
				err: model.ErrPollClosed,
			},
			want: &dto.MattermostResponse{
				ResponseType: "ephemeral",
				Text:         "Error: poll is already closed",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatError(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormatError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatHelp(t *testing.T) {
	tests := []struct {
		name string
		want *dto.MattermostResponse
	}{
		{
			name: "Help message",
			want: &dto.MattermostResponse{
				ResponseType: "ephemeral",
				Text:         GetHelpText(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatHelp(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormatHelp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatPollCreated(t *testing.T) {
	now := time.Now()
	future := now.Add(2 * time.Hour)

	type args struct {
		poll *model.Poll
	}
	tests := []struct {
		name string
		args args
		want *dto.MattermostResponse
	}{
		{
			name: "New poll created",
			args: args{
				poll: &model.Poll{
					ID:        "poll123",
					Question:  "What's your favorite language?",
					Options:   []string{"Go", "Rust", "Python"},
					CreatedBy: "user123",
					ChannelID: "channel456",
					CreatedAt: now.Unix(),
					ExpiresAt: future.Unix(),
					Status:    model.PollStatusActive,
				},
			},
			want: &dto.MattermostResponse{
				ResponseType: "in_channel",
				Text: "### What's your favorite language?\n\n" +
					"**Poll ID:** poll123\n\n" +
					"1. Go\n" +
					"2. Rust\n" +
					"3. Python\n\n" +
					"**How to vote:**\n" +
					"Use `/poll vote poll123 NUMBER` to vote\n\n" +
					"**Expires in:** 2 hours 0 minutes\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPollCreated(tt.args.poll)

			if got.ResponseType != tt.want.ResponseType {
				t.Errorf("FormatPollCreated() ResponseType = %v, want %v", got.ResponseType, tt.want.ResponseType)
			}

			checkTextContains(t, got.Text, []string{
				tt.args.poll.Question,
				tt.args.poll.ID,
				"How to vote",
				"Expires in",
			})

			for _, option := range tt.args.poll.Options {
				checkTextContains(t, got.Text, []string{option})
			}
		})
	}
}

func TestFormatPollDeleted(t *testing.T) {
	type args struct {
		pollID string
	}
	tests := []struct {
		name string
		args args
		want *dto.MattermostResponse
	}{
		{
			name: "Poll deleted message",
			args: args{
				pollID: "poll123",
			},
			want: &dto.MattermostResponse{
				ResponseType: "ephemeral",
				Text:         "Poll with ID `poll123` has been deleted.",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatPollDeleted(tt.args.pollID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormatPollDeleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatPollEnded(t *testing.T) {
	type args struct {
		results *service.VoteResults
	}
	tests := []struct {
		name string
		args args
		want *dto.MattermostResponse
	}{
		{
			name: "Poll ended with results",
			args: args{
				results: &service.VoteResults{
					PollID:     "poll123",
					Question:   "What's your favorite language?",
					TotalVotes: 5,
					Results: []service.VoteCountResult{
						{OptionIndex: 0, OptionText: "Go", Count: 3},
						{OptionIndex: 1, OptionText: "Rust", Count: 1},
						{OptionIndex: 2, OptionText: "Python", Count: 1},
					},
					IsActive: false,
				},
			},
			want: &dto.MattermostResponse{
				ResponseType: "in_channel",
			},
		},
		{
			name: "Poll ended with tie",
			args: args{
				results: &service.VoteResults{
					PollID:     "poll123",
					Question:   "What's your favorite language?",
					TotalVotes: 4,
					Results: []service.VoteCountResult{
						{OptionIndex: 0, OptionText: "Go", Count: 2},
						{OptionIndex: 1, OptionText: "Rust", Count: 2},
						{OptionIndex: 2, OptionText: "Python", Count: 0},
					},
					IsActive: false,
				},
			},
			want: &dto.MattermostResponse{
				ResponseType: "in_channel",
			},
		},
		{
			name: "Poll ended with no votes",
			args: args{
				results: &service.VoteResults{
					PollID:     "poll123",
					Question:   "What's your favorite language?",
					TotalVotes: 0,
					Results: []service.VoteCountResult{
						{OptionIndex: 0, OptionText: "Go", Count: 0},
						{OptionIndex: 1, OptionText: "Rust", Count: 0},
						{OptionIndex: 2, OptionText: "Python", Count: 0},
					},
					IsActive: false,
				},
			},
			want: &dto.MattermostResponse{
				ResponseType: "in_channel",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPollEnded(tt.args.results)

			if got.ResponseType != tt.want.ResponseType {
				t.Errorf("FormatPollEnded() ResponseType = %v, want %v", got.ResponseType, tt.want.ResponseType)
			}

			checkTextContains(t, got.Text, []string{
				"Poll Ended",
				tt.args.results.Question,
				tt.args.results.PollID,
				"Total votes",
			})

			switch tt.name {
			case "Poll ended with results":
				checkTextContains(t, got.Text, []string{"Winner", "Go", "3 votes"})
			case "Poll ended with tie":
				checkTextContains(t, got.Text, []string{"Tie between", "Go", "Rust"})
			case "Poll ended with no votes":

			}

			for _, result := range tt.args.results.Results {
				checkTextContains(t, got.Text, []string{result.OptionText})
			}
		})
	}
}

func TestFormatPollInfo(t *testing.T) {
	now := time.Now()
	future := now.Add(2 * time.Hour)

	type args struct {
		poll *model.Poll
	}
	tests := []struct {
		name string
		args args
		want *dto.MattermostResponse
	}{
		{
			name: "Active poll info",
			args: args{
				poll: &model.Poll{
					ID:        "poll123",
					Question:  "What's your favorite language?",
					Options:   []string{"Go", "Rust", "Python"},
					CreatedBy: "user123",
					ChannelID: "channel456",
					CreatedAt: now.Unix(),
					ExpiresAt: future.Unix(),
					Status:    model.PollStatusActive,
				},
			},
			want: &dto.MattermostResponse{
				ResponseType: "ephemeral",
			},
		},
		{
			name: "Closed poll info",
			args: args{
				poll: &model.Poll{
					ID:        "poll123",
					Question:  "What's your favorite language?",
					Options:   []string{"Go", "Rust", "Python"},
					CreatedBy: "user123",
					ChannelID: "channel456",
					CreatedAt: now.Unix(),
					ExpiresAt: future.Unix(),
					Status:    model.PollStatusClosed,
				},
			},
			want: &dto.MattermostResponse{
				ResponseType: "ephemeral",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPollInfo(tt.args.poll)

			if got.ResponseType != tt.want.ResponseType {
				t.Errorf("FormatPollInfo() ResponseType = %v, want %v", got.ResponseType, tt.want.ResponseType)
			}

			expectedContent := []string{
				"Poll Information",
				tt.args.poll.Question,
				tt.args.poll.ID,
				"Status",
				string(tt.args.poll.Status),
				"Created by",
				tt.args.poll.CreatedBy,
				"Created at",
			}

			if tt.args.poll.Status == model.PollStatusActive {
				expectedContent = append(expectedContent, "Expires at", "Remaining time")
			} else {
				expectedContent = append(expectedContent, "Expired at")
			}

			checkTextContains(t, got.Text, expectedContent)

			for _, option := range tt.args.poll.Options {
				checkTextContains(t, got.Text, []string{option})
			}
		})
	}
}

func TestFormatPollResults(t *testing.T) {
	type args struct {
		results   *service.VoteResults
		ephemeral bool
	}
	tests := []struct {
		name string
		args args
		want *dto.MattermostResponse
	}{
		{
			name: "Active poll results (ephemeral)",
			args: args{
				results: &service.VoteResults{
					PollID:     "poll123",
					Question:   "What's your favorite language?",
					TotalVotes: 5,
					Results: []service.VoteCountResult{
						{OptionIndex: 0, OptionText: "Go", Count: 3},
						{OptionIndex: 1, OptionText: "Rust", Count: 1},
						{OptionIndex: 2, OptionText: "Python", Count: 1},
					},
					IsActive:      true,
					RemainingTime: "2 hours 30 minutes",
				},
				ephemeral: true,
			},
			want: &dto.MattermostResponse{
				ResponseType: "ephemeral",
			},
		},
		{
			name: "Active poll results (in channel)",
			args: args{
				results: &service.VoteResults{
					PollID:     "poll123",
					Question:   "What's your favorite language?",
					TotalVotes: 5,
					Results: []service.VoteCountResult{
						{OptionIndex: 0, OptionText: "Go", Count: 3},
						{OptionIndex: 1, OptionText: "Rust", Count: 1},
						{OptionIndex: 2, OptionText: "Python", Count: 1},
					},
					IsActive:      true,
					RemainingTime: "2 hours 30 minutes",
				},
				ephemeral: false,
			},
			want: &dto.MattermostResponse{
				ResponseType: "in_channel",
			},
		},
		{
			name: "Closed poll results",
			args: args{
				results: &service.VoteResults{
					PollID:     "poll123",
					Question:   "What's your favorite language?",
					TotalVotes: 5,
					Results: []service.VoteCountResult{
						{OptionIndex: 0, OptionText: "Go", Count: 3},
						{OptionIndex: 1, OptionText: "Rust", Count: 1},
						{OptionIndex: 2, OptionText: "Python", Count: 1},
					},
					IsActive: false,
				},
				ephemeral: false,
			},
			want: &dto.MattermostResponse{
				ResponseType: "in_channel",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPollResults(tt.args.results, tt.args.ephemeral)

			if got.ResponseType != tt.want.ResponseType {
				t.Errorf("FormatPollResults() ResponseType = %v, want %v", got.ResponseType, tt.want.ResponseType)
			}

			expectedContent := []string{
				"Results",
				tt.args.results.Question,
				tt.args.results.PollID,
				"Total votes",
			}

			if tt.args.results.IsActive {
				expectedContent = append(expectedContent,
					"**Status:** Active",
					"Remaining time",
					tt.args.results.RemainingTime,
					"To vote",
				)
			} else {
				expectedContent = append(expectedContent, "**Status:** Closed")
			}

			checkTextContains(t, got.Text, expectedContent)

			for _, result := range tt.args.results.Results {
				if tt.args.results.TotalVotes > 0 {
				}
				checkTextContains(t, got.Text, []string{
					result.OptionText,
				})
			}
		})
	}
}

func TestFormatVoteConfirmed(t *testing.T) {
	type args struct {
		poll      *model.Poll
		optionIdx int
	}
	tests := []struct {
		name string
		args args
		want *dto.MattermostResponse
	}{
		{
			name: "Vote confirmation message",
			args: args{
				poll: &model.Poll{
					ID:      "poll123",
					Options: []string{"Go", "Rust", "Python"},
				},
				optionIdx: 0,
			},
			want: &dto.MattermostResponse{
				ResponseType: "ephemeral",
				Text:         "Your vote for option 1: \"Go\" has been recorded.",
			},
		},
		{
			name: "Vote confirmation for second option",
			args: args{
				poll: &model.Poll{
					ID:      "poll123",
					Options: []string{"Go", "Rust", "Python"},
				},
				optionIdx: 1,
			},
			want: &dto.MattermostResponse{
				ResponseType: "ephemeral",
				Text:         "Your vote for option 2: \"Rust\" has been recorded.",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatVoteConfirmed(tt.args.poll, tt.args.optionIdx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormatVoteConfirmed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func checkTextContains(t *testing.T, text string, expectedContents []string) {
	for _, expected := range expectedContents {
		if !contains(text, expected) {
			t.Errorf("Expected text to contain %q, but it does not.\nText: %s", expected, text)
		}
	}
}

func contains(text, substr string) bool {
	return strings.Contains(text, substr)
}

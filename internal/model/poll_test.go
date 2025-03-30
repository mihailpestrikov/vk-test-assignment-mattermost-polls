package model

import (
	"reflect"
	"testing"
	"time"
)

func TestNewPoll(t *testing.T) {
	type args struct {
		question   string
		options    []string
		createdBy  string
		channelID  string
		duration   int
		maxOptions int
	}
	tests := []struct {
		name    string
		args    args
		want    *Poll
		wantErr bool
	}{
		{
			name: "Valid poll creation",
			args: args{
				question:   "Test Question",
				options:    []string{"Option 1", "Option 2", "Option 3"},
				createdBy:  "user123",
				channelID:  "channel456",
				duration:   3600,
				maxOptions: 10,
			},
			want: &Poll{
				Question:  "Test Question",
				Options:   []string{"Option 1", "Option 2", "Option 3"},
				CreatedBy: "user123",
				ChannelID: "channel456",
				Status:    PollStatusActive,
			},
			wantErr: false,
		},
		{
			name: "Empty question",
			args: args{
				question:   "",
				options:    []string{"Option 1", "Option 2"},
				createdBy:  "user123",
				channelID:  "channel456",
				duration:   3600,
				maxOptions: 10,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Too few options",
			args: args{
				question:   "Test Question",
				options:    []string{"Option 1"},
				createdBy:  "user123",
				channelID:  "channel456",
				duration:   3600,
				maxOptions: 10,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Too many options",
			args: args{
				question:   "Test Question",
				options:    []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
				createdBy:  "user123",
				channelID:  "channel456",
				duration:   3600,
				maxOptions: 10,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Duplicate options",
			args: args{
				question:   "Test Question",
				options:    []string{"Option 1", "Option 1", "Option 3"},
				createdBy:  "user123",
				channelID:  "channel456",
				duration:   3600,
				maxOptions: 10,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPoll(tt.args.question, tt.args.options, tt.args.createdBy, tt.args.channelID, tt.args.duration, tt.args.maxOptions)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPoll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got.Question != tt.want.Question {
				t.Errorf("NewPoll().Question = %v, want %v", got.Question, tt.want.Question)
			}

			if !reflect.DeepEqual(got.Options, tt.want.Options) {
				t.Errorf("NewPoll().Options = %v, want %v", got.Options, tt.want.Options)
			}

			if got.CreatedBy != tt.want.CreatedBy {
				t.Errorf("NewPoll().CreatedBy = %v, want %v", got.CreatedBy, tt.want.CreatedBy)
			}

			if got.ChannelID != tt.want.ChannelID {
				t.Errorf("NewPoll().ChannelID = %v, want %v", got.ChannelID, tt.want.ChannelID)
			}

			if got.Status != tt.want.Status {
				t.Errorf("NewPoll().Status = %v, want %v", got.Status, tt.want.Status)
			}

			if got.ExpiresAt != got.CreatedAt+int64(tt.args.duration) {
				t.Errorf("NewPoll().ExpiresAt = %v, want %v", got.ExpiresAt, got.CreatedAt+int64(tt.args.duration))
			}
		})
	}
}

func TestPollFromTarantoolTuple(t *testing.T) {
	type args struct {
		tuple []interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *Poll
		wantErr bool
	}{
		{
			name: "Valid tuple conversion",
			args: args{
				tuple: []interface{}{
					"poll123",
					"Test Question",
					[]interface{}{"Option 1", "Option 2", "Option 3"},
					"user123",
					"channel456",
					int64(1648234567),
					int64(1648238167),
					"ACTIVE",
				},
			},
			want: &Poll{
				ID:        "poll123",
				Question:  "Test Question",
				Options:   []string{"Option 1", "Option 2", "Option 3"},
				CreatedBy: "user123",
				ChannelID: "channel456",
				CreatedAt: 1648234567,
				ExpiresAt: 1648238167,
				Status:    PollStatusActive,
			},
			wantErr: false,
		},
		{
			name: "Insufficient tuple data",
			args: args{
				tuple: []interface{}{
					"poll123",
					"Test Question",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PollFromTarantoolTuple(tt.args.tuple)
			if (err != nil) != tt.wantErr {
				t.Errorf("PollFromTarantoolTuple() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PollFromTarantoolTuple() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoll_CanBeManipulatedBy(t *testing.T) {
	type fields struct {
		ID        string
		Question  string
		Options   []string
		CreatedBy string
		ChannelID string
		CreatedAt int64
		ExpiresAt int64
		Status    PollStatus
	}
	type args struct {
		userID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Can be manipulated by creator",
			fields: fields{
				ID:        "poll123",
				CreatedBy: "user123",
				Status:    PollStatusActive,
			},
			args: args{
				userID: "user123",
			},
			want: true,
		},
		{
			name: "Cannot be manipulated by other user",
			fields: fields{
				ID:        "poll123",
				CreatedBy: "user123",
				Status:    PollStatusActive,
			},
			args: args{
				userID: "user456",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poll{
				ID:        tt.fields.ID,
				Question:  tt.fields.Question,
				Options:   tt.fields.Options,
				CreatedBy: tt.fields.CreatedBy,
				ChannelID: tt.fields.ChannelID,
				CreatedAt: tt.fields.CreatedAt,
				ExpiresAt: tt.fields.ExpiresAt,
				Status:    tt.fields.Status,
			}
			if got := p.CanBeManipulatedBy(tt.args.userID); got != tt.want {
				t.Errorf("CanBeManipulatedBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoll_Close(t *testing.T) {
	type fields struct {
		ID        string
		Question  string
		Options   []string
		CreatedBy string
		ChannelID string
		CreatedAt int64
		ExpiresAt int64
		Status    PollStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   PollStatus
	}{
		{
			name: "Close active poll",
			fields: fields{
				ID:     "poll123",
				Status: PollStatusActive,
			},
			want: PollStatusClosed,
		},
		{
			name: "Close closed poll",
			fields: fields{
				ID:     "poll123",
				Status: PollStatusClosed,
			},
			want: PollStatusClosed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poll{
				ID:        tt.fields.ID,
				Question:  tt.fields.Question,
				Options:   tt.fields.Options,
				CreatedBy: tt.fields.CreatedBy,
				ChannelID: tt.fields.ChannelID,
				CreatedAt: tt.fields.CreatedAt,
				ExpiresAt: tt.fields.ExpiresAt,
				Status:    tt.fields.Status,
			}
			p.Close()
			if p.Status != tt.want {
				t.Errorf("After Close() Status = %v, want %v", p.Status, tt.want)
			}
		})
	}
}

func TestPoll_Delete(t *testing.T) {
	type fields struct {
		ID        string
		Question  string
		Options   []string
		CreatedBy string
		ChannelID string
		CreatedAt int64
		ExpiresAt int64
		Status    PollStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   PollStatus
	}{
		{
			name: "Delete active poll",
			fields: fields{
				ID:     "poll123",
				Status: PollStatusActive,
			},
			want: PollStatusDeleted,
		},
		{
			name: "Delete closed poll",
			fields: fields{
				ID:     "poll123",
				Status: PollStatusClosed,
			},
			want: PollStatusDeleted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poll{
				ID:        tt.fields.ID,
				Question:  tt.fields.Question,
				Options:   tt.fields.Options,
				CreatedBy: tt.fields.CreatedBy,
				ChannelID: tt.fields.ChannelID,
				CreatedAt: tt.fields.CreatedAt,
				ExpiresAt: tt.fields.ExpiresAt,
				Status:    tt.fields.Status,
			}
			p.Delete()
			if p.Status != tt.want {
				t.Errorf("After Delete() Status = %v, want %v", p.Status, tt.want)
			}
		})
	}
}

func TestPoll_GetExpirationTime(t *testing.T) {
	timestamp := int64(1648234567)
	loc := time.Local
	want := time.Unix(timestamp, 0).In(loc).Format("2006-01-02 15:04:05")

	type fields struct {
		ID        string
		Question  string
		Options   []string
		CreatedBy string
		ChannelID string
		CreatedAt int64
		ExpiresAt int64
		Status    PollStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Format expiration time",
			fields: fields{
				ID:        "poll123",
				ExpiresAt: timestamp,
			},
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poll{
				ID:        tt.fields.ID,
				Question:  tt.fields.Question,
				Options:   tt.fields.Options,
				CreatedBy: tt.fields.CreatedBy,
				ChannelID: tt.fields.ChannelID,
				CreatedAt: tt.fields.CreatedAt,
				ExpiresAt: tt.fields.ExpiresAt,
				Status:    tt.fields.Status,
			}
			if got := p.GetExpirationTime(); got != tt.want {
				t.Errorf("GetExpirationTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoll_GetFormattedCreationTime(t *testing.T) {
	timestamp := int64(1648234567)
	loc := time.Local
	want := time.Unix(timestamp, 0).In(loc).Format("2006-01-02 15:04:05")

	type fields struct {
		ID        string
		Question  string
		Options   []string
		CreatedBy string
		ChannelID string
		CreatedAt int64
		ExpiresAt int64
		Status    PollStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Format creation time",
			fields: fields{
				ID:        "poll123",
				CreatedAt: timestamp,
			},
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poll{
				ID:        tt.fields.ID,
				Question:  tt.fields.Question,
				Options:   tt.fields.Options,
				CreatedBy: tt.fields.CreatedBy,
				ChannelID: tt.fields.ChannelID,
				CreatedAt: tt.fields.CreatedAt,
				ExpiresAt: tt.fields.ExpiresAt,
				Status:    tt.fields.Status,
			}
			if got := p.GetFormattedCreationTime(); got != tt.want {
				t.Errorf("GetFormattedCreationTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoll_GetRemainingTime(t *testing.T) {
	now := time.Now()
	oneHourLater := now.Add(1 * time.Hour).Unix()
	twoHoursLater := now.Add(2 * time.Hour).Unix()

	pastTime := now.Add(-1 * time.Hour).Unix()

	type fields struct {
		ID        string
		Question  string
		Options   []string
		CreatedBy string
		ChannelID string
		CreatedAt int64
		ExpiresAt int64
		Status    PollStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "One hour remaining",
			fields: fields{
				ID:        "poll123",
				ExpiresAt: oneHourLater,
				Status:    PollStatusActive,
			},
			want: "1 hours 0 minutes",
		},
		{
			name: "Two hours remaining",
			fields: fields{
				ID:        "poll123",
				ExpiresAt: twoHoursLater,
				Status:    PollStatusActive,
			},
			want: "2 hours 0 minutes",
		},
		{
			name: "Expired poll",
			fields: fields{
				ID:        "poll123",
				ExpiresAt: pastTime,
				Status:    PollStatusActive,
			},
			want: "Time has expired",
		},
		{
			name: "Closed poll",
			fields: fields{
				ID:        "poll123",
				ExpiresAt: oneHourLater,
				Status:    PollStatusClosed,
			},
			want: "Poll has ended",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poll{
				ID:        tt.fields.ID,
				Question:  tt.fields.Question,
				Options:   tt.fields.Options,
				CreatedBy: tt.fields.CreatedBy,
				ChannelID: tt.fields.ChannelID,
				CreatedAt: tt.fields.CreatedAt,
				ExpiresAt: tt.fields.ExpiresAt,
				Status:    tt.fields.Status,
			}
			got := p.GetRemainingTime()

			if tt.fields.Status != PollStatusActive || p.ExpiresAt < now.Unix() {
				if got != tt.want {
					t.Errorf("GetRemainingTime() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestPoll_HasExpired(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(1 * time.Hour).Unix()
	pastTime := now.Add(-1 * time.Hour).Unix()

	type fields struct {
		ID        string
		Question  string
		Options   []string
		CreatedBy string
		ChannelID string
		CreatedAt int64
		ExpiresAt int64
		Status    PollStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Not expired poll",
			fields: fields{
				ID:        "poll123",
				ExpiresAt: futureTime,
			},
			want: false,
		},
		{
			name: "Expired poll",
			fields: fields{
				ID:        "poll123",
				ExpiresAt: pastTime,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poll{
				ID:        tt.fields.ID,
				Question:  tt.fields.Question,
				Options:   tt.fields.Options,
				CreatedBy: tt.fields.CreatedBy,
				ChannelID: tt.fields.ChannelID,
				CreatedAt: tt.fields.CreatedAt,
				ExpiresAt: tt.fields.ExpiresAt,
				Status:    tt.fields.Status,
			}
			if got := p.HasExpired(); got != tt.want {
				t.Errorf("HasExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoll_IsActive(t *testing.T) {
	type fields struct {
		ID        string
		Question  string
		Options   []string
		CreatedBy string
		ChannelID string
		CreatedAt int64
		ExpiresAt int64
		Status    PollStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Active poll",
			fields: fields{
				ID:     "poll123",
				Status: PollStatusActive,
			},
			want: true,
		},
		{
			name: "Closed poll",
			fields: fields{
				ID:     "poll123",
				Status: PollStatusClosed,
			},
			want: false,
		},
		{
			name: "Deleted poll",
			fields: fields{
				ID:     "poll123",
				Status: PollStatusDeleted,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poll{
				ID:        tt.fields.ID,
				Question:  tt.fields.Question,
				Options:   tt.fields.Options,
				CreatedBy: tt.fields.CreatedBy,
				ChannelID: tt.fields.ChannelID,
				CreatedAt: tt.fields.CreatedAt,
				ExpiresAt: tt.fields.ExpiresAt,
				Status:    tt.fields.Status,
			}
			if got := p.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoll_IsValidOptionIndex(t *testing.T) {
	type fields struct {
		ID        string
		Question  string
		Options   []string
		CreatedBy string
		ChannelID string
		CreatedAt int64
		ExpiresAt int64
		Status    PollStatus
	}
	type args struct {
		index int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Valid index",
			fields: fields{
				ID:      "poll123",
				Options: []string{"Option 1", "Option 2", "Option 3"},
			},
			args: args{
				index: 1,
			},
			want: true,
		},
		{
			name: "Negative index",
			fields: fields{
				ID:      "poll123",
				Options: []string{"Option 1", "Option 2", "Option 3"},
			},
			args: args{
				index: -1,
			},
			want: false,
		},
		{
			name: "Index too large",
			fields: fields{
				ID:      "poll123",
				Options: []string{"Option 1", "Option 2", "Option 3"},
			},
			args: args{
				index: 3,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poll{
				ID:        tt.fields.ID,
				Question:  tt.fields.Question,
				Options:   tt.fields.Options,
				CreatedBy: tt.fields.CreatedBy,
				ChannelID: tt.fields.ChannelID,
				CreatedAt: tt.fields.CreatedAt,
				ExpiresAt: tt.fields.ExpiresAt,
				Status:    tt.fields.Status,
			}
			if got := p.IsValidOptionIndex(tt.args.index); got != tt.want {
				t.Errorf("IsValidOptionIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoll_ToTarantoolTuple(t *testing.T) {
	type fields struct {
		ID        string
		Question  string
		Options   []string
		CreatedBy string
		ChannelID string
		CreatedAt int64
		ExpiresAt int64
		Status    PollStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   []interface{}
	}{
		{
			name: "Convert to Tarantool tuple",
			fields: fields{
				ID:        "poll123",
				Question:  "Test Question",
				Options:   []string{"Option 1", "Option 2", "Option 3"},
				CreatedBy: "user123",
				ChannelID: "channel456",
				CreatedAt: 1648234567,
				ExpiresAt: 1648238167,
				Status:    PollStatusActive,
			},
			want: []interface{}{
				"poll123",
				"Test Question",
				[]string{"Option 1", "Option 2", "Option 3"},
				"user123",
				"channel456",
				int64(1648234567),
				int64(1648238167),
				"ACTIVE",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poll{
				ID:        tt.fields.ID,
				Question:  tt.fields.Question,
				Options:   tt.fields.Options,
				CreatedBy: tt.fields.CreatedBy,
				ChannelID: tt.fields.ChannelID,
				CreatedAt: tt.fields.CreatedAt,
				ExpiresAt: tt.fields.ExpiresAt,
				Status:    tt.fields.Status,
			}
			got := p.ToTarantoolTuple()

			// Проверяем длину
			if len(got) != len(tt.want) {
				t.Errorf("ToTarantoolTuple() length = %v, want %v", len(got), len(tt.want))
				return
			}

			// Проверяем каждый элемент
			for i := 0; i < len(got); i++ {
				if !reflect.DeepEqual(got[i], tt.want[i]) {
					t.Errorf("ToTarantoolTuple()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

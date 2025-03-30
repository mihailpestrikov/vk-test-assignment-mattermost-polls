package model

import (
	"reflect"
	"testing"
)

func TestNewVote(t *testing.T) {
	type args struct {
		pollID    string
		userID    string
		optionIdx int
	}
	tests := []struct {
		name string
		args args
		want *Vote
	}{
		{
			name: "Create new vote",
			args: args{
				pollID:    "poll123",
				userID:    "user456",
				optionIdx: 2,
			},
			want: &Vote{
				PollID:    "poll123",
				UserID:    "user456",
				OptionIdx: 2,
			},
		},
		{
			name: "Create vote with negative index",
			args: args{
				pollID:    "poll123",
				userID:    "user456",
				optionIdx: -1,
			},
			want: &Vote{
				PollID:    "poll123",
				UserID:    "user456",
				OptionIdx: -1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewVote(tt.args.pollID, tt.args.userID, tt.args.optionIdx)

			if got.PollID != tt.want.PollID {
				t.Errorf("NewVote().PollID = %v, want %v", got.PollID, tt.want.PollID)
			}

			if got.UserID != tt.want.UserID {
				t.Errorf("NewVote().UserID = %v, want %v", got.UserID, tt.want.UserID)
			}

			if got.OptionIdx != tt.want.OptionIdx {
				t.Errorf("NewVote().OptionIdx = %v, want %v", got.OptionIdx, tt.want.OptionIdx)
			}

			if got.ID == "" {
				t.Errorf("NewVote().ID should not be empty")
			}
		})
	}
}

func TestVoteFromTarantoolTuple(t *testing.T) {
	type args struct {
		tuple []interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *Vote
		wantErr bool
	}{
		{
			name: "Valid tuple conversion",
			args: args{
				tuple: []interface{}{
					"vote123",
					"poll123",
					"user456",
					int64(2),
					int64(1648234567),
				},
			},
			want: &Vote{
				ID:        "vote123",
				PollID:    "poll123",
				UserID:    "user456",
				OptionIdx: 2,
				CreatedAt: 1648234567,
			},
			wantErr: false,
		},
		{
			name: "Tuple with int option index",
			args: args{
				tuple: []interface{}{
					"vote123",
					"poll123",
					"user456",
					2,
					int64(1648234567),
				},
			},
			want: &Vote{
				ID:        "vote123",
				PollID:    "poll123",
				UserID:    "user456",
				OptionIdx: 2,
				CreatedAt: 1648234567,
			},
			wantErr: false,
		},
		{
			name: "Tuple with float option index",
			args: args{
				tuple: []interface{}{
					"vote123",
					"poll123",
					"user456",
					float64(2),
					int64(1648234567),
				},
			},
			want: &Vote{
				ID:        "vote123",
				PollID:    "poll123",
				UserID:    "user456",
				OptionIdx: 2,
				CreatedAt: 1648234567,
			},
			wantErr: false,
		},
		{
			name: "Insufficient tuple data",
			args: args{
				tuple: []interface{}{
					"vote123",
					"poll123",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid option index type",
			args: args{
				tuple: []interface{}{
					"vote123",
					"poll123",
					"user456",
					"not_an_int",
					int64(1648234567),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VoteFromTarantoolTuple(tt.args.tuple)
			if (err != nil) != tt.wantErr {
				t.Errorf("VoteFromTarantoolTuple() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VoteFromTarantoolTuple() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVote_ToTarantoolTuple(t *testing.T) {
	type fields struct {
		ID        string
		PollID    string
		UserID    string
		OptionIdx int
		CreatedAt int64
	}
	tests := []struct {
		name   string
		fields fields
		want   []interface{}
	}{
		{
			name: "Convert to Tarantool tuple",
			fields: fields{
				ID:        "vote123",
				PollID:    "poll123",
				UserID:    "user456",
				OptionIdx: 2,
				CreatedAt: 1648234567,
			},
			want: []interface{}{
				"vote123",
				"poll123",
				"user456",
				2,
				int64(1648234567),
			},
		},
		{
			name: "Convert with negative option index",
			fields: fields{
				ID:        "vote123",
				PollID:    "poll123",
				UserID:    "user456",
				OptionIdx: -1,
				CreatedAt: 1648234567,
			},
			want: []interface{}{
				"vote123",
				"poll123",
				"user456",
				-1,
				int64(1648234567),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Vote{
				ID:        tt.fields.ID,
				PollID:    tt.fields.PollID,
				UserID:    tt.fields.UserID,
				OptionIdx: tt.fields.OptionIdx,
				CreatedAt: tt.fields.CreatedAt,
			}
			got := v.ToTarantoolTuple()

			if len(got) != len(tt.want) {
				t.Errorf("ToTarantoolTuple() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for i := 0; i < len(got); i++ {
				if !reflect.DeepEqual(got[i], tt.want[i]) {
					t.Errorf("ToTarantoolTuple()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

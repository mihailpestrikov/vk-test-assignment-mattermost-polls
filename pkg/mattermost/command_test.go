package mattermost

import (
	"reflect"
	"testing"
)

func TestGetHelpText(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Help text contains essential commands",
			want: `Available commands:

/poll create "Question" "Option 1" "Option 2" [--duration=86400]
    Create a new poll with specified options and optional duration in seconds

/poll vote POLL_ID OPTION_NUMBER
    Vote for an option in the specified poll

/poll results POLL_ID
    Show current results of the poll

/poll end POLL_ID
    End the poll and show final results (only creator can end)

/poll delete POLL_ID
    Delete the poll (only creator can delete)

/poll info POLL_ID
    Show detailed information about the poll`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHelpText(); got != tt.want {
				t.Errorf("GetHelpText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name    string
		args    args
		want    *Command
		wantErr bool
	}{
		{
			name: "Empty command defaults to help",
			args: args{
				text: "",
			},
			want: &Command{
				SubCommand: CommandHelp,
			},
			wantErr: false,
		},
		{
			name: "Help command",
			args: args{
				text: "help",
			},
			want: &Command{
				SubCommand: CommandHelp,
			},
			wantErr: false,
		},
		{
			name: "Create command with two options",
			args: args{
				text: `create "What's your favorite language?" "Go" "Rust"`,
			},
			want: &Command{
				SubCommand: CommandCreate,
				Question:   "What's your favorite language?",
				Options:    []string{"Go", "Rust"},
				Duration:   0,
			},
			wantErr: false,
		},
		{
			name: "Create command with duration",
			args: args{
				text: `create "What's your favorite language?" "Go" "Rust" --duration=3600`,
			},
			want: &Command{
				SubCommand: CommandCreate,
				Question:   "What's your favorite language?",
				Options:    []string{"Go", "Rust"},
				Duration:   3600,
			},
			wantErr: false,
		},
		{
			name: "Create with invalid duration",
			args: args{
				text: `create "What's your favorite language?" "Go" "Rust" --duration=invalid`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Create with empty question",
			args: args{
				text: `create "" "Go" "Rust"`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Create with only one option",
			args: args{
				text: `create "What's your favorite language?" "Go"`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Create with duplicate options",
			args: args{
				text: `create "What's your favorite language?" "Go" "Go"`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Vote command",
			args: args{
				text: "vote poll123 1",
			},
			want: &Command{
				SubCommand: CommandVote,
				PollID:     "poll123",
				OptionIdx:  0,
			},
			wantErr: false,
		},
		{
			name: "Vote command with missing poll ID",
			args: args{
				text: "vote",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Vote command with missing option index",
			args: args{
				text: "vote poll123",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Vote command with invalid option index",
			args: args{
				text: "vote poll123 invalid",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Vote command with negative option index",
			args: args{
				text: "vote poll123 -1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Results command",
			args: args{
				text: "results poll123",
			},
			want: &Command{
				SubCommand: CommandResults,
				PollID:     "poll123",
			},
			wantErr: false,
		},
		{
			name: "Results command with missing poll ID",
			args: args{
				text: "results",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "End command",
			args: args{
				text: "end poll123",
			},
			want: &Command{
				SubCommand: CommandEnd,
				PollID:     "poll123",
			},
			wantErr: false,
		},
		{
			name: "End command with missing poll ID",
			args: args{
				text: "end",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Delete command",
			args: args{
				text: "delete poll123",
			},
			want: &Command{
				SubCommand: CommandDelete,
				PollID:     "poll123",
			},
			wantErr: false,
		},
		{
			name: "Delete command with missing poll ID",
			args: args{
				text: "delete",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Info command",
			args: args{
				text: "info poll123",
			},
			want: &Command{
				SubCommand: CommandInfo,
				PollID:     "poll123",
			},
			wantErr: false,
		},
		{
			name: "Info command with missing poll ID",
			args: args{
				text: "info",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Unknown command",
			args: args{
				text: "unknown",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommand(tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.SubCommand != tt.want.SubCommand {
				t.Errorf("ParseCommand() got SubCommand = %v, want %v", got.SubCommand, tt.want.SubCommand)
			}
			switch got.SubCommand {
			case CommandCreate:
				if got.Question != tt.want.Question {
					t.Errorf("ParseCommand() got Question = %v, want %v", got.Question, tt.want.Question)
				}
				if !reflect.DeepEqual(got.Options, tt.want.Options) {
					t.Errorf("ParseCommand() got Options = %v, want %v", got.Options, tt.want.Options)
				}
				if got.Duration != tt.want.Duration {
					t.Errorf("ParseCommand() got Duration = %v, want %v", got.Duration, tt.want.Duration)
				}
			case CommandVote:
				if got.PollID != tt.want.PollID {
					t.Errorf("ParseCommand() got PollID = %v, want %v", got.PollID, tt.want.PollID)
				}
				if got.OptionIdx != tt.want.OptionIdx {
					t.Errorf("ParseCommand() got OptionIdx = %v, want %v", got.OptionIdx, tt.want.OptionIdx)
				}
			case CommandResults, CommandEnd, CommandDelete, CommandInfo:
				if got.PollID != tt.want.PollID {
					t.Errorf("ParseCommand() got PollID = %v, want %v", got.PollID, tt.want.PollID)
				}
			}
		})
	}
}

func Test_parseCreateCommand(t *testing.T) {
	type args struct {
		args    []string
		command *Command
	}
	tests := []struct {
		name    string
		args    args
		want    *Command
		wantErr bool
	}{
		{
			name: "Valid create command with two options",
			args: args{
				args:    []string{"create", "Test Question", "Option 1", "Option 2"},
				command: &Command{SubCommand: CommandCreate},
			},
			want: &Command{
				SubCommand: CommandCreate,
				Question:   "Test Question",
				Options:    []string{"Option 1", "Option 2"},
				Duration:   0,
			},
			wantErr: false,
		},
		{
			name: "Valid create command with duration",
			args: args{
				args:    []string{"create", "Test Question", "Option 1", "Option 2", "--duration=3600"},
				command: &Command{SubCommand: CommandCreate},
			},
			want: &Command{
				SubCommand: CommandCreate,
				Question:   "Test Question",
				Options:    []string{"Option 1", "Option 2"},
				Duration:   3600,
			},
			wantErr: false,
		},
		{
			name: "Create with too few arguments",
			args: args{
				args:    []string{"create", "Test Question"},
				command: &Command{SubCommand: CommandCreate},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Create with empty question",
			args: args{
				args:    []string{"create", "", "Option 1", "Option 2"},
				command: &Command{SubCommand: CommandCreate},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Create with duplicate options",
			args: args{
				args:    []string{"create", "Test Question", "Option 1", "Option 1"},
				command: &Command{SubCommand: CommandCreate},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Create with invalid duration format",
			args: args{
				args:    []string{"create", "Test Question", "Option 1", "Option 2", "--duration:3600"},
				command: &Command{SubCommand: CommandCreate},
			},
			want: &Command{
				SubCommand: CommandCreate,
				Question:   "Test Question",
				Options:    []string{"Option 1", "Option 2", "--duration:3600"},
				Duration:   0,
			},
			wantErr: false,
		},
		{
			name: "Create with invalid duration value",
			args: args{
				args:    []string{"create", "Test Question", "Option 1", "Option 2", "--duration=invalid"},
				command: &Command{SubCommand: CommandCreate},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Create with duration but too few options",
			args: args{
				args:    []string{"create", "Test Question", "Option 1", "--duration=3600"},
				command: &Command{SubCommand: CommandCreate},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCreateCommand(tt.args.args, tt.args.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCreateCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Question != tt.want.Question {
				t.Errorf("parseCreateCommand() got Question = %v, want %v", got.Question, tt.want.Question)
			}
			if !reflect.DeepEqual(got.Options, tt.want.Options) {
				t.Errorf("parseCreateCommand() got Options = %v, want %v", got.Options, tt.want.Options)
			}
			if got.Duration != tt.want.Duration {
				t.Errorf("parseCreateCommand() got Duration = %v, want %v", got.Duration, tt.want.Duration)
			}
		})
	}
}

func Test_parseSimpleCommand(t *testing.T) {
	type args struct {
		args    []string
		command *Command
	}
	tests := []struct {
		name    string
		args    args
		want    *Command
		wantErr bool
	}{
		{
			name: "Valid results command",
			args: args{
				args:    []string{"results", "poll123"},
				command: &Command{SubCommand: CommandResults},
			},
			want: &Command{
				SubCommand: CommandResults,
				PollID:     "poll123",
			},
			wantErr: false,
		},
		{
			name: "Valid end command",
			args: args{
				args:    []string{"end", "poll123"},
				command: &Command{SubCommand: CommandEnd},
			},
			want: &Command{
				SubCommand: CommandEnd,
				PollID:     "poll123",
			},
			wantErr: false,
		},
		{
			name: "Valid delete command",
			args: args{
				args:    []string{"delete", "poll123"},
				command: &Command{SubCommand: CommandDelete},
			},
			want: &Command{
				SubCommand: CommandDelete,
				PollID:     "poll123",
			},
			wantErr: false,
		},
		{
			name: "Valid info command",
			args: args{
				args:    []string{"info", "poll123"},
				command: &Command{SubCommand: CommandInfo},
			},
			want: &Command{
				SubCommand: CommandInfo,
				PollID:     "poll123",
			},
			wantErr: false,
		},
		{
			name: "Command with missing poll ID",
			args: args{
				args:    []string{"results"},
				command: &Command{SubCommand: CommandResults},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Command with extra arguments (ignored)",
			args: args{
				args:    []string{"results", "poll123", "extra"},
				command: &Command{SubCommand: CommandResults},
			},
			want: &Command{
				SubCommand: CommandResults,
				PollID:     "poll123",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSimpleCommand(tt.args.args, tt.args.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSimpleCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.SubCommand != tt.want.SubCommand {
				t.Errorf("parseSimpleCommand() got SubCommand = %v, want %v", got.SubCommand, tt.want.SubCommand)
			}
			if got.PollID != tt.want.PollID {
				t.Errorf("parseSimpleCommand() got PollID = %v, want %v", got.PollID, tt.want.PollID)
			}
		})
	}
}

func Test_parseVoteCommand(t *testing.T) {
	type args struct {
		args    []string
		command *Command
	}
	tests := []struct {
		name    string
		args    args
		want    *Command
		wantErr bool
	}{
		{
			name: "Valid vote command",
			args: args{
				args:    []string{"vote", "poll123", "1"},
				command: &Command{SubCommand: CommandVote},
			},
			want: &Command{
				SubCommand: CommandVote,
				PollID:     "poll123",
				OptionIdx:  0,
			},
			wantErr: false,
		},
		{
			name: "Vote with larger index",
			args: args{
				args:    []string{"vote", "poll123", "5"},
				command: &Command{SubCommand: CommandVote},
			},
			want: &Command{
				SubCommand: CommandVote,
				PollID:     "poll123",
				OptionIdx:  4,
			},
			wantErr: false,
		},
		{
			name: "Vote with missing poll ID",
			args: args{
				args:    []string{"vote"},
				command: &Command{SubCommand: CommandVote},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Vote with missing option index",
			args: args{
				args:    []string{"vote", "poll123"},
				command: &Command{SubCommand: CommandVote},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Vote with invalid option index",
			args: args{
				args:    []string{"vote", "poll123", "invalid"},
				command: &Command{SubCommand: CommandVote},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Vote with negative option index",
			args: args{
				args:    []string{"vote", "poll123", "0"},
				command: &Command{SubCommand: CommandVote},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Vote with negative option value",
			args: args{
				args:    []string{"vote", "poll123", "-1"},
				command: &Command{SubCommand: CommandVote},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Vote with extra arguments (ignored)",
			args: args{
				args:    []string{"vote", "poll123", "1", "extra"},
				command: &Command{SubCommand: CommandVote},
			},
			want: &Command{
				SubCommand: CommandVote,
				PollID:     "poll123",
				OptionIdx:  0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVoteCommand(tt.args.args, tt.args.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVoteCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.SubCommand != tt.want.SubCommand {
				t.Errorf("parseVoteCommand() got SubCommand = %v, want %v", got.SubCommand, tt.want.SubCommand)
			}
			if got.PollID != tt.want.PollID {
				t.Errorf("parseVoteCommand() got PollID = %v, want %v", got.PollID, tt.want.PollID)
			}
			if got.OptionIdx != tt.want.OptionIdx {
				t.Errorf("parseVoteCommand() got OptionIdx = %v, want %v", got.OptionIdx, tt.want.OptionIdx)
			}
		})
	}
}

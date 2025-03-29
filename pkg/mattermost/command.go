package mattermost

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mattn/go-shellwords"

	"vk-test-assignment-mattermost-polls/internal/model"
)

const (
	CommandCreate  = "create"
	CommandVote    = "vote"
	CommandResults = "results"
	CommandEnd     = "end"
	CommandDelete  = "delete"
	CommandInfo    = "info"
	CommandHelp    = "help"
)

var (
	ErrInvalidSubCommand  = errors.New("invalid subcommand")
	ErrMissingPollID      = errors.New("poll ID is required")
	ErrMissingOptionIndex = errors.New("option index is required")
	ErrInvalidDuration    = errors.New("invalid duration format, use --duration=SECONDS")
)

type Command struct {
	SubCommand string   // Тип команды (create, vote, results, etc.)
	PollID     string   // ID голосования
	OptionIdx  int      // Индекс выбранного варианта (для vote)
	Question   string   // Вопрос голосования (для create)
	Options    []string // Варианты ответов (для create)
	Duration   int      // Продолжительность голосования в секундах (для create)
}

func ParseCommand(text string) (*Command, error) {
	if text == "" {
		return &Command{SubCommand: CommandHelp}, nil
	}

	args, err := shellwords.Parse(text)
	if err != nil {
		return nil, fmt.Errorf("error parsing command: %w", err)
	}

	if len(args) == 0 {
		return &Command{SubCommand: CommandHelp}, nil
	}

	command := &Command{
		SubCommand: strings.ToLower(args[0]),
	}

	switch command.SubCommand {
	case CommandCreate:
		return parseCreateCommand(args, command)
	case CommandVote:
		return parseVoteCommand(args, command)
	case CommandResults, CommandEnd, CommandDelete, CommandInfo:
		return parseSimpleCommand(args, command)
	case CommandHelp, "":
		command.SubCommand = CommandHelp
		return command, nil
	default:
		return nil, ErrInvalidSubCommand
	}
}

// parseCreateCommand create "question" "variant1" "variant2 --duration=[int]"
func parseCreateCommand(args []string, command *Command) (*Command, error) {
	if len(args) < 3 {
		return nil, model.ErrTooFewOptions
	}

	command.Question = args[1]
	command.Options = args[2:]

	if command.Question == "" {
		return nil, model.ErrEmptyQuestion
	}

	for i, opt := range command.Options {
		if strings.HasPrefix(opt, "--duration=") {
			durationStr := strings.TrimPrefix(opt, "--duration=")

			duration, err := strconv.Atoi(durationStr)
			if err != nil {
				return nil, ErrInvalidDuration
			}

			command.Duration = duration

			command.Options = append(command.Options[:i], command.Options[i+1:]...)
			break
		}
	}

	if len(command.Options) < 2 {
		return nil, model.ErrTooFewOptions
	}

	optionMap := make(map[string]struct{}, len(command.Options))
	for _, opt := range command.Options {
		if _, exists := optionMap[opt]; exists {
			return nil, model.ErrDuplicateOption
		}
		optionMap[opt] = struct{}{}
	}

	return command, nil
}

// parseVoteCommand vote [poll_id] [option_index]
func parseVoteCommand(args []string, command *Command) (*Command, error) {
	if len(args) < 2 {
		return nil, ErrMissingPollID
	}

	command.PollID = args[1]

	if len(args) < 3 {
		return nil, ErrMissingOptionIndex
	}

	optionIdx, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, model.ErrInvalidOption
	}

	command.OptionIdx = optionIdx - 1

	if command.OptionIdx < 0 {
		return nil, model.ErrInvalidOption
	}

	return command, nil
}

// parseSimpleCommand разбивает простые команды, которым нужен только ID голосования
// [subcommand] [poll_id]
func parseSimpleCommand(args []string, command *Command) (*Command, error) {
	if len(args) < 2 {
		return nil, ErrMissingPollID
	}

	command.PollID = args[1]
	return command, nil
}

func GetHelpText() string {
	return `Available commands:

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
    Show detailed information about the poll`
}

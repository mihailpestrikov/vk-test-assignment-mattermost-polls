package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PollStatus string

const (
	PollStatusActive  PollStatus = "ACTIVE"
	PollStatusClosed  PollStatus = "CLOSED"
	PollStatusDeleted PollStatus = "DELETED"
)

var (
	ErrPollNotFound    = errors.New("poll not found")
	ErrPollClosed      = errors.New("poll is already closed")
	ErrInvalidOption   = errors.New("invalid option")
	ErrEmptyQuestion   = errors.New("question cannot be empty")
	ErrTooFewOptions   = errors.New("at least 2 options are required")
	ErrTooManyOptions  = errors.New("too many options")
	ErrNotPollCreator  = errors.New("only the poll creator can perform this action")
	ErrDuplicateOption = errors.New("duplicate options detected")
)

type Poll struct {
	ID        string     `json:"id"`
	Question  string     `json:"question"`
	Options   []string   `json:"options"`
	CreatedBy string     `json:"created_by"`
	ChannelID string     `json:"channel_id"`
	CreatedAt int64      `json:"created_at"`
	ExpiresAt int64      `json:"expires_at"`
	Status    PollStatus `json:"status"`
}

func NewPoll(question string, options []string, createdBy, channelID string, duration int, maxOptions int) (*Poll, error) {
	if question == "" {
		return nil, ErrEmptyQuestion
	}

	if len(options) < 2 {
		return nil, ErrTooFewOptions
	}

	if len(options) > maxOptions {
		return nil, fmt.Errorf("%w: maximum %d options allowed", ErrTooManyOptions, maxOptions)
	}

	optionMap := make(map[string]struct{}, len(options))
	for _, opt := range options {
		if _, exists := optionMap[opt]; exists {
			return nil, ErrDuplicateOption
		}
		optionMap[opt] = struct{}{}
	}

	now := time.Now().Unix()

	return &Poll{
		ID:        uuid.New().String(),
		Question:  question,
		Options:   options,
		CreatedBy: createdBy,
		ChannelID: channelID,
		CreatedAt: now,
		ExpiresAt: now + int64(duration),
		Status:    PollStatusActive,
	}, nil
}

func (p *Poll) IsActive() bool {
	return p.Status == PollStatusActive
}

func (p *Poll) HasExpired() bool {
	return time.Now().Unix() > p.ExpiresAt
}

func (p *Poll) Close() {
	p.Status = PollStatusClosed
}

func (p *Poll) Delete() {
	p.Status = PollStatusDeleted
}

func (p *Poll) CanBeManipulatedBy(userID string) bool {
	return p.CreatedBy == userID
}

func (p *Poll) IsValidOptionIndex(index int) bool {
	return index >= 0 && index < len(p.Options)
}

func (p *Poll) GetExpirationTime() string {
	return time.Unix(p.ExpiresAt, 0).Format("2006-01-02 15:04:05")
}

func (p *Poll) GetFormattedCreationTime() string {
	return time.Unix(p.CreatedAt, 0).Format("2006-01-02 15:04:05")
}

func (p *Poll) GetRemainingTime() string {
	if !p.IsActive() {
		return "Poll has ended"
	}

	now := time.Now().Unix()
	if now >= p.ExpiresAt {
		return "Time has expired"
	}

	remaining := p.ExpiresAt - now
	hours := remaining / 3600
	minutes := (remaining % 3600) / 60

	if hours > 0 {
		return fmt.Sprintf("%d hours %d minutes", hours, minutes)
	}
	return fmt.Sprintf("%d minutes", minutes)
}

func (p *Poll) ToTarantoolTuple() []interface{} {
	return []interface{}{
		p.ID,
		p.Question,
		p.Options,
		p.CreatedBy,
		p.ChannelID,
		p.CreatedAt,
		p.ExpiresAt,
		string(p.Status),
	}
}

func PollFromTarantoolTuple(tuple []interface{}) (*Poll, error) {
	if len(tuple) < 8 {
		return nil, errors.New("not enough data in tuple")
	}

	var options []string
	if optionsInterface, ok := tuple[2].([]interface{}); ok {
		options = make([]string, len(optionsInterface))
		for i, opt := range optionsInterface {
			if strOpt, ok := opt.(string); ok {
				options[i] = strOpt
			} else {
				options[i] = fmt.Sprintf("%v", opt)
			}
		}
	}

	return &Poll{
		ID:        tuple[0].(string),
		Question:  tuple[1].(string),
		Options:   options,
		CreatedBy: tuple[3].(string),
		ChannelID: tuple[4].(string),
		CreatedAt: tuple[5].(int64),
		ExpiresAt: tuple[6].(int64),
		Status:    PollStatus(tuple[7].(string)),
	}, nil
}

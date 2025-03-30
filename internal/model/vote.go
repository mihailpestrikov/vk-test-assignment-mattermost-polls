package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAlreadyVoted = errors.New("you have already voted in this poll")
	ErrVoteNotFound = errors.New("vote not found")
)

type Vote struct {
	ID        string `json:"id"`
	PollID    string `json:"poll_id"`
	UserID    string `json:"user_id"`
	OptionIdx int    `json:"option_idx"`
	CreatedAt int64  `json:"created_at"`
}

func NewVote(pollID, userID string, optionIdx int) *Vote {
	return &Vote{
		ID:        uuid.New().String(),
		PollID:    pollID,
		UserID:    userID,
		OptionIdx: optionIdx,
		CreatedAt: time.Now().Unix(),
	}
}

func (v *Vote) ToTarantoolTuple() []interface{} {
	return []interface{}{
		v.ID,
		v.PollID,
		v.UserID,
		v.OptionIdx,
		v.CreatedAt,
	}
}

func VoteFromTarantoolTuple(tuple []interface{}) (*Vote, error) {
	if len(tuple) < 5 {
		return nil, errors.New("not enough data in tuple")
	}

	var optionIdx int
	switch v := tuple[3].(type) {
	case int:
		optionIdx = v
	case int8:
		optionIdx = int(v)
	default:
		return nil, fmt.Errorf("unexpected option index type: %T", v)
	}

	return &Vote{
		ID:        tuple[0].(string),
		PollID:    tuple[1].(string),
		UserID:    tuple[2].(string),
		OptionIdx: int(optionIdx),
		CreatedAt: tuple[4].(int64),
	}, nil
}

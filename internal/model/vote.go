package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAlreadyVoted     = errors.New("you have already voted in this poll")
	ErrVoteNotFound     = errors.New("vote not found")
	ErrInvalidVoteID    = errors.New("invalid vote ID")
	ErrVoteCreateFailed = errors.New("failed to create vote")
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

	optionIdx, ok := tuple[3].(int64)
	if !ok {
		if idx, ok := tuple[3].(int); ok {
			optionIdx = int64(idx)
		} else if idx, ok := tuple[3].(float64); ok {
			optionIdx = int64(idx)
		} else {
			return nil, errors.New("invalid option index type")
		}
	}

	return &Vote{
		ID:        tuple[0].(string),
		PollID:    tuple[1].(string),
		UserID:    tuple[2].(string),
		OptionIdx: int(optionIdx),
		CreatedAt: tuple[4].(int64),
	}, nil
}

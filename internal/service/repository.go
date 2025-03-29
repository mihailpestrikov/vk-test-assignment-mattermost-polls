package service

import (
	"time"
	"vk-test-assignment-mattermost-polls/internal/model"
)

type PollReader interface {
	GetPoll(id string) (*model.Poll, error)
	GetPollsByChannel(channelID string) ([]*model.Poll, error)
	GetPollsByCreator(userID string) ([]*model.Poll, error)
	GetExpiredActivePolls() ([]*model.Poll, error)
}

type PollWriter interface {
	CreatePoll(poll *model.Poll) error
	UpdatePollStatus(id string, status model.PollStatus) error
	DeletePoll(id string) error
	PurgeDeletedPolls(olderThan time.Duration) error
}

type VoteReader interface {
	GetVote(pollID, userID string) (*model.Vote, error)
	GetVotesByPollID(pollID string) ([]*model.Vote, error)
}

type VoteWriter interface {
	AddVote(vote *model.Vote) error
}

type Repository interface {
	PollReader
	PollWriter
	VoteReader
	VoteWriter
	Close() error
}

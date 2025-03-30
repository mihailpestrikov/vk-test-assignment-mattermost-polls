package repository

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tarantool/go-tarantool/v2"

	"vk-test-assignment-mattermost-polls/internal/model"
	"vk-test-assignment-mattermost-polls/internal/service"
	"vk-test-assignment-mattermost-polls/pkg/config"
)

type TarantoolRepository struct {
	conn       *tarantool.Connection
	spacePolls string
	spaceVotes string
}

func NewTarantoolRepository(cfg config.TarantoolConfig) (service.Repository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	dialer := tarantool.NetDialer{
		Address: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		User:    "guest",
	}

	conn, err := tarantool.Connect(ctx, dialer, tarantool.Opts{
		Timeout:     5 * time.Second,
		Concurrency: 32,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Tarantool: %w", err)
	}
	if conn != nil {
		log.Info().Msg("Connected to Tarantool successfully")
	}

	return &TarantoolRepository{
		conn:       conn,
		spacePolls: cfg.SpacePolls,
		spaceVotes: cfg.SpaceVotes,
	}, nil
}

func (r *TarantoolRepository) CreatePoll(poll *model.Poll) error {
	resp, err := r.conn.Do(tarantool.NewInsertRequest(r.spacePolls).Tuple(poll.ToTarantoolTuple())).Get()
	if err != nil {
		return fmt.Errorf("error creating poll: %w", err)
	}

	log.Debug().
		Str("poll_id", poll.ID).
		Str("channel_id", poll.ChannelID).
		Interface("response", resp).
		Msg("Poll created successfully")

	return nil
}

func (r *TarantoolRepository) GetPoll(id string) (*model.Poll, error) {
	resp, err := r.conn.Do(tarantool.NewSelectRequest(r.spacePolls).
		Index("primary").
		Offset(0).
		Limit(1).
		Iterator(tarantool.IterEq).
		Key([]interface{}{id})).Get()
	if err != nil {
		return nil, fmt.Errorf("error getting poll: %w", err)
	}

	if len(resp) == 0 {
		return nil, model.ErrPollNotFound
	}

	poll, err := model.PollFromTarantoolTuple(resp[0].([]interface{}))
	if err != nil {
		return nil, fmt.Errorf("error converting poll data: %w", err)
	}

	log.Debug().
		Str("poll_id", poll.ID).
		Str("channel_id", poll.ChannelID).
		Interface("response", resp).
		Msg("Poll received successfully")

	return poll, nil
}

func (r *TarantoolRepository) UpdatePollStatus(id string, status model.PollStatus) error {
	poll, err := r.GetPoll(id)
	if err != nil {
		return err
	}

	const statusIndex = 7

	req := tarantool.NewUpdateRequest(r.spacePolls).
		Index("primary").
		Key([]interface{}{id}).
		Operations(tarantool.NewOperations().Assign(statusIndex, string(status)))

	resp, err := r.conn.Do(req).Get()
	if err != nil {
		return fmt.Errorf("error updating poll status: %w", err)
	}

	log.Debug().
		Str("poll_id", id).
		Str("old_status", string(poll.Status)).
		Str("new_status", string(status)).
		Interface("response", resp).
		Msg("Poll status updated")

	return nil
}

func (r *TarantoolRepository) DeletePoll(id string) error {
	return r.UpdatePollStatus(id, model.PollStatusDeleted)
}

func (r *TarantoolRepository) PurgeDeletedPolls(olderThan time.Duration) error {
	cutoffTime := time.Now().Add(-olderThan).Unix()

	resp, err := r.conn.Do(tarantool.NewSelectRequest(r.spacePolls).
		Index("status_expires").
		Offset(0).
		Limit(1000).
		Iterator(tarantool.IterEq).
		Key([]interface{}{string(model.PollStatusDeleted)})).
		Get()

	if err != nil {
		return fmt.Errorf("error getting deleted polls: %w", err)
	}

	var purgedCount int

	for _, tupleData := range resp {
		poll, err := model.PollFromTarantoolTuple(tupleData.([]interface{}))
		if err != nil {
			continue
		}

		if poll.CreatedAt <= cutoffTime {
			r.conn.Do(tarantool.NewDeleteRequest(r.spacePolls).
				Index("primary").
				Key([]interface{}{poll.ID}))

			r.conn.Do(tarantool.NewDeleteRequest(r.spaceVotes).
				Index("poll_id").
				Key([]interface{}{poll.ID}))

			purgedCount++
		}
	}

	log.Info().
		Int("purged_count", purgedCount).
		Dur("older_than", olderThan).
		Msg("Completed purging deleted polls")

	return nil
}

func (r *TarantoolRepository) GetPollsByChannel(channelID string) ([]*model.Poll, error) {
	resp, err := r.conn.Do(tarantool.NewSelectRequest(r.spacePolls).
		Index("channel").
		Offset(0).
		Limit(100).
		Iterator(tarantool.IterEq).
		Key([]interface{}{channelID})).
		Get()

	if err != nil {
		return nil, fmt.Errorf("error getting channel polls: %w", err)
	}

	var polls []*model.Poll
	for _, tuple := range resp {
		poll, err := model.PollFromTarantoolTuple(tuple.([]interface{}))
		if err != nil {
			log.Error().Err(err).Msg("Error converting poll data")
			continue
		}

		if poll.Status != model.PollStatusDeleted {
			polls = append(polls, poll)
		}
	}

	return polls, nil
}

func (r *TarantoolRepository) GetPollsByCreator(userID string) ([]*model.Poll, error) {
	resp, err := r.conn.Do(tarantool.NewSelectRequest(r.spacePolls).
		Index("creator").
		Offset(0).
		Limit(100).
		Iterator(tarantool.IterEq).
		Key([]interface{}{userID})).
		Get()

	if err != nil {
		return nil, fmt.Errorf("error getting user polls: %w", err)
	}

	var polls []*model.Poll
	for _, tuple := range resp {
		poll, err := model.PollFromTarantoolTuple(tuple.([]interface{}))
		if err != nil {
			log.Error().Err(err).Msg("Error converting voting data")
			continue
		}

		if poll.Status != model.PollStatusDeleted {
			polls = append(polls, poll)
		}
	}

	return polls, nil
}

func (r *TarantoolRepository) GetExpiredActivePolls() ([]*model.Poll, error) {
	now := time.Now().Unix()

	resp, err := r.conn.Do(tarantool.NewSelectRequest(r.spacePolls).
		Index("status_expires").
		Offset(0).
		Limit(100).
		Iterator(tarantool.IterLe).
		Key([]interface{}{string(model.PollStatusActive), now})).
		Get()

	if err != nil {
		return nil, fmt.Errorf("error getting expired votes: %w", err)
	}

	var polls []*model.Poll
	for _, tuple := range resp {
		poll, err := model.PollFromTarantoolTuple(tuple.([]interface{}))
		if err != nil {
			log.Error().Err(err).Msg("Error converting voting data")
			continue
		}
		polls = append(polls, poll)
	}

	return polls, nil
}

func (r *TarantoolRepository) AddVote(vote *model.Vote) error {
	poll, err := r.GetPoll(vote.PollID)
	if err != nil {
		return err
	}

	if poll.Status != model.PollStatusActive {
		return model.ErrPollClosed
	}

	if poll.HasExpired() {
		err = r.UpdatePollStatus(poll.ID, model.PollStatusClosed)
		if err != nil {
			log.Error().Err(err).Str("poll_id", poll.ID).Msg("Failed to close expired poll")
		}
		return model.ErrPollClosed
	}

	existingVote, err := r.GetVote(vote.PollID, vote.UserID)
	if err != nil && !errors.Is(err, model.ErrVoteNotFound) {
		return err
	}

	if existingVote != nil {
		return model.ErrAlreadyVoted
	}

	if !poll.IsValidOptionIndex(vote.OptionIdx) {
		return model.ErrInvalidOption
	}

	resp, err := r.conn.Do(tarantool.NewInsertRequest(r.spaceVotes).Tuple(vote.ToTarantoolTuple())).Get()
	if err != nil {
		return fmt.Errorf("error adding vote: %w", err)
	}

	log.Debug().
		Str("vote_id", vote.ID).
		Str("poll_id", vote.PollID).
		Str("user_id", vote.UserID).
		Interface("response", resp).
		Msg("Vote added successfully")

	return nil
}

func (r *TarantoolRepository) GetVote(pollID, userID string) (*model.Vote, error) {
	resp, err := r.conn.Do(tarantool.NewSelectRequest(r.spaceVotes).
		Index("user_poll").
		Offset(0).
		Limit(1).
		Iterator(tarantool.IterEq).
		Key([]interface{}{userID, pollID})).
		Get()

	if err != nil {
		return nil, fmt.Errorf("error receiving vote: %w", err)
	}

	if len(resp) == 0 {
		return nil, model.ErrVoteNotFound
	}

	vote, err := model.VoteFromTarantoolTuple(resp[0].([]interface{}))
	if err != nil {
		return nil, fmt.Errorf("error converting vote data: %w", err)
	}

	return vote, nil
}

func (r *TarantoolRepository) GetVotesByPollID(pollID string) ([]*model.Vote, error) {
	resp, err := r.conn.Do(tarantool.NewSelectRequest(r.spaceVotes).
		Index("poll_id").
		Offset(0).
		Limit(1000).
		Iterator(tarantool.IterEq).
		Key([]interface{}{pollID})).
		Get()

	if err != nil {
		return nil, fmt.Errorf("error receiving votes: %w", err)
	}

	var votes []*model.Vote
	for _, tuple := range resp {
		vote, err := model.VoteFromTarantoolTuple(tuple.([]interface{}))
		if err != nil {
			log.Error().Err(err).Msg("Error converting vote data")
			continue
		}
		votes = append(votes, vote)
	}

	return votes, nil
}

func (r *TarantoolRepository) Close() error {
	if r.conn != nil {
		err := r.conn.Close()
		if err != nil {
			return fmt.Errorf("error closing connection to Tarantool: %w", err)
		}
		log.Info().Msg("Connection to Tarantool closed")
	}
	return nil
}

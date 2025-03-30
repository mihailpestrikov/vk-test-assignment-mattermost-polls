package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"vk-test-assignment-mattermost-polls/internal/model"
	"vk-test-assignment-mattermost-polls/pkg/config"
)

type VoteCountResult struct {
	OptionIndex int    `json:"option_index"`
	OptionText  string `json:"option_text"`
	Count       int    `json:"count"`
}

type VoteResults struct {
	PollID        string            `json:"poll_id"`
	Question      string            `json:"question"`
	TotalVotes    int               `json:"total_votes"`
	Results       []VoteCountResult `json:"results"`
	IsActive      bool              `json:"is_active"`
	RemainingTime string            `json:"remaining_time,omitempty"`
}

type IPollService interface {
	CreatePoll(question string, options []string, createdBy, channelID string, duration int) (*model.Poll, error)
	GetPoll(id string) (*model.Poll, error)
	Vote(pollID, userID string, optionIdx int) error
	GetResults(pollID string) (*VoteResults, error)
	EndPoll(pollID, userID string) (*VoteResults, error)
	DeletePoll(pollID, userID string) error
}

type PollService struct {
	repo       Repository
	pollConfig config.PollConfig
}

func NewPollService(repo Repository, pollConfig config.PollConfig) *PollService {
	return &PollService{
		repo:       repo,
		pollConfig: pollConfig,
	}
}

func (s *PollService) CreatePoll(question string, options []string, createdBy, channelID string, duration int) (*model.Poll, error) {

	if duration <= 0 {
		duration = s.pollConfig.DefaultDuration
	}

	if len(options) > s.pollConfig.MaxOptions {
		return nil, fmt.Errorf("%w: maximum %d options", model.ErrTooManyOptions, s.pollConfig.MaxOptions)
	}

	poll, err := model.NewPoll(question, options, createdBy, channelID, duration, s.pollConfig.MaxOptions)
	if err != nil {
		return nil, err
	}

	err = s.repo.CreatePoll(poll)
	if err != nil {
		return nil, err
	}

	log.Info().
		Str("poll_id", poll.ID).
		Str("created_by", createdBy).
		Str("channel_id", channelID).
		Int("options_count", len(options)).
		Msg("New poll created")

	return poll, nil
}

func (s *PollService) GetPoll(id string) (*model.Poll, error) {
	poll, err := s.repo.GetPoll(id)
	if err != nil {
		return nil, err
	}

	if poll.IsActive() && poll.HasExpired() {
		err = s.repo.UpdatePollStatus(poll.ID, model.PollStatusClosed)
		if err != nil {
			log.Error().Err(err).Str("poll_id", poll.ID).Msg("Failed to close expired poll")
		} else {
			poll.Status = model.PollStatusClosed
		}
	}

	return poll, nil
}

func (s *PollService) Vote(pollID, userID string, optionIdx int) error {

	poll, err := s.GetPoll(pollID)
	if err != nil {
		return err
	}

	if !poll.IsActive() {
		return model.ErrPollClosed
	}

	if !poll.IsValidOptionIndex(optionIdx) {
		return model.ErrInvalidOption
	}

	vote := model.NewVote(pollID, userID, optionIdx)

	err = s.repo.AddVote(vote)
	if err != nil {
		if errors.Is(err, model.ErrAlreadyVoted) {
			return err
		}
		return fmt.Errorf("error adding vote: %w", err)
	}

	log.Info().
		Str("poll_id", pollID).
		Str("user_id", userID).
		Int("option_idx", optionIdx).
		Msg("User voted")

	return nil
}

func (s *PollService) CalculateResults(poll *model.Poll) (*VoteResults, error) {

	votes, err := s.repo.GetVotesByPollID(poll.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting votes: %w", err)
	}

	results := &VoteResults{
		PollID:     poll.ID,
		Question:   poll.Question,
		TotalVotes: len(votes),
		IsActive:   poll.IsActive(),
		Results:    make([]VoteCountResult, len(poll.Options)),
	}

	if poll.IsActive() {
		results.RemainingTime = poll.GetRemainingTime()
	}

	for i, opt := range poll.Options {
		results.Results[i] = VoteCountResult{
			OptionIndex: i,
			OptionText:  opt,
			Count:       0,
		}
	}

	for _, vote := range votes {
		if vote.OptionIdx >= 0 && vote.OptionIdx < len(results.Results) {
			results.Results[vote.OptionIdx].Count++
		}
	}

	return results, nil
}

func (s *PollService) GetResults(pollID string) (*VoteResults, error) {

	poll, err := s.GetPoll(pollID)
	if err != nil {
		return nil, err
	}

	results, err := s.CalculateResults(poll)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *PollService) EndPoll(pollID, userID string) (*VoteResults, error) {

	poll, err := s.GetPoll(pollID)
	if err != nil {
		return nil, err
	}

	if !poll.IsActive() {
		return nil, model.ErrPollClosed
	}

	if !poll.CanBeManipulatedBy(userID) {
		return nil, model.ErrNotPollCreator
	}

	err = s.repo.UpdatePollStatus(pollID, model.PollStatusClosed)
	if err != nil {
		return nil, fmt.Errorf("error closing poll: %w", err)
	}

	poll.Status = model.PollStatusClosed
	results, err := s.CalculateResults(poll)
	if err != nil {
		return nil, err
	}

	log.Info().
		Str("poll_id", pollID).
		Str("user_id", userID).
		Msg("Poll closed")

	return results, nil
}

func (s *PollService) DeletePoll(pollID, userID string) error {

	poll, err := s.GetPoll(pollID)
	if err != nil {
		return err
	}

	if !poll.CanBeManipulatedBy(userID) {
		return model.ErrNotPollCreator
	}

	err = s.repo.UpdatePollStatus(pollID, model.PollStatusDeleted)
	if err != nil {
		return fmt.Errorf("error deleting poll: %w", err)
	}

	log.Info().
		Str("poll_id", pollID).
		Str("user_id", userID).
		Msg("Poll deleted")

	return nil
}

func (s *PollService) FinishExpiredPolls() error {

	expiredPolls, err := s.repo.GetExpiredActivePolls()
	if err != nil {
		return fmt.Errorf("error getting expired polls: %w", err)
	}

	if len(expiredPolls) == 0 {
		log.Debug().Msg("No expired polls to close")
		return nil
	}

	for _, poll := range expiredPolls {
		err := s.repo.UpdatePollStatus(poll.ID, model.PollStatusClosed)
		if err != nil {
			log.Error().
				Err(err).
				Str("poll_id", poll.ID).
				Msg("Error closing expired poll")
			continue
		}

		log.Info().
			Str("poll_id", poll.ID).
			Str("channel_id", poll.ChannelID).
			Msg("Automatically closed expired poll")

	}

	log.Info().
		Int("count", len(expiredPolls)).
		Msg("All expired polls closed")

	return nil
}

func (s *PollService) StartPollWatcher(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.FinishExpiredPolls(); err != nil {
					log.Error().
						Err(err).
						Msg("Error closing expired polls")
				}
			case <-ctx.Done():
				log.Info().Msg("Poll watcher stopped")
				return
			}
		}
	}()

	log.Info().Msg("Poll watcher started")
}

// StartPollCleaner Очистка голосований, помеченных как удаленные, старще 30 дней
func (s *PollService) StartPollCleaner(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.repo.PurgeDeletedPolls(30 * 24 * time.Hour); err != nil {
					log.Error().Err(err).Msg("Error purging deleted polls")
				}
			case <-ctx.Done():
				log.Info().Msg("Poll cleaner stopped")
				return
			}
		}
	}()
}

func (s *PollService) Close() error {
	if s.repo != nil {
		return s.repo.Close()
	}
	return nil
}

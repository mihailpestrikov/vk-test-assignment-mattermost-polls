package mattermost

import (
	"fmt"
	"strings"

	"vk-test-assignment-mattermost-polls/internal/api/dto"
	"vk-test-assignment-mattermost-polls/internal/model"
	"vk-test-assignment-mattermost-polls/internal/service"
)

func FormatError(err error) *dto.MattermostResponse {
	return &dto.MattermostResponse{
		ResponseType: dto.ResponseTypeEphemeral,
		Text:         fmt.Sprintf("Error: %s", err.Error()),
	}
}

func FormatPollCreated(poll *model.Poll) *dto.MattermostResponse {
	var sb strings.Builder

	sb.WriteString("### " + poll.Question + "\n\n")
	sb.WriteString("**Poll ID:** " + poll.ID + "\n\n")

	for i, option := range poll.Options {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, option))
	}

	sb.WriteString("\n**How to vote:**\n")
	sb.WriteString("Use `/poll vote " + poll.ID + " NUMBER` to vote\n\n")
	sb.WriteString("**Expires in:** " + poll.GetRemainingTime() + "\n")

	return &dto.MattermostResponse{
		ResponseType: dto.ResponseTypeInChannel,
		Text:         sb.String(),
	}
}

func FormatVoteConfirmed(poll *model.Poll, optionIdx int) *dto.MattermostResponse {
	return &dto.MattermostResponse{
		ResponseType: dto.ResponseTypeEphemeral,
		Text:         fmt.Sprintf("Your vote for option %d: \"%s\" has been recorded.", optionIdx+1, poll.Options[optionIdx]),
	}
}

func FormatPollResults(results *service.VoteResults, ephemeral bool) *dto.MattermostResponse {
	var sb strings.Builder

	responseType := dto.ResponseTypeInChannel
	if ephemeral {
		responseType = dto.ResponseTypeEphemeral
	}

	sb.WriteString("### Results: " + results.Question + "\n\n")
	sb.WriteString(fmt.Sprintf("**Poll ID:** %s\n", results.PollID))
	sb.WriteString(fmt.Sprintf("**Total votes:** %d\n\n", results.TotalVotes))

	if results.IsActive {
		sb.WriteString(fmt.Sprintf("**Status:** Active (Remaining time: %s)\n\n", results.RemainingTime))
	} else {
		sb.WriteString("**Status:** Closed\n\n")
	}

	for _, result := range results.Results {
		var percentage int
		if results.TotalVotes > 0 {
			percentage = (result.Count * 100) / results.TotalVotes
		}

		sb.WriteString(fmt.Sprintf("%d. **%s** - **%d votes** (%d%%)\n\n",
			result.OptionIndex+1,
			result.OptionText,
			result.Count,
			percentage))
	}

	if results.IsActive {
		sb.WriteString("**To vote:** `/poll vote " + results.PollID + " NUMBER`")
	}

	return &dto.MattermostResponse{
		ResponseType: responseType,
		Text:         sb.String(),
	}
}

func FormatPollEnded(results *service.VoteResults) *dto.MattermostResponse {
	var sb strings.Builder

	sb.WriteString("### Poll Ended: " + results.Question + "\n\n")
	sb.WriteString(fmt.Sprintf("**Poll ID:** %s\n", results.PollID))
	sb.WriteString(fmt.Sprintf("**Total votes:** %d\n\n", results.TotalVotes))

	var maxVotes int
	var winners []string

	for _, result := range results.Results {
		if result.Count > maxVotes {
			maxVotes = result.Count
			winners = []string{result.OptionText}
		} else if result.Count == maxVotes && maxVotes > 0 {
			winners = append(winners, result.OptionText)
		}
	}

	if len(winners) > 0 && maxVotes > 0 {
		if len(winners) == 1 {
			sb.WriteString(fmt.Sprintf("**Winner:** %s with %d votes\n\n", winners[0], maxVotes))
		} else {
			sb.WriteString(fmt.Sprintf("**Tie between:** %s with %d votes each\n\n", strings.Join(winners, ", "), maxVotes))
		}
	}

	for _, result := range results.Results {
		var percentage int
		if results.TotalVotes > 0 {
			percentage = (result.Count * 100) / results.TotalVotes
		}

		sb.WriteString(fmt.Sprintf("%d. **%s** - **%d votes** (%d%%)\n\n",
			result.OptionIndex+1,
			result.OptionText,
			result.Count,
			percentage))
	}

	return &dto.MattermostResponse{
		ResponseType: dto.ResponseTypeInChannel,
		Text:         sb.String(),
	}
}

func FormatPollDeleted(pollID string) *dto.MattermostResponse {
	return &dto.MattermostResponse{
		ResponseType: dto.ResponseTypeEphemeral,
		Text:         fmt.Sprintf("Poll with ID `%s` has been deleted.", pollID),
	}
}

func FormatPollInfo(poll *model.Poll) *dto.MattermostResponse {
	var sb strings.Builder

	sb.WriteString("### Poll Information\n\n")
	sb.WriteString(fmt.Sprintf("**Question:** %s\n\n", poll.Question))
	sb.WriteString(fmt.Sprintf("**Poll ID:** %s\n", poll.ID))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", poll.Status))
	sb.WriteString(fmt.Sprintf("**Created by:** %s\n", poll.CreatedBy))
	sb.WriteString(fmt.Sprintf("**Created at:** %s\n", poll.GetFormattedCreationTime()))

	if poll.IsActive() {
		sb.WriteString(fmt.Sprintf("**Expires at:** %s\n", poll.GetExpirationTime()))
		sb.WriteString(fmt.Sprintf("**Remaining time:** %s\n\n", poll.GetRemainingTime()))
	} else {
		sb.WriteString(fmt.Sprintf("**Expired at:** %s\n\n", poll.GetExpirationTime()))
	}

	sb.WriteString("**Options:**\n")
	for i, option := range poll.Options {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, option))
	}

	return &dto.MattermostResponse{
		ResponseType: dto.ResponseTypeEphemeral,
		Text:         sb.String(),
	}
}

func FormatHelp() *dto.MattermostResponse {
	return &dto.MattermostResponse{
		ResponseType: dto.ResponseTypeEphemeral,
		Text:         GetHelpText(),
	}
}

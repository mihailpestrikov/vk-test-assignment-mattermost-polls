package api

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"

	"vk-test-assignment-mattermost-polls/internal/api/dto"
	"vk-test-assignment-mattermost-polls/internal/model"
	"vk-test-assignment-mattermost-polls/internal/service"
	"vk-test-assignment-mattermost-polls/pkg/config"
	"vk-test-assignment-mattermost-polls/pkg/mattermost"
)

type Handler struct {
	pollService      service.IPollService
	mattermostCfg    config.MattermostConfig
	mattermostClient *mattermost.Client
}

func NewHandler(pollService *service.PollService, mattermostCfg config.MattermostConfig) *Handler {
	return &Handler{
		pollService:      pollService,
		mattermostCfg:    mattermostCfg,
		mattermostClient: mattermost.NewClient(mattermostCfg),
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.healthCheck) // Добавляем endpoint для проверки здоровья
	r.Post("/command", h.handleCommand)
}

type HealthCheckResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// @Summary Проверка здоровья сервиса
// @Description Проверяет доступность сервиса
// @ID health-check
// @Produce json
// @Tags Сервис
// @Success 200 {object} HealthCheckResponse
// @Router /health [get]
func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := HealthCheckResponse{
		Status:  "available",
		Version: "1.0.0",
	}

	render.JSON(w, r, response)
}

// @Summary Обработка команд Mattermost
// @Description Обработка всех slash-команд от Mattermost
// @ID process-command
// @Accept x-www-form-urlencoded
// @Produce json
// @Tags Команды
// @Param token formData string true "Токен верификации от Mattermost"
// @Param team_id formData string true "ID команды Mattermost"
// @Param team_domain formData string false "Домен команды Mattermost"
// @Param channel_id formData string true "ID канала, из которого отправлена команда"
// @Param channel_name formData string false "Название канала"
// @Param user_id formData string true "ID пользователя, отправившего команду"
// @Param user_name formData string false "Имя пользователя, отправившего команду"
// @Param command formData string true "Slash-команда (например, /poll)"
// @Param text formData string false "Текст, следующий за командой"
// @Param response_url formData string false "URL для отправки отложенных ответов"
// @Param trigger_id formData string false "ID триггера для интерактивных диалогов"
// @Success 200 {object} dto.MattermostResponse "Успешный ответ на команду"
// @Failure 400 {object} dto.MattermostResponse "Неправильный формат команды"
// @Failure 401 {object} dto.MattermostResponse "Недействительный токен"
// @Failure 403 {object} dto.MattermostResponse "Недостаточно прав для выполнения операции"
// @Failure 404 {object} dto.MattermostResponse "Голосование не найдено"
// @Failure 500 {object} dto.MattermostResponse "Внутренняя ошибка сервера"
// @Router /command [post]
func (h *Handler) handleCommand(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse form data")
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, mattermost.FormatError(errors.New("invalid request format")))
		return
	}

	var req dto.MattermostCommandRequest
	req.Token = r.FormValue("token")
	req.TeamID = r.FormValue("team_id")
	req.TeamDomain = r.FormValue("team_domain")
	req.ChannelID = r.FormValue("channel_id")
	req.ChannelName = r.FormValue("channel_name")
	req.UserID = r.FormValue("user_id")
	req.UserName = r.FormValue("user_name")
	req.Command = r.FormValue("command")
	req.Text = r.FormValue("text")
	req.ResponseURL = r.FormValue("response_url")
	req.TriggerID = r.FormValue("trigger_id")

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		log.Error().Err(err).Msg("Invalid request structure")
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, mattermost.FormatError(errors.New("invalid request parameters")))
		return
	}

	if req.Token != h.mattermostCfg.WebhookSecret {
		log.Warn().
			Str("received_token", req.Token).
			Str("expected_token", h.mattermostCfg.WebhookSecret).
			Msg("Invalid webhook token")
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, mattermost.FormatError(errors.New("invalid token")))
		return
	}

	log.Debug().
		Str("user_id", req.UserID).
		Str("channel_id", req.ChannelID).
		Str("command", req.Command).
		Str("text", req.Text).
		Msg("Received command")

	cmd, err := mattermost.ParseCommand(req.Text)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse command")
		render.JSON(w, r, mattermost.FormatError(err))
		return
	}

	switch cmd.SubCommand {
	case mattermost.CommandCreate:
		h.handleCreateCommand(w, r, req, cmd)

	case mattermost.CommandVote:
		h.handleVoteCommand(w, r, req, cmd)

	case mattermost.CommandResults:
		h.handleResultsCommand(w, r, req, cmd)

	case mattermost.CommandEnd:
		h.handleEndCommand(w, r, req, cmd)

	case mattermost.CommandDelete:
		h.handleDeleteCommand(w, r, req, cmd)

	case mattermost.CommandInfo:
		h.handleInfoCommand(w, r, req, cmd)

	case mattermost.CommandHelp:
		h.handleHelpCommand(w, r, req)

	default:
		log.Error().Str("subcommand", cmd.SubCommand).Msg("Unknown subcommand")
		render.JSON(w, r, mattermost.FormatError(errors.New("unknown command")))
	}
}

func (h *Handler) handleCreateCommand(w http.ResponseWriter, r *http.Request, req dto.MattermostCommandRequest, cmd *mattermost.Command) {
	poll, err := h.pollService.CreatePoll(cmd.Question, cmd.Options, req.UserID, req.ChannelID, cmd.Duration)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create poll")
		render.JSON(w, r, mattermost.FormatError(err))
		return
	}

	log.Info().
		Str("poll_id", poll.ID).
		Str("user_id", req.UserID).
		Str("channel_id", req.ChannelID).
		Msg("Poll created")

	render.JSON(w, r, mattermost.FormatPollCreated(poll))
}

func (h *Handler) handleVoteCommand(w http.ResponseWriter, r *http.Request, req dto.MattermostCommandRequest, cmd *mattermost.Command) {
	poll, err := h.pollService.GetPoll(cmd.PollID)
	if err != nil {
		log.Error().Err(err).Str("poll_id", cmd.PollID).Msg("Failed to get poll")
		render.JSON(w, r, mattermost.FormatError(err))
		return
	}

	err = h.pollService.Vote(cmd.PollID, req.UserID, cmd.OptionIdx)
	if err != nil {
		log.Error().Err(err).
			Str("poll_id", cmd.PollID).
			Str("user_id", req.UserID).
			Int("option_idx", cmd.OptionIdx).
			Msg("Failed to vote")
		render.JSON(w, r, mattermost.FormatError(err))
		return
	}

	log.Info().
		Str("poll_id", cmd.PollID).
		Str("user_id", req.UserID).
		Int("option_idx", cmd.OptionIdx).
		Msg("Vote recorded")

	render.JSON(w, r, mattermost.FormatVoteConfirmed(poll, cmd.OptionIdx))
}

func (h *Handler) handleResultsCommand(w http.ResponseWriter, r *http.Request, req dto.MattermostCommandRequest, cmd *mattermost.Command) {
	results, err := h.pollService.GetResults(cmd.PollID)
	if err != nil {
		log.Error().Err(err).Str("poll_id", cmd.PollID).Msg("Failed to get poll results")
		render.JSON(w, r, mattermost.FormatError(err))
		return
	}

	poll, err := h.pollService.GetPoll(cmd.PollID)
	if err != nil {
		log.Error().Err(err).Str("poll_id", cmd.PollID).Msg("Failed to get poll")
		render.JSON(w, r, mattermost.FormatError(err))
		return
	}

	ephemeral := poll.CreatedBy != req.UserID

	log.Info().
		Str("poll_id", cmd.PollID).
		Str("user_id", req.UserID).
		Bool("ephemeral", ephemeral).
		Msg("Poll results requested")

	render.JSON(w, r, mattermost.FormatPollResults(results, ephemeral))
}

func (h *Handler) handleEndCommand(w http.ResponseWriter, r *http.Request, req dto.MattermostCommandRequest, cmd *mattermost.Command) {
	results, err := h.pollService.EndPoll(cmd.PollID, req.UserID)
	if err != nil {
		if errors.Is(err, model.ErrNotPollCreator) {
			log.Warn().
				Err(err).
				Str("poll_id", cmd.PollID).
				Str("user_id", req.UserID).
				Msg("Unauthorized attempt to end poll")
			render.JSON(w, r, mattermost.FormatError(err))
			return
		}

		log.Error().Err(err).Str("poll_id", cmd.PollID).Msg("Failed to end poll")
		render.JSON(w, r, mattermost.FormatError(err))
		return
	}

	log.Info().
		Str("poll_id", cmd.PollID).
		Str("user_id", req.UserID).
		Msg("Poll ended")

	render.JSON(w, r, mattermost.FormatPollEnded(results))
}

func (h *Handler) handleDeleteCommand(w http.ResponseWriter, r *http.Request, req dto.MattermostCommandRequest, cmd *mattermost.Command) {
	err := h.pollService.DeletePoll(cmd.PollID, req.UserID)
	if err != nil {
		if errors.Is(err, model.ErrNotPollCreator) {
			log.Warn().
				Err(err).
				Str("poll_id", cmd.PollID).
				Str("user_id", req.UserID).
				Msg("Unauthorized attempt to delete poll")
			render.JSON(w, r, mattermost.FormatError(err))
			return
		}

		log.Error().Err(err).Str("poll_id", cmd.PollID).Msg("Failed to delete poll")
		render.JSON(w, r, mattermost.FormatError(err))
		return
	}

	log.Info().
		Str("poll_id", cmd.PollID).
		Str("user_id", req.UserID).
		Msg("Poll deleted")

	render.JSON(w, r, mattermost.FormatPollDeleted(cmd.PollID))
}

func (h *Handler) handleInfoCommand(w http.ResponseWriter, r *http.Request, req dto.MattermostCommandRequest, cmd *mattermost.Command) {
	poll, err := h.pollService.GetPoll(cmd.PollID)
	if err != nil {
		log.Error().Err(err).Str("poll_id", cmd.PollID).Msg("Failed to get poll")
		render.JSON(w, r, mattermost.FormatError(err))
		return
	}

	log.Info().
		Str("poll_id", cmd.PollID).
		Str("user_id", req.UserID).
		Msg("Poll info requested")

	render.JSON(w, r, mattermost.FormatPollInfo(poll))
}

func (h *Handler) handleHelpCommand(w http.ResponseWriter, r *http.Request, req dto.MattermostCommandRequest) {
	log.Debug().
		Str("user_id", req.UserID).
		Msg("Help requested")

	render.JSON(w, r, mattermost.FormatHelp())
}

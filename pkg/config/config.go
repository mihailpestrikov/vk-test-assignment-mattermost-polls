package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config представляет полную конфигурацию приложения
type Config struct {
	Server     ServerConfig
	Logger     LoggerConfig
	Tarantool  TarantoolConfig
	Mattermost MattermostConfig
	Poll       PollConfig
}

// ServerConfig содержит настройки HTTP-сервера
type ServerConfig struct {
	Host           string
	Port           string
	RequestTimeout time.Duration
	GinMode        string
	AppEnv         string
}

// LoggerConfig содержит настройки логирования
type LoggerConfig struct {
	Level      string
	Format     string // "json" или "pretty"
	Output     string // "stdout" или "file"
	File       string // путь к файлу логов
	WithCaller bool   // добавлять ли информацию о вызывающем файле и строке
}

// TarantoolConfig содержит настройки подключения к Tarantool
type TarantoolConfig struct {
	Host       string
	Port       string
	User       string
	Pass       string
	SpacePolls string
	SpaceVotes string
}

// MattermostConfig содержит настройки интеграции с Mattermost
type MattermostConfig struct {
	URL           string
	Token         string
	WebhookSecret string
}

// PollConfig содержит настройки для голосований
type PollConfig struct {
	DefaultDuration int
	MaxOptions      int
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found. Using environment variables.\n")
	}

	viper.AutomaticEnv()

	setDefaults()

	config := &Config{
		Server: ServerConfig{
			Host:           viper.GetString("HOST"),
			Port:           viper.GetString("PORT"),
			RequestTimeout: viper.GetDuration("REQUEST_TIMEOUT") * time.Second,
			GinMode:        viper.GetString("GIN_MODE"),
			AppEnv:         viper.GetString("APP_ENV"),
		},
		Logger: LoggerConfig{
			Level:      viper.GetString("LOG_LEVEL"),
			Format:     viper.GetString("LOG_FORMAT"),
			Output:     viper.GetString("LOG_OUTPUT"),
			File:       viper.GetString("LOG_FILE"),
			WithCaller: viper.GetBool("LOG_WITH_CALLER"),
		},
		Tarantool: TarantoolConfig{
			Host:       viper.GetString("TARANTOOL_HOST"),
			Port:       viper.GetString("TARANTOOL_PORT"),
			User:       viper.GetString("TARANTOOL_USER"),
			Pass:       viper.GetString("TARANTOOL_PASS"),
			SpacePolls: viper.GetString("TARANTOOL_SPACE_POLLS"),
			SpaceVotes: viper.GetString("TARANTOOL_SPACE_VOTES"),
		},
		Mattermost: MattermostConfig{
			URL:           viper.GetString("MATTERMOST_URL"),
			Token:         viper.GetString("MATTERMOST_TOKEN"),
			WebhookSecret: viper.GetString("MATTERMOST_WEBHOOK_SECRET"),
		},
		Poll: PollConfig{
			DefaultDuration: viper.GetInt("DEFAULT_POLL_DURATION"),
			MaxOptions:      viper.GetInt("MAX_OPTIONS"),
		},
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

func setDefaults() {
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("REQUEST_TIMEOUT", 30)
	viper.SetDefault("GIN_MODE", "release")
	viper.SetDefault("APP_ENV", "development")

	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "pretty")
	viper.SetDefault("LOG_OUTPUT", "stdout")
	viper.SetDefault("LOG_FILE", "logs/app.log")
	viper.SetDefault("LOG_WITH_CALLER", true)

	viper.SetDefault("TARANTOOL_HOST", "tarantool")
	viper.SetDefault("TARANTOOL_PORT", "3301")
	viper.SetDefault("TARANTOOL_USER", "guest")
	viper.SetDefault("TARANTOOL_SPACE_POLLS", "polls")
	viper.SetDefault("TARANTOOL_SPACE_VOTES", "votes")

	viper.SetDefault("DEFAULT_POLL_DURATION", 86400)
	viper.SetDefault("MAX_OPTIONS", 10)
}

func validateConfig(cfg *Config) error {
	if cfg.Mattermost.Token == "" {
		return fmt.Errorf("MATTERMOST_TOKEN is required")
	}
	if cfg.Mattermost.WebhookSecret == "" {
		return fmt.Errorf("MATTERMOST_WEBHOOK_SECRET is required")
	}

	return nil
}

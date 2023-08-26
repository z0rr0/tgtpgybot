package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TimeDuration is a wrapper for time.Duration to marshal/unmarshal to/from JSON
type TimeDuration struct {
	Duration time.Duration
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *TimeDuration) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	duration, err := time.ParseDuration(v)
	if err != nil {
		return err
	}

	d.Duration = duration
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (d *TimeDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

// String implements the fmt.Stringer interface.
func (d *TimeDuration) String() string {
	return d.Duration.String()
}

// Config is main config structure.
type Config struct {
	Token      string       `json:"token"`
	Timeout    TimeDuration `json:"timeout"`
	DebugLevel string       `json:"debug_level"`
	Users      []int64      `json:"users"`
	Chat       Chat         `json:"chat"`
	VerboseBot bool         `json:"-"`
	Offline    bool         `json:"-"`
}

// New creates new config from file.
func New(configFile string) (*Config, error) {
	fullPath, err := filepath.Abs(strings.Trim(configFile, " "))
	if err != nil {
		return nil, fmt.Errorf("config file: %w", err)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("config read: %w", err)
	}

	c := &Config{}
	if err = json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("config unmarshal: %w", err)
	}

	if err = c.Chat.init(); err != nil {
		return nil, fmt.Errorf("config init GPT: %w", err)
	}

	if err = c.initLogger(); err != nil {
		return nil, fmt.Errorf("config init logger: %w", err)
	}

	return c, nil
}

func (c *Config) initLogger() error {
	var level = new(slog.LevelVar)

	switch c.DebugLevel {
	case "debug":
		level.Set(slog.LevelDebug)
		c.VerboseBot = true
	case "info":
		level.Set(slog.LevelInfo)
	case "warn":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		return fmt.Errorf("unknown debug level: %q", c.DebugLevel)
	}

	handleOps := &slog.HandlerOptions{Level: level}
	logger := slog.New(slog.NewTextHandler(os.Stdout, handleOps))

	slog.SetDefault(logger)
	return nil
}

// LogValue implements slog.Value interface.
func (c *Config) LogValue() slog.Value {
	var b strings.Builder

	b.WriteString("token=")
	b.WriteString(hideParam(c.Token) + ", ")

	b.WriteString("timeout=" + c.Timeout.String() + ", ")
	b.WriteString(fmt.Sprintf("debug_level=%v, ", c.DebugLevel))

	b.WriteString("chat.api_key=")
	b.WriteString(hideParam(c.Chat.APIKey) + ", ")

	b.WriteString("chat.proxy=")
	b.WriteString(hideParam(c.Chat.Proxy))

	return slog.StringValue(b.String())
}

func hideParam(param string) string {
	if param == "" {
		return "empty"
	}

	return "****"
}

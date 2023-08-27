package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"

	"github.com/z0rr0/tgtpgybot/config"
)

// Bot is main bot structure.
type Bot struct {
	cfg  *config.Config
	bot  *telebot.Bot
	stop chan struct{}
}

// New creates new bot.
func New(cfg *config.Config) (*Bot, error) {
	poller := telebot.LongPoller{Timeout: 30 * time.Second, AllowedUpdates: []string{"message", "edited_message"}}

	pref := telebot.Settings{
		Token:       cfg.Token,
		Poller:      &poller,
		Synchronous: true,
		Verbose:     cfg.VerboseBot,
		Offline:     cfg.Offline,
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	// print some log info
	b.Use(middleware.Whitelist(cfg.Users...))
	// allow only users from config
	b.Use(durationMiddleware())

	return &Bot{cfg: cfg, bot: b, stop: make(chan struct{})}, nil
}

// Start starts the bot.
func (b *Bot) Start(sigChan chan os.Signal) {
	slog.Info("starting")

	go func() {
		//sigint := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, os.Signal(syscall.SIGTERM), os.Signal(syscall.SIGQUIT))
		s := <-sigChan

		slog.Info("stopping", "signal", s)
		b.bot.Stop()
		close(b.stop)
	}()

	b.bot.Handle(telebot.OnText, b.rootHandler)
	b.bot.Handle(telebot.OnEdited, b.rootHandler)

	b.bot.Start() // run forever, wait signal to stop
}

// Stop waits bot to stop.
func (b *Bot) Stop() {
	<-b.stop // wait graceful bot stop
}

// rootHandler handles incoming completion messages.
func (b *Bot) rootHandler(c telebot.Context) error {
	var (
		user      = c.Sender()
		messageID = c.Message().ID
		content   = strings.TrimSpace(c.Text())
	)

	slog.Info("generation", "id", messageID, "userID", user.ID)
	slog.Debug("generation", "id", messageID, "userID", user.ID, "text", content)

	ctx, cancel := context.WithTimeout(context.Background(), b.cfg.Timeout.Duration)
	defer cancel()

	result, err := b.cfg.Chat.Generation(ctx, content, messageID)
	if err != nil {
		slog.Error("failed", "id", messageID, "error", err)
		result = "ERROR: failed to get completion: " + err.Error()
	}

	return c.Send(result, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

// durationMiddleware is common middleware function to log duration of handler.
func durationMiddleware() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			var (
				start     = time.Now()
				messageID = c.Message().ID
				user      = c.Sender()
			)
			defer func() {
				slog.Info(
					"handled",
					"id", messageID,
					"duration", time.Since(start).Truncate(100*time.Millisecond),
				)
			}()
			slog.Info("got", "id", messageID, "user", user.Username)

			if err := next(c); err != nil {
				// the error occurred inside the handler
				return c.Send("oops, an error has occurred\n\n" + err.Error())
			}

			return nil
		}
	}
}

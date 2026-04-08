// Package bot contains chat adapters.
package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/coffeegraph/coffeegraph/internal/queue"
	"github.com/coffeegraph/coffeegraph/internal/runner"
)

// StatusProvider returns status metadata for /status command.
type StatusProvider interface {
	QueueDepth() (int, error)
	Skills() ([]string, error)
}

// TelegramAdapter handles Telegram messaging for queue + execution.
type TelegramAdapter struct {
	Token      string
	AllowedIDs map[int64]struct{}
	Engine     *runner.Engine
	Status     StatusProvider
}

// Run starts polling Telegram updates until context cancellation.
func (a *TelegramAdapter) Run(ctx context.Context) error {
	b, err := tgbotapi.NewBotAPI(a.Token)
	if err != nil {
		return fmt.Errorf("create telegram bot: %w", err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := b.GetUpdatesChan(u)

	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			return nil
		case up, ok := <-updates:
			if !ok {
				time.Sleep(backoff)
				if backoff < 30*time.Second {
					backoff *= 2
				}
				continue
			}
			backoff = time.Second
			if up.Message == nil {
				continue
			}
			chatID := up.Message.Chat.ID
			if !a.isAllowed(chatID) {
				continue
			}
			txt := strings.TrimSpace(up.Message.Text)
			switch {
			case strings.EqualFold(txt, "/status"):
				a.handleStatus(ctx, b, chatID)
			case strings.EqualFold(txt, "/coffee"):
				a.handleCoffee(ctx, b, chatID)
			case strings.HasPrefix(txt, "@"):
				a.handleTask(ctx, b, chatID, txt)
			default:
				a.send(b, chatID, "Use @skill task, /status, or /coffee")
			}
		}
	}
}

func (a *TelegramAdapter) handleStatus(_ context.Context, b *tgbotapi.BotAPI, chatID int64) {
	depth := 0
	if a.Status != nil {
		d, err := a.Status.QueueDepth()
		if err == nil {
			depth = d
		}
	}
	a.send(b, chatID, fmt.Sprintf("Queue depth: %d", depth))
}

func (a *TelegramAdapter) handleCoffee(ctx context.Context, b *tgbotapi.BotAPI, chatID int64) {
	a.send(b, chatID, "Your agency is brewing ☕")
	s, err := a.Engine.ExecutePending(ctx, 0, runner.ModeNormal)
	if err != nil {
		a.send(b, chatID, "Error running coffee: "+err.Error())
		return
	}
	a.send(b, chatID, fmt.Sprintf("Done. Completed %d tasks. Errors: %d", s.Completed, len(s.Errors)))
}

func (a *TelegramAdapter) handleTask(ctx context.Context, b *tgbotapi.BotAPI, chatID int64, text string) {
	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 2 {
		a.send(b, chatID, "Format: @skill your task")
		return
	}
	skill := strings.TrimPrefix(parts[0], "@")
	task := strings.TrimSpace(parts[1])
	it := queue.Item{Skill: skill, Task: task, Priority: 3}
	a.send(b, chatID, "Your agency is brewing ☕")
	r, err := a.Engine.ExecuteTask(ctx, it)
	if err != nil {
		a.send(b, chatID, "Task failed: "+err.Error())
		return
	}
	msg := r.Text
	if len(msg) > 3800 {
		msg = msg[:3800] + "..."
	}
	a.send(b, chatID, msg)
}

func (a *TelegramAdapter) send(b *tgbotapi.BotAPI, chatID int64, text string) {
	_, _ = b.Send(tgbotapi.NewMessage(chatID, text))
}

func (a *TelegramAdapter) isAllowed(id int64) bool {
	if len(a.AllowedIDs) == 0 {
		return false
	}
	_, ok := a.AllowedIDs[id]
	return ok
}

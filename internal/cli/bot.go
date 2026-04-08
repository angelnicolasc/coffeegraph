package cli

import (
	"context"
	"fmt"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/coffeegraph/coffeegraph/internal/bot"
	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/project"
	"github.com/coffeegraph/coffeegraph/internal/queue"
	"github.com/coffeegraph/coffeegraph/internal/runner"
)

// RunBot starts the Telegram bot loop.
func RunBot() error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(root, "config.yaml"))
	if err != nil {
		return err
	}
	token := cfg.TelegramToken()
	if token == "" {
		return fmt.Errorf("could not connect to Telegram. Make sure TELEGRAM_BOT_TOKEN is set in your environment or config.yaml")
	}
	allowed := make(map[int64]struct{}, len(cfg.Bot.AllowedChatIDs))
	for _, id := range cfg.Bot.AllowedChatIDs {
		allowed[id] = struct{}{}
	}
	if len(allowed) == 0 {
		return fmt.Errorf("telegram bot requires bot.allowed_chat_ids in config.yaml for safety")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	engine := &runner.Engine{Root: root, Cfg: cfg}
	adapter := &bot.TelegramAdapter{
		Token:      token,
		AllowedIDs: allowed,
		Engine:     engine,
		Status:     botStatus{root: root},
	}
	fmt.Println("Telegram bot is running. Press Ctrl+C to stop.")
	return adapter.Run(ctx)
}

type botStatus struct{ root string }

func (b botStatus) QueueDepth() (int, error) {
	items, err := queue.Read(b.root)
	if err != nil {
		return 0, err
	}
	return len(items), nil
}

func (b botStatus) Skills() ([]string, error) {
	return nil, nil
}

package notifier

import (
	"context"
	"fmt"
	"log"
	"time"

	"tradingview-lite/backend/internal/alert"
	"tradingview-lite/backend/internal/store"
)

type Outbox struct {
	store       store.Store
	discord     *Discord
	maxAttempts int
}

func NewOutbox(repo store.Store, discord *Discord) *Outbox {
	return &Outbox{store: repo, discord: discord, maxAttempts: 5}
}

func (o *Outbox) Send(ctx context.Context, event alert.Event) error {
	if o == nil || o.store == nil {
		return fmt.Errorf("discord outbox is not configured")
	}
	_, err := o.store.EnqueueDiscordDelivery(ctx, store.DiscordDelivery{IdempotencyKey: "discord:" + event.ID, Event: event, NextAttemptAt: time.Now().UTC()})
	return err
}

func (o *Outbox) DispatchOnce(ctx context.Context) error {
	if o == nil || o.store == nil || o.discord == nil {
		return nil
	}
	rows, err := o.store.ClaimDiscordDeliveries(ctx, time.Now().UTC(), 20)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := o.discord.Send(ctx, row.Event); err == nil {
			if err := o.store.CompleteDiscordDelivery(ctx, row.ID); err != nil {
				return err
			}
			continue
		}
		backoff := time.Duration(1<<min(row.Attempts, 6)) * time.Second
		dead := row.Attempts >= o.maxAttempts
		if retryErr := o.store.RetryDiscordDelivery(ctx, row.ID, time.Now().UTC().Add(backoff), err.Error(), dead); retryErr != nil {
			return retryErr
		}
		log.Printf("discord delivery deferred: id=%d attempts=%d dead=%t err=%v", row.ID, row.Attempts, dead, err)
	}
	return nil
}

func (o *Outbox) Start(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 10 * time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			if err := o.DispatchOnce(ctx); err != nil {
				log.Printf("discord outbox dispatch failed: %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package domain

import (
	"context"
	"time"
)

type SubscriptionRepository interface {
	FindDue(ctx context.Context, due time.Time) ([]Subscription, error)
	UpdateSubscription(ctx context.Context, userId string, interval time.Duration, nextSendAt time.Time) error 
}


package domain

import "time"

type Subscription struct {
	UserId     string
	NextSendAt time.Time
}
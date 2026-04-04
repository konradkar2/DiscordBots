package domain

import "context"

type UsersRepository interface {
	Insert(context.Context, User) error
}
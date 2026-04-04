package domain

import "context"

type NotesRepository interface {
	GetRandom(context.Context) (Note, error)
}
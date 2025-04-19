package pips

import "context"

type Stage interface {
	Run(context.Context, <-chan D[any]) <-chan D[any]
}

type stages[I any, O any] []Stage

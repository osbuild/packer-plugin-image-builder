package ibk

import "context"

type Pusher interface {
	Push(ctx context.Context, contents, extension string) (string, error)
}

type Executor interface {
	Execute(ctx context.Context, cmd Command, opts ...ExecuteOpt) error
}

type Closer interface {
	Close(ctx context.Context) error
}

type Transport interface {
	Pusher
	Executor
	Closer
}

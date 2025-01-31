package ibk

import "context"

// StringCommand is a command that is just a string. Useful for testing purposes.
// The caller is responsible for escaping the command properly.
type StringCommand string

var _ Command = StringCommand("")

func (c StringCommand) Configure(ctx context.Context, exec Executor) error {
	return nil
}

func (c StringCommand) Push(ctx context.Context, pusher Pusher) error {
	return nil
}

func (c StringCommand) Build() string {
	return string(c)
}

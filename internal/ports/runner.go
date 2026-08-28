package ports

import "context"

type Command struct {
	Program string
	Args    []string
}

type Result struct {
	Command Command
	Output  string
	Err     error
}

type Runner interface {
	Run(context.Context, Command) Result
}

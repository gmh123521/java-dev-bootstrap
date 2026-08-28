package logging

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type Logger struct {
	out io.Writer
	mu  sync.Mutex
}

func New(out io.Writer) *Logger { return &Logger{out: out} }

type Runner struct {
	Inner  ports.Runner
	Logger *Logger
}

func (r Runner) Run(ctx context.Context, command ports.Command) ports.Result {
	result := r.Inner.Run(ctx, command)
	l := r.Logger
	if l == nil || l.out == nil {
		return result
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	status := "成功"
	if result.Err != nil {
		status = result.Err.Error()
	}
	fmt.Fprintf(l.out, "[%s] %s %v\n", time.Now().Format(time.RFC3339), status, command)
	if result.Output != "" {
		fmt.Fprintf(l.out, "输出：%s\n", result.Output)
	}
	return result
}

package logx

import (
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

type Level int

const (
	LevelInfo Level = iota
	LevelSuccess
	LevelWarn
	LevelError
)

type Logger struct {
	Out        io.Writer
	Err        io.Writer
	ForceColor bool
}

func New(out, err io.Writer) *Logger {
	if out == nil {
		out = os.Stdout
	}
	if err == nil {
		err = os.Stderr
	}
	return &Logger{Out: out, Err: err}
}

func (l *Logger) Info(msg string) {
	l.write(LevelInfo, msg)
}

func (l *Logger) Infof(format string, args ...any) {
	l.write(LevelInfo, fmt.Sprintf(format, args...))
}

func (l *Logger) Success(msg string) {
	l.write(LevelSuccess, msg)
}

func (l *Logger) Successf(format string, args ...any) {
	l.write(LevelSuccess, fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(msg string) {
	l.write(LevelWarn, msg)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.write(LevelWarn, fmt.Sprintf(format, args...))
}

func (l *Logger) Error(msg string) {
	l.write(LevelError, msg)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.write(LevelError, fmt.Sprintf(format, args...))
}

func (l *Logger) write(level Level, msg string) {
	w := l.outputFor(level)
	colorize := l.ForceColor || shouldColor(w)
	line := fmt.Sprintf("[%s] %s", levelLabel(level), msg)
	if colorize {
		line = color(level) + line + "\033[0m"
	}
	fmt.Fprintln(w, line)
}

func (l *Logger) outputFor(level Level) io.Writer {
	switch level {
	case LevelWarn, LevelError:
		return l.Err
	default:
		return l.Out
	}
}

func shouldColor(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fd := f.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

func levelLabel(level Level) string {
	switch level {
	case LevelSuccess:
		return "SUCCESS"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

func color(level Level) string {
	switch level {
	case LevelSuccess:
		return "\033[32m"
	case LevelWarn:
		return "\033[33m"
	case LevelError:
		return "\033[31m"
	default:
		return "\033[34m"
	}
}

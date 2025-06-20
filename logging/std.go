package logging

import (
	"fmt"
	"log"
	"sync"
)

type StdLogger struct {
	mu    sync.RWMutex
	level Level
}

func NewStdLogger() *StdLogger {
	return &StdLogger{ level: LevelInfo | LevelError | LevelDebug | LevelWarn }
}

func (l *StdLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *StdLogger) filterLevel(level Level) Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return level & l.level
}

func (l *StdLogger) Log(level Level, msg string, args ...any) {
	filtered := l.filterLevel(level)
    	if filtered > 0 {
		var level Level = Level(1)
		for filtered > 0 {
			if filtered & 1 > 0 {
				log.Printf("[%s] %s", level.String(), fmt.Sprintf(msg, args...))
			}
			filtered >>= 1
			level <<= 1
		}
	}
}

func (l *StdLogger) Debug(msg string, args ...any) { l.Log(LevelDebug, msg, args...) }
func (l *StdLogger) Info(msg string, args ...any)  { l.Log(LevelInfo, msg, args...) }
func (l *StdLogger) Warn(msg string, args ...any)  { l.Log(LevelWarn, msg, args...) }
func (l *StdLogger) Error(msg string, args ...any) { l.Log(LevelError, msg, args...) }

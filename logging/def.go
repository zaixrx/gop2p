package logging

type Level int

const (
    LevelDebug Level = 1 << iota
    LevelInfo
    LevelWarn
    LevelError
)

func (l Level) String() string {
	return map[Level]string{
		LevelDebug: "DEBUG",
		LevelInfo: "INFO",
		LevelWarn: "WARN",
		LevelError: "ERROR",
	}[l]
}

type Logger interface {
    SetLevel(level Level)
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
}

package gamelink

import "fmt"

type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
)

func (l LogLevel) DefaultEmoji() string {
	switch l {
	case LogDebug:
		return "🔍"
	case LogInfo:
		return "ℹ️"
	case LogWarn:
		return "⚠️"
	case LogError:
		return "❌"
	default:
		return "❓"
	}
}

type LogEntry struct {
	Level   LogLevel
	Emoji   string
	Message string
}

func (e LogEntry) ResolvedEmoji() string {
	if e.Emoji != "" {
		return e.Emoji
	}
	return e.Level.DefaultEmoji()
}

func (e LogEntry) String() string {
	return fmt.Sprintf("%s %s", e.ResolvedEmoji(), e.Message)
}

func ParseLogLevel(s string) LogLevel {
	switch s {
	case "debug":
		return LogDebug
	case "info":
		return LogInfo
	case "warn":
		return LogWarn
	case "error":
		return LogError
	default:
		return LogInfo
	}
}

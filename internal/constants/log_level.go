// Package constants provides log level constants for the controller.
package constants

type LogLevel string

const (
	LogLevelDebug   LogLevel = "DEBUG"
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarning LogLevel = "WARNING"
	LogLevelError   LogLevel = "ERROR"
)

func (l LogLevel) IsValid() bool {
	switch l {
	case LogLevelDebug, LogLevelInfo, LogLevelWarning, LogLevelError:
		return true
	default:
		return false
	}
}

func (l LogLevel) String() string {
	return string(l)
}

func GetSupportedLevels() []LogLevel {
	return []LogLevel{
		LogLevelDebug,
		LogLevelInfo,
		LogLevelWarning,
		LogLevelError,
	}
}

package core

import "time"

type AlertLevel int

const (
	Info AlertLevel = iota
	Warn
	Error
	Fatal
	Debug
)

const (
	CriticalEmoji = "🤯"
	ErrorEmoji    = "😵‍💫"
	WarnEmoji     = "😶‍🌫"
)

func (al AlertLevel) Color() int {
	switch al {
	case Fatal:
		return 0xFF0000
	case Error:
		return 0xFF8000
	case Warn:
		return 0xFFBF00
	case Info:
		return 0xFFFFFF
	case Debug:
		return 0x00FF00
	}

	return 0
}

func (al AlertLevel) String() string {
	switch al {
	case Fatal:
		return CriticalEmoji + " Critical"
	case Error:
		return ErrorEmoji + " Error"
	case Warn:
		return WarnEmoji + " Warning"
	case Info:
		return "Information"
	case Debug:
		return "Debug"
	}
	return ""
}

type Alert struct {
	service   string
	level     AlertLevel
	msg       string
	logs      []string
	timestamp time.Time
}

func NewAlert(service string, level AlertLevel, msg string, timestamp time.Time, logs []string) *Alert {
	return &Alert{service, level, msg, logs, timestamp}
}

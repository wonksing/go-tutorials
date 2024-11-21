package types

type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
	FatalLevel Level = "fatal"
	PanicLevel Level = "panic"
)

// type LogLevel int8

// const (
// 	DebugLogLevel LogLevel = -1 + iota
// 	InfoLogLevel
// 	WarnLogLevel
// 	ErrorLogLevel
// 	FatalLogLevel
// 	PanicLogLevel
// 	IgnoreLogLevel
// )

// func ToLogLevel(lvl Level) LogLevel {
// 	switch lvl {
// 	case DebugLevel:
// 		return DebugLogLevel
// 	case InfoLevel:
// 		return InfoLogLevel
// 	case WarnLevel:
// 		return WarnLogLevel
// 	case ErrorLevel:
// 		return ErrorLogLevel
// 	case FatalLevel:
// 		return FatalLogLevel
// 	case PanicLevel:
// 		return PanicLogLevel
// 	default:
// 		return DebugLogLevel
// 	}
// }

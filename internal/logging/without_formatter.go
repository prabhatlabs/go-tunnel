package logging

func Info(m ...any) {
	logger(InfoLevel, m...)
}

func Warn(m ...any) {
	logger(WarnLevel, m...)
}

func Error(m ...any) {
	logger(ErrorLevel, m...)
}

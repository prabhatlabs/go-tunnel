package logging

import "fmt"

func Infof(s string, a ...any) {
	f := fmt.Sprintf(s, a...)
	logger(InfoLevel, f)
}

func Warnf(s string, a ...any) {
	f := fmt.Sprintf(s, a...)
	logger(WarnLevel, f)
}

func Errorf(s string, a ...any) {
	f := fmt.Sprintf(s, a...)
	logger(ErrorLevel, f)
}

package logging

import (
	"log"
)

const (
	InfoLevel  = "[INFO]"
	WarnLevel  = "[WARN]"
	ErrorLevel = "[ERROR]"
)

func logger(t string, m ...any) {
	m = append([]any{t}, m...)
	if t == ErrorLevel {
		log.Fatalln(m...)
		return
	}
	log.Println(m...)
}

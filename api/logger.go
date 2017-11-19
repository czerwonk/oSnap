package api

import "log"

type logger struct {

}

func (l *logger) Debug(args ...string) {
	log.Println(args)
}

func (l *logger) Info(args ...string) {
	log.Println(args)
}

func (l *logger) Error(args ...string) {
	log.Println(args)
}

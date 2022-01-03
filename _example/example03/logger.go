package main

import (
	"fmt"

	"github.com/appleboy/graceful"

	"github.com/rs/zerolog/log"
)

var _ graceful.Logger = (*logger)(nil)

type logger struct{}

func (l logger) Infof(format string, args ...interface{}) {
	log.Info().Msgf(format, args...)
}

func (l logger) Errorf(format string, args ...interface{}) {
	log.Error().Msgf(format, args...)
}

func (l logger) Fatalf(format string, args ...interface{}) {
	log.Fatal().Msgf(format, args...)
}

func (l logger) Info(args ...interface{}) {
	log.Info().Msg(fmt.Sprint(args...))
}

func (l logger) Error(args ...interface{}) {
	log.Error().Msg(fmt.Sprint(args...))
}

func (l logger) Fatal(args ...interface{}) {
	log.Fatal().Msg(fmt.Sprint(args...))
}

package main

import (
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

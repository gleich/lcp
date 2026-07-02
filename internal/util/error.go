package util

import (
	"net/http"

	"github.com/rs/zerolog"
)

func InternalServerError(w http.ResponseWriter, err error, logger *zerolog.Logger, msg string) {
	logger.Error().Err(err).Msg(msg)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

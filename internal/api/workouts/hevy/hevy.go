package hevy

import (
	"sync"

	"github.com/rs/zerolog"
	"go.mattglei.ch/lcp/internal/cache"
)

var logger = sync.OnceValue(func() *zerolog.Logger { return cache.Workouts.Logger() })

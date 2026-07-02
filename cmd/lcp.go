package main

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mattglei.ch/lcp/internal/api"
	"go.mattglei.ch/lcp/internal/api/applemusic"
	"go.mattglei.ch/lcp/internal/api/github"
	"go.mattglei.ch/lcp/internal/api/steam"
	"go.mattglei.ch/lcp/internal/api/workouts"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/internal/middleware"
	"go.mattglei.ch/lcp/internal/secrets"
)

func main() {
	start := time.Now()
	secrets.Load()
	if !secrets.ENV.StructuredLogging {
		ny, err := time.LoadLocation("America/New_York")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load new york timezone")
		}
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:          os.Stderr,
			TimeFormat:   "01/02 03:04:05 PM MST",
			TimeLocation: ny,
		})
	}

	log.Info().Dur("duration", time.Since(start)).Msg("booted")

	minioClient, err := minio.New(secrets.ENV.MinioEndpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			secrets.ENV.MinioAccessKeyID,
			secrets.ENV.MinioSecretKey,
			"",
		),
		Secure: true,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create minio client")
	}

	var (
		client = api.IPV4OnlyClient()
		mux    = http.NewServeMux()
		rdb    = redis.NewClient(&redis.Options{
			Addr: secrets.ENV.RedisAddress,
			DB:   0,
			MaintNotificationsConfig: &maintnotifications.Config{
				Mode: maintnotifications.ModeDisabled,
			},
		})
	)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://mattglei.ch/writing/lcp", http.StatusPermanentRedirect)
	})

	setups := map[cache.CacheInstance]func(){
		cache.GitHub:     func() { github.Setup(mux) },
		cache.Workouts:   func() { workouts.Setup(mux, client, minioClient, rdb) },
		cache.Steam:      func() { steam.Setup(mux, client, rdb) },
		cache.AppleMusic: func() { applemusic.Setup(mux, client, rdb) },
	}
	var wg sync.WaitGroup
	for cacheInstance, setup := range setups {
		wg.Go(func() {
			start := time.Now()
			logger := cacheInstance.Logger()
			logger.Info().Msg("setting up")
			setup()
			logger.Info().Dur("duration", time.Since(start)).Msg("setup")
		})
	}
	wg.Wait()

	log.Info().Dur("duration", time.Since(start)).Msg("starting server")
	server := &http.Server{
		Addr:         ":8000",
		Handler:      middleware.Log(mux),
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start router")
	}
}

package strava

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/internal/secrets"
	"go.mattglei.ch/lcp/internal/util"
	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
)

type event struct {
	AspectType     string            `json:"aspect_type"`
	EventTime      int64             `json:"event_time"`
	ObjectID       int64             `json:"object_id"`
	ObjectType     string            `json:"object_type"`
	OwnerID        int64             `json:"owner_id"`
	SubscriptionID int64             `json:"subscription_id"`
	Updates        map[string]string `json:"updates"`
}

func EventRoute(
	client *http.Client,
	workoutsCache *cache.Cache[[]lcp.Workout],
	minioClient *minio.Client,
	rdb *redis.Client,
	fetch func(client *http.Client,
		minioClient *minio.Client,
		rdb *redis.Client,
		stravaTokens Tokens) ([]lcp.Workout, error),
	tokens Tokens,
) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		r.Body = http.MaxBytesReader(w, r.Body, 5<<20) // 5 MiB
		defer func() { _ = r.Body.Close() }()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			util.InternalServerError(w, err, logAttr, "reading response body failed")
			timber.Error(err, "reading response body failed", logAttr)
			return
		}

		var eventData event
		err = json.Unmarshal(body, &eventData)
		if err != nil {
			timber.Debug(string(body), logAttr)
			util.InternalServerError(w, err, logAttr, "failed to parse json")
			return
		}

		if eventData.SubscriptionID != secrets.ENV.StravaSubscriptionID {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		err = tokens.RefreshIfExpired(client)
		if err != nil {
			util.InternalServerError(w, err, logAttr, "failed to refresh token")
			return
		}

		activities, err := fetch(client, minioClient, rdb, tokens)
		if err != nil {
			util.InternalServerError(w, err, logAttr, "failed to update strava cache")
			return
		}
		workoutsCache.Update(start, activities)
	})
}

func ChallengeRoute(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	verifyToken := r.URL.Query().Get("hub.verify_token")
	if verifyToken != secrets.ENV.StravaVerifyToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	challenge := r.URL.Query().Get("hub.challenge")
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(struct {
		Challenge string `json:"hub.challenge"`
	}{Challenge: challenge})
	if err != nil {
		timber.Error(err, "failed to write json")
	}
	timber.DoneSince(start, "handled challenge successfully", logAttr)
}

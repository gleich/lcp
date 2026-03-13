package tasks

import "go.mattglei.ch/tlog"

var (
	StartServer tlog.Task
	Endpoint    tlog.Task

	Cache = tlog.Group("cache", &struct {
		MarshalResponse tlog.Task
		Update          tlog.Task
		StreamUpdate    tlog.Task

		ServeStream tlog.Task

		AppleMusic struct {
			Setup               tlog.Task
			FetchRecentlyPlayed tlog.Task
			FetchPlaylist       tlog.Task
			Songs               struct {
				DiffList tlog.Task
			}
		}
		GitHub struct {
			Setup            tlog.Task
			FetchPinnedRepos tlog.Task
		}
		Steam struct {
			Setup tlog.Task
			Fetch struct {
				RecentlyPlayed         tlog.Task
				AchievementsPercentage tlog.Task
				GameSchema             tlog.Task
			}
		}
		Workouts struct {
			Setup  tlog.Task
			Strava struct {
				Fetch struct {
					RefreshTokens   tlog.Task
					Activities      tlog.Task
					ActivityDetails tlog.Task
					Heartrate       tlog.Task
					Location        tlog.Task
					Map             tlog.Task
				}
				Event tlog.Task
			}
		}
	}{})

	Images = tlog.Group("images", &struct {
		CreateBlurHash tlog.Task
	}{})
)

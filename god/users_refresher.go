package god

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/wholesome-ghoul/persona-prototype-6/constants"
	"github.com/wholesome-ghoul/persona-prototype-6/models"
)

var yes = struct{}{}

// usersRefresher adds new users from database to redis sorted set and map
func (g *God) UsersRefresher(totalCrawlers int, usersAvailable chan<- interface{}) {
	timer := time.NewTicker(10 * time.Second)
	defer timer.Stop()

	if err := g.usersRefresher(totalCrawlers, usersAvailable); err != nil {
		panic(err)
	}

	for {
		<-timer.C
		if l := g.CrawlPQueue.TotalElems(); l > constants.USERS_REFRESHER_THRESHOLD {
			continue
		}

		_ = g.usersRefresher(totalCrawlers, usersAvailable)
	}
}

func (g *God) usersRefresher(totalCrawlers int, usersAvailable chan<- interface{}) error {
	var excludedUsernames []string
	g.redditUsersMap.Lock()
	for username := range g.redditUsersMap.Items {
		excludedUsernames = append(excludedUsernames, username)
	}
	g.redditUsersMap.Unlock()

	weekday := int8(time.Now().Weekday())
	personasWithActivitiesMap, err := models.SelectByRedisScore(g.personaDB.DB(), weekday, excludedUsernames...)
	if err != nil {
		g.Log.Error().Stack().Err(err).Msg("couldn't select entries by redis score")
		return err
	}

	var newMembers []redis.Z
	var totalUsersInMap int
	g.redditUsersMap.Lock()
	for username, value := range personasWithActivitiesMap {
		newMembers = append(newMembers, redis.Z{
			Score:  value.RedisScore,
			Member: username,
		})

		tmp := value
		g.redditUsersMap.Items[username] = &tmp
	}
	for range g.redditUsersMap.Items {
		totalUsersInMap++
	}
	g.redditUsersMap.Unlock()
	g.Log.Info().Msgf("map: new users %d, total %d", len(newMembers), totalUsersInMap)

	select {
	case <-g.CrawlPQueue.Context().Done():
		g.Log.Info().Msg("pqueue insertion canceled")
		return nil
	default:
	}

	g.CrawlPQueue.Add(newMembers...)
	g.Log.Info().Msgf("pqueue: new users %d, total %d", len(newMembers), g.CrawlPQueue.TotalElems())

	totalSignals := max(len(newMembers), g.CrawlPQueue.TotalElems())
	totalSignals = min(totalSignals, totalCrawlers) - len(usersAvailable)
	for i := 0; i < totalSignals; i++ {
		usersAvailable <- yes
	}

	return nil
}

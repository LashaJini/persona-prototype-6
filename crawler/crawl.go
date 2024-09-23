package crawler

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/wholesome-ghoul/persona-prototype-6/constants"
	"github.com/wholesome-ghoul/persona-prototype-6/god"
	"github.com/wholesome-ghoul/persona-prototype-6/models"
	pb "github.com/wholesome-ghoul/persona-prototype-6/protos"
	"github.com/wholesome-ghoul/persona-prototype-6/reddit"
	"github.com/wholesome-ghoul/persona-prototype-6/storage/database"
)

var limit = constants.REDDIT_CONTENT_LIMIT

func Crawl(ctx context.Context, usersAvailable chan any, id int, g *god.God) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var topElement redis.Z
		for {
			if g.CrawlPQueue.TotalElems() > 0 {
				topElement = g.CrawlPQueue.PopMax()
				break
			}

			g.Log.Info().Msgf("crawler #%d waiting", id)
			<-usersAvailable
		}

		username, ok := topElement.Member.(string)
		if !ok {
			continue
		}
		var err error

		// If a user was previously scanned and priority score was not good enough,
		// give them another chance. This also means that somehow crawler scanned all
		// other important (higher priority score) users
		personaWithActivities := models.NewPersonaWithActivities()
		personaWithActivities.Persona.Thirdparty.About, err = fetchAbout(ctx, g, id, username)
		if err != nil {
			g.Log.Error().Stack().Err(err).Send()
			continue
		}
		priorityScore := calculatePriorityScore(personaWithActivities)

		var comments Content[reddit.Comment]
		var posts Content[reddit.Post]
		if priorityScore > constants.PRIORITY_SCORE_THRESHOLD {
			var wg sync.WaitGroup
			lastCommentsAreSpamC := make(chan bool, 1)
			lastPostsAreSpamC := make(chan bool, 1)

			wg.Add(1)
			go func() {
				defer wg.Done()
				comments = fetchComments(ctx, g, id, username, lastCommentsAreSpamC, lastPostsAreSpamC)
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				posts = fetchPosts(ctx, g, id, username, lastPostsAreSpamC, lastCommentsAreSpamC)
			}()

			wg.Wait()
		}

		go func() {
			usernames := filterNewUsernames(g, comments)
			if len(usernames) == 0 {
				return
			}

			insertPersonas(g, usernames)
		}()

		priorityScore = calculatePriorityScore(personaWithActivities)
		commentLatestActivity := reddit.GetLatestActivityEpoch(&comments.items)
		postLatestActivity := reddit.GetLatestActivityEpoch(&posts.items)

		g.Lock()
		personaID := g.Get(username).Persona.ID.String()
		g.Unlock()

		func() {
			var err error
			instruction := database.NewMultiInstruction(ctx, g.PersonaDB())
			if err := instruction.Begin(); err != nil {
				g.Log.Error().Stack().Err(err).Send()
				return
			}

			var commentIDs []string
			var postIDs []string
			contents := &pb.Contents{}
			defer func() {
				if err != nil {
					g.Log.Error().Stack().Err(err).Send()

					if err := instruction.Rollback(); err != nil {
						g.Log.Error().Stack().Err(err).Send()
					}

					contentIDs := &pb.ContentIDs{Items: contents.Ids}
					if err := g.RPCClient.Rollback(ctx, contentIDs); err != nil {
						g.Log.Error().Stack().Err(err).Send()
					}
					return
				}

				if err := instruction.Commit(); err != nil {
					g.Log.Error().Stack().Err(err).Send()
				}
			}()

			commentIDs, err = models.InsertContentsTx(instruction, personaID, comments.items, comments.spamProbs)
			if err != nil {
				return
			}
			postIDs, err = models.InsertContentsTx(instruction, personaID, posts.items, posts.spamProbs)
			if err != nil {
				return
			}

			g.Lock()
			defer g.Unlock()

			contents.Ids = append(commentIDs, postIDs...)
			contents.Texts = append(comments.items.Texts(), posts.items.Texts()...)
			numberOfContents := len(contents.Ids)
			if numberOfContents > 0 {
				if err = g.RPCClient.InsertContents(ctx, contents); err != nil {
					return
				}
			}

			user := g.Get(username)

			oldestCommentEpoch := reddit.GetOldestActivityEpoch(&comments.items)
			oldestPostEpoch := reddit.GetOldestActivityEpoch(&posts.items)
			oldestContentEpoch := min(oldestCommentEpoch, oldestPostEpoch)

			startingYear := time.Unix(oldestContentEpoch, 0).UTC().Year()
			activeDays := initActiveDaysMap(startingYear)
			updateActivities(user, activeDays, &comments.items)
			updateActivities(user, activeDays, &posts.items)

			now := time.Now().UTC()
			startScanningFromEpoch := user.Persona.FirstScannedAt
			if userIsScannedFirstTime(startScanningFromEpoch) {
				startScanningFromEpoch = oldestContentEpoch
				user.Persona.FirstScannedAt = startScanningFromEpoch
			}

			user.WeeklyActivity.CalculateProbs(startScanningFromEpoch, now.Unix())

			if err = user.WeeklyActivity.UpdateWeeklyActivityTx(instruction); err != nil {
				g.Log.Error().Stack().Err(err).Send()
				return
			}
			if err = user.YearlyActivity.UpdateYearlyActivityTx(instruction); err != nil {
				g.Log.Error().Stack().Err(err).Send()
				return
			}

			newAverageSpam :=
				averageWithLength(
					user.Persona.TotalContentCount+comments.averageSpamCalculatedFor+posts.averageSpamCalculatedFor,
					user.Persona.AverageSpam*float32(user.Persona.TotalContentCount),
					comments.averageSpam*float32(comments.averageSpamCalculatedFor),
					posts.averageSpam*float32(posts.averageSpamCalculatedFor),
				)

			user.Persona.LatestActivityAt = max(commentLatestActivity, postLatestActivity)
			user.Persona.PriorityScore = priorityScore
			user.Persona.LastScannedAt = now.Unix()
			user.Persona.TotalContentCount += numberOfContents
			user.Persona.Thirdparty = personaWithActivities.Persona.Thirdparty
			user.Persona.AverageSpam = newAverageSpam
			user.Persona.TimesScanned += 1
			if commentLatestActivity > postLatestActivity {
				if len(comments.items) > 0 {
					user.Persona.Thirdparty.LatestScannedContentName = comments.items[0].GetName()
				}
			} else {
				if len(posts.items) > 0 {
					user.Persona.Thirdparty.LatestScannedContentName = posts.items[0].GetName()
				}
			}
			suspended := user.Persona.Thirdparty.IsSuspended()

			err = user.Persona.UpdatePersona(
				g.PersonaDB(),
				"latest_activity_at",
				"priority_score",
				"last_scanned_at",
				"first_scanned_at",
				"times_scanned",
				"average_spam",
				"thirdparty",
				"total_content_count",
			)
			if err != nil {
				g.Log.Error().Stack().Err(err).Msgf("`%s` failure average_spam=%f suspended=%t", username, newAverageSpam, suspended)
				return
			}

			if numberOfContents > 0 {
				// we should pass empty ctx when goroutine is started
				// go func() {
				// 	g.AIPQueue.Add(topElement)
				g.RPCClient.Personalize(ctx, contents)
				// 	g.AIPQueue.ZRem(g.AIPQueue.Key, topElement)
				// }()

				// g.Log.Info().Msgf("`%s` added to AI pqueue", username)
			}
			g.Log.Info().Msgf("`%s` success average_spam=%f suspended=%t", username, newAverageSpam, suspended)
		}()
	}
}

type Content[V reddit.PostOrComment] struct {
	items     reddit.ContentSlice[V]
	spamProbs []float32
	// averageSpam is average spam of only the FIRST request content
	averageSpam float32
	// averageSpamCalculatedFor is the number of content for which average spam was calculated
	averageSpamCalculatedFor int
}

// WARNING: not thread safe
func calculatePriorityScore(personaWithActivities *models.PersonaWithActivities) int64 {
	priority := 1

	if personaWithActivities.Persona.IsSuspended() {
		return 0
	}

	if personaWithActivities.Persona.AverageSpam >= constants.SPAM_THRESHOLD {
		return 0
	}

	if personaWithActivities.Persona.IsEmployee() {
		priority += 10
	}

	if personaWithActivities.Persona.IsGold() {
		priority += 20
	}

	if personaWithActivities.Persona.TotalKarma() > 1 {
		priority += 5
	}

	return int64(priority)
}

func buildGetRequest(ctx context.Context, accessToken, url string) *http.Request {
	get, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	get.Header.Add("Authorization", "bearer "+accessToken)
	get.Header.Add("User-Agent", "V9dHxtB00lR5dh00BXJ-4g:v0.0.1 (by /u/BeautifulMandarin123)")
	get.Header.Add("Content-Type", "application/json")

	return get
}

type SpamResponseBody struct {
	Probs []float32 `json:"probs"`
}

type SpamRequestBody struct {
	Sentences []string `json:"sentences"`
}

type Spam struct {
	isSpam      bool
	spamProbs   []float32
	averageSpam float32
}

func freshContentIsSpam(
	ctx context.Context,
	g *god.God,
	sentences []string,
	receiver chan<- bool,
	sender <-chan bool,
) Spam {
	spam := Spam{}
	if len(sentences) > 0 {
		sentences := &pb.Sentences{Items: sentences}
		spamProbs, err := g.RPCClient.CalculateSpamProbs(ctx, sentences)
		if err != nil {
			g.Log.Error().Stack().Err(err).Send()
			return spam
		}

		spam.spamProbs = spamProbs.Items
		spam.averageSpam = average(spam.spamProbs)
		if spam.averageSpam >= constants.SPAM_THRESHOLD {
			spam.isSpam = true
		}
	}

	otherContentIsSpam := <-Rendezvous(receiver, sender)(spam.isSpam)
	spam.isSpam = spam.isSpam || otherContentIsSpam

	return spam
}

// filterNewContent grabs contents that are new
// (their creation time is new, or newer than latestContentName of a user)
func filterNewContent[V reddit.PostOrComment](newContents *[]V, dest *[]V, latestContentName string, latestActivityAt int64) {
	for _, content := range *newContents {
		// INFO: we have reached previously scanned contents' starting point
		if content.GetName() == latestContentName {
			break
		}

		// INFO: just in case if previous fails for some reason
		if content.CreatedAtEpoch() <= latestActivityAt {
			break
		}

		*dest = append(*dest, content)
	}
}

// WARNING: not thread safe
//
// updates weekly and yearly activities
func updateActivities[V reddit.PostOrComment](
	user *models.PersonaWithActivities,
	activeDays map[int][constants.DAYS_IN_A_YEAR]bool,
	contents *reddit.ContentSlice[V],
) {
	for _, content := range *contents {
		utc := time.Unix(content.CreatedAtEpoch(), 0).UTC()

		yearday := utc.YearDay() - 1
		// INFO: because we have a 366 length array, non-leap year yearday index since March
		// would be -1 than the actual
		if !isLeapYear(utc.Year()) && yearday >= constants.NON_LEAP_YEAR_FIRST_MARCH_YEARDAY {
			yearday += 1
		}
		hour := utc.Hour()
		user.YearlyActivity.YearlyActivity[yearday][hour]++

		weekday := utc.Weekday()
		if !activeDays[utc.Year()][yearday] {
			tmp := activeDays[utc.Year()]
			tmp[yearday] = true
			activeDays[utc.Year()] = tmp

			user.WeeklyActivity.Sun2MonCounts[weekday] += 1
		}
	}
}

func isLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || (year%400 == 0)
}

func initActiveDaysMap(startingYear int) map[int][constants.DAYS_IN_A_YEAR]bool {
	// INFO: Imagine user has been scanned before, and somehow we get to scan this user again
	// on current day (on which we scanned them before), there would be a probability bug,
	// because we are going to increase the weekly+yearly activities by 1 again for the same day.
	isActiveDayMap := make(map[int][constants.DAYS_IN_A_YEAR]bool)
	for i := startingYear; i < constants.MAX_YEAR; i++ {
		isActiveDayMap[i] = [constants.DAYS_IN_A_YEAR]bool{}
	}

	return isActiveDayMap
}

func userIsScannedFirstTime(startScanningFromEpoch int64) bool {
	return startScanningFromEpoch == 0
}

func average(list []float32) float32 {
	var _average float32
	if len(list) == 0 {
		return _average
	}

	for _, l := range list {
		_average += l
	}
	_average /= float32(len(list))

	return _average
}

func averageWithLength(l int, elems ...float32) float32 {
	if l == 0 {
		return 0
	}

	var average float32
	for _, elem := range elems {
		average += elem
	}
	average /= float32(l)

	return average
}

func insertPersonas(g *god.God, usernames []string) {
	_, err := models.InsertPersonas(g.PersonaDB(), usernames, models.SOCIAL_NET_REDDIT)
	if err != nil {
		g.Log.Error().Stack().Err(err).Send()
		return
	}
}

func filterNewUsernames(g *god.God, content Content[reddit.Comment]) []string {
	newUsernames := make(map[string]interface{})
	g.Lock()
	for _, comment := range content.items {
		author := comment.Data.LinkAuthor
		if exists := g.Exists(author); author != "[deleted]" && !exists {
			newUsernames[author] = struct{}{}
		}
	}
	g.Unlock()

	var usernames []string
	for username := range newUsernames {
		usernames = append(usernames, username)
	}

	return usernames
}

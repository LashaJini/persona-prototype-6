package god

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wholesome-ghoul/persona-prototype-6/constants"
	"github.com/wholesome-ghoul/persona-prototype-6/logger"
)

type RedditClientI interface {
	DoWithRefreshToken(f func(accessToken string) *http.Request) (*http.Response, error)
}

type RedditClient struct {
	token       Token
	AccessToken string
	Interval    *Interval
	http.Client
}

func NewRedditClient(accessToken string) *RedditClient {
	var intervalC = make(chan time.Duration, 1)
	interval := &Interval{
		Current:   constants.DEFAULT_INTERVAL,
		IntervalC: intervalC,
	}
	client := &RedditClient{
		AccessToken: accessToken,
		Interval:    interval,
	}
	return client
}

type Interval struct {
	Current   time.Duration
	IntervalC chan time.Duration
	sync.Mutex
}

type Token struct {
	refreshed bool
	sync.Mutex
}

func (h *RedditClient) DoWithRefreshToken(buildRequest func(accessToken string) *http.Request) (*http.Response, error) {
	var res *http.Response
	var err error
	l := logger.Log()

	for i := 0; i < constants.MAX_CLIENT_DO_RETRIES; i++ {
		h.token.Lock()
		accessToken := h.AccessToken
		h.token.Unlock()

		res, err = h.Do(buildRequest(accessToken))
		if err != nil {
			return nil, err
		}

		if res != nil {
			h.handleInterval(res, l)

			if tokenExpired(res) {
				h.token.Lock()
				h.token.refreshed = false
				h.token.Unlock()

				if err = h.refreshToken(); err != nil {
					return nil, err
				}
				continue
			}
		}

		return res, nil
	}

	return nil, errors.New("maximum retries exceeded")
}

func tokenExpired(res *http.Response) bool {
	return res.StatusCode == http.StatusUnauthorized
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *RedditClient) refreshToken() error {
	h.token.Lock()
	defer h.token.Unlock()
	// someone might have already refreshed it while we were waiting
	if h.token.refreshed {
		return nil
	}

	l := logger.Log()
	apiURL := "https://www.reddit.com"
	endpoint := "/api/v1/access_token"
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	u, _ := url.ParseRequestURI(apiURL)
	u.Path = endpoint

	post, _ := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(data.Encode()))
	post.SetBasicAuth(CLIENT_ID, CLIENT_SECRET)
	post.Header.Add("User-Agent", "V9dHxtB00lR5dh00BXJ-4g:v0.0.1 (by /u/BeautifulMandarin123)")
	post.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	l.Info().Msg("refreshing Reddit access token")
	res, err := h.Do(post)
	if err != nil {
		return err
	}
	var body AccessTokenResponse
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return err
	}

	h.token.refreshed = true
	h.AccessToken = body.AccessToken
	return nil
}

func (h *RedditClient) handleInterval(res *http.Response, l logger.Logger) {
	ratelimitReset, _ := strconv.Atoi(res.Header.Get("X-Ratelimit-Reset"))
	ratelimitUsed, _ := strconv.Atoi(res.Header.Get("X-Ratelimit-Used"))
	ratelimitRemaining := constants.REDDIT_MAX_REQUESTS - ratelimitUsed

	newInterval := constants.DEFAULT_INTERVAL
	diff := ratelimitRemaining - ratelimitReset
	// we don't want to spam the channel if ratelimitRemaining and ratelimitReset are equal
	if diff >= 2 {
		newInterval /= 2
	}

	h.Interval.Lock()
	currInterval := h.Interval.Current
	if currInterval != newInterval {
		h.Interval.Current = newInterval
	}
	h.Interval.Unlock()

	if currInterval != newInterval {
		h.Interval.IntervalC <- newInterval
	}

	l.Debug().Msgf("Reddit API: code=%d ratelimit_reset=%d ratelimit_remaining=%d interval=%s", res.StatusCode, ratelimitReset, ratelimitRemaining, newInterval)
}

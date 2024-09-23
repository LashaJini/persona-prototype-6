package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wholesome-ghoul/persona-prototype-6/config"
	"github.com/wholesome-ghoul/persona-prototype-6/constants"
	"github.com/wholesome-ghoul/persona-prototype-6/crawler"
	"github.com/wholesome-ghoul/persona-prototype-6/god"
	rpcserver "github.com/wholesome-ghoul/persona-prototype-6/server/rpc"
)

const (
	REDIS_ADDRESS  = "localhost:6379"
	REDIS_PASSWORD = ""
	REDIS_DB       = 0

	REDDIT_QPS = 1

	TOTAL_CRAWLERS = 2
)

var ACCESS_TOKEN = "eyJhbGciOiJSUzI1NiIsImtpZCI6IlNIQTI1NjpzS3dsMnlsV0VtMjVmcXhwTU40cWY4MXE2OWFFdWFyMnpLMUdhVGxjdWNZIiwidHlwIjoiSldUIn0.eyJzdWIiOiJsb2lkIiwiZXhwIjoxNzEwNzU2MzM5LjgxNzQ1OCwiaWF0IjoxNzEwNjY5OTM5LjgxNzQ1OCwianRpIjoicXR6dk5XZU9QckdWSnRlVU1La2ltTTZ2ZlpybXlRIiwiY2lkIjoiVjlkSHh0QjAwbFI1ZGgwMEJYSi00ZyIsImxpZCI6InQyX3dkNWpvcGhndCIsImxjYSI6MTcxMDY2OTkzOTgwNiwic2NwIjoiZUp5S1Z0SlNpZ1VFQUFEX193TnpBU2MiLCJmbG8iOjZ9.nHliCv6obEEVoXSngHtNOMQuleM9X1wDdS_W6x8bobvLo6gOT5zFPNptMrvEsl28dcUi5Wo0tX3pSkAzynZBwCUbhGZfh7_szMbBFSnIYp-ntvrE-x5rTO4E60FOK6BMhu5dlM7tFfFr6K4Dn1bSvnlr64VBL5z15dOxb_Re7outVW8NE4EvzdLKxca-cBVz9GbWJ6uEHmWDQoOAteWHDOBq6A5pQuMLMfn6J2TOHhBzAa_lcVhS2DgaQuQg7QPRPN21uqOHHkW0zMNm7nkVKevqaPnkgt7kmHwkdhK2lBZQnDP3wNdxbPvjErKIkUiBr2aAndNaSOhMutkOJTh1Sg"

func sigtermListener(g *god.God, c chan os.Signal, cancel context.CancelFunc) {
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()

		g.CrawlPQueue.Release()
		g.AIPQueue.Release()

		os.Exit(0)
	}()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal)

	config := &config.Config{
		RedisCrawlPQueueKey: constants.REDIS_CRAWL_PQUEUE_KEY,
		RedisAIPQueueKey:    constants.REDIS_AI_PQUEUE_KEY,
		RedisAddress:        REDIS_ADDRESS,
		RedisPassword:       REDIS_PASSWORD,
		RedisDB:             REDIS_DB,
		DBUser:              constants.DB_USER,
		DBName:              constants.DB_NAME,
		VDBCollectionName:   constants.VDB_COLLECTION_NAME,
	}
	client := god.NewRedditClient(ACCESS_TOKEN)
	g := god.NewGod(ctx, config, client)

	go rpcserver.StartServer(constants.GO_GRPC_SERVER_PORT)
	sigtermListener(g, c, cancel)

	go requestTicker(ctx, g.CanMakeRequest, g.Client.Interval.IntervalC, REDDIT_QPS*time.Second)

	totalCrawlers := TOTAL_CRAWLERS
	usersAvailable := make(chan interface{}, totalCrawlers)
	for id := 0; id < totalCrawlers; id++ {
		go crawler.Crawl(ctx, usersAvailable, id+1, g)
	}
	go g.UsersRefresher(totalCrawlers, usersAvailable)

	// just a helper
	timer := time.NewTicker(10 * time.Second)
	defer timer.Stop()
	for {
		<-timer.C
	}
}

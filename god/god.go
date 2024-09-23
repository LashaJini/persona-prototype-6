package god

import (
	"context"
	"database/sql"

	rpcclient "github.com/wholesome-ghoul/persona-prototype-6/client/rpc"
	"github.com/wholesome-ghoul/persona-prototype-6/config"
	"github.com/wholesome-ghoul/persona-prototype-6/constants"
	"github.com/wholesome-ghoul/persona-prototype-6/logger"
	"github.com/wholesome-ghoul/persona-prototype-6/models"
	safemap "github.com/wholesome-ghoul/persona-prototype-6/safe-map"
	"github.com/wholesome-ghoul/persona-prototype-6/storage/database"
	redisstorage "github.com/wholesome-ghoul/persona-prototype-6/storage/redis-storage"
	"github.com/wholesome-ghoul/persona-prototype-6/storage/vdb"
)

// TODO:
const CLIENT_SECRET = "0TwBYGvVqGfbHUaE4URxI_JO8sGCkg"
const CLIENT_ID = "V9dHxtB00lR5dh00BXJ-4g"

type God struct {
	CrawlPQueue    *redisstorage.Redis
	AIPQueue       *redisstorage.Redis
	redditUsersMap *safemap.SafeMap[*models.PersonaWithActivities]
	personaDB      *database.PersonaDB
	contentVDB     *vdb.ContentVDB
	RPCClient      *rpcclient.Client
	Client         *RedditClient

	Log            logger.Logger
	CanMakeRequest chan struct{}
}

func NewGod(ctx context.Context, config *config.Config, httpClient *RedditClient) *God {
	crawlPQueue := redisstorage.NewRedisClient(ctx, config.RedisCrawlPQueueKey, config)
	AIPQueue := redisstorage.NewRedisClient(ctx, config.RedisAIPQueueKey, config)
	redditUsersMap := safemap.NewSafeMap[*models.PersonaWithActivities]()
	personaDB := database.InitDB(config.DBUser, config.DBName)
	contentVDB := vdb.InitClient(ctx, config.VDBCollectionName)
	rpcClient := rpcclient.NewClient(constants.PYTHON_GRPC_SERVER_PORT)

	var canMakeRequest = make(chan struct{}, 1)
	god := &God{
		CrawlPQueue:    crawlPQueue,
		AIPQueue:       AIPQueue,
		redditUsersMap: redditUsersMap,
		personaDB:      personaDB,
		contentVDB:     contentVDB,
		RPCClient:      rpcClient,
		Client:         httpClient,
		Log:            logger.Log(),
		CanMakeRequest: canMakeRequest,
	}

	return god
}

/* db */

func (g *God) PersonaDB() *sql.DB { return g.personaDB.DB() }

/* map */

func (g *God) Lock()   { g.redditUsersMap.Lock() }
func (g *God) Unlock() { g.redditUsersMap.Unlock() }

// WARN: not thread safe
func (g *God) Get(username string) *models.PersonaWithActivities {
	return g.redditUsersMap.Items[username]
}

// WARN: not thread safe
func (g *God) Exists(username string) bool {
	_, ok := g.redditUsersMap.Items[username]
	return ok
}

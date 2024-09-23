package constants

import "time"

const (
	GO_GRPC_SERVER_PORT     = 50051
	PYTHON_GRPC_SERVER_PORT = 50052

	CONTENT_TYPE_COMMENT              = "comment"
	CONTENT_TYPE_POST                 = "post"
	HOURS_IN_A_DAY                    = 24
	DAYS_IN_A_YEAR                    = 366
	SECONDS_IN_A_DAY                  = 60 * 60 * 24
	USERS_REFRESHER_THRESHOLD         = 10
	DAYS_SINCE_LAST_SCAN              = 2 * SECONDS_IN_A_DAY
	EPOCH_DIFF_THRESHOLD              = SECONDS_IN_A_DAY
	NON_LEAP_YEAR_FIRST_MARCH_YEARDAY = 60
	MAX_YEAR                          = 2034
	PRIORITY_SCORE_THRESHOLD          = 1
	MAX_CLIENT_DO_RETRIES             = 3

	REDDIT_MAX_REQUESTS  = 600
	REDDIT_CONTENT_LIMIT = 100
	SPAM_THRESHOLD       = 0.6
	DEFAULT_INTERVAL     = 1000 * time.Millisecond

	// redis
	REDIS_CRAWL_PQUEUE_KEY = "reddit_user_crawl_priority"
	REDIS_AI_PQUEUE_KEY    = "reddit_user_ai_priority"

	// psql
	DB_USER = "postgres"
	DB_NAME = "persona-prototype-6-dev"

	DEV       = "dev"
	APP_ENV   = "APP_ENV"
	LOG_LEVEL = "LOG_LEVEL"

	// vdb
	VDB_COLLECTION_NAME = "content"
)

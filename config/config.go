package config

type Config struct {
	// redis sorted set
	RedisCrawlPQueueKey string
	RedisAIPQueueKey    string
	RedisAddress        string
	RedisPassword       string
	RedisDB             int

	// psql persona db
	DBUser string
	DBName string

	// vdb
	VDBCollectionName string
}

package main

import "flag"

type Config struct {
	RedisAddress  string
	RedisStoreKey string

	ListenAddress string

	HabitRPGUserID   string
	HabitRPGAPIToken string
}

func LoadConfig() *Config {
	var (
		redisAddress = flag.String("redis-url", "", "Connectionstring to redis server")
		redisKey     = flag.String("redis-key", "habitrpg-tasks", "Key to store the data in")

		listenAddress = flag.String("listen", ":3000", "Address incl. port to have the API listen on")

		habitRPGUserID   = flag.String("habit-user", "", "User-ID from API page in HabitRPG")
		habitRPGAPIToken = flag.String("habit-token", "", "API-Token for that HabitRPG user")
	)

	flag.Parse()

	return &Config{
		RedisAddress:  *redisAddress,
		RedisStoreKey: *redisKey,
		ListenAddress: *listenAddress,

		HabitRPGUserID:   *habitRPGUserID,
		HabitRPGAPIToken: *habitRPGAPIToken,
	}
}

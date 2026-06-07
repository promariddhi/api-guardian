package ratelimiter

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func RedisAllow(key string, ctx context.Context, rdb *redis.Client) bool {
	now := float64(time.Now().Unix())
	script := `
	local key = KEYS[1]
	local capacity = tonumber(ARGV[1]) 
	local fillrate = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])

	local tokens = tonumber(redis.call("HGET", key, "tokens"))
	local lastSeen = tonumber(redis.call("HGET", key, "lastSeen"))

	if not tokens then
		redis.call("HSET", key, "tokens", capacity - 1, "lastSeen", now)
		redis.call("EXPIRE", key, 3600)
		return 1
	end

	local elapsed = now - lastSeen
	tokens = tokens + elapsed * fillrate
	if  tokens > capacity then
		tokens = capacity
	end

	if tokens < 1 then
		redis.call("HSET", key, "tokens", tokens, "lastSeen", now)
		redis.call("EXPIRE", key, 3600)
		return 0
	end

	tokens = tokens - 1
	
	redis.call("HSET", key, "tokens", tokens, "lastSeen", now)
	redis.call("EXPIRE", key, 3600)
	return 1
	`

	allowed, err := rdb.Eval(ctx, script, []string{key}, BUCKET_CAPACITY, BUCKET_FILL_RATE, now).Bool()
	if err != nil {
		log.Println(err)
		return false
	}
	return allowed
}

package ratelimiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func RedisAllow(ip string, ctx context.Context, rdb *redis.Client) bool {
	now := float64(time.Now().Unix())
	key := fmt.Sprintf("ip:%s", ip)
	IPData, _ := rdb.HGetAll(ctx, key).Result()
	tokens := refillTokens(IPData, now)
	if tokens < 1 {
		save(rdb, ctx, key, tokens, now)
		return false
	}
	tokens--
	save(rdb, ctx, key, tokens, now)
	return true
}

func refillTokens(IPData map[string]string, now float64) float64 {
	if len(IPData) == 0 {
		return BUCKET_CAPACITY
	}
	tokens, _ := strconv.ParseFloat(IPData["tokens"], 64)
	lastSeen, _ := strconv.ParseFloat(IPData["lastSeen"], 64)
	elapsed := now - lastSeen
	tokens = min(tokens+elapsed*BUCKET_FILL_RATE, BUCKET_CAPACITY)
	return tokens
}

func save(rdb *redis.Client, ctx context.Context, key string, tokens float64, lastSeen float64) {
	rdb.HSet(ctx, key, "tokens", tokens, "lastSeen", lastSeen)
}

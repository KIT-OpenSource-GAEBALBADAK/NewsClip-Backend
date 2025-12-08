package redis

import (
	"context"
	"newsclip/backend/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
)

// Redis 연결 초기화
func ConnectRedis() {
	Client = redis.NewClient(&redis.Options{
		Addr:     config.GetEnv("REDIS_ADDR"),     // "localhost:6379"
		Password: config.GetEnv("REDIS_PASSWORD"), // "" (보통 로컬은 없음)
		DB:       0,                               // 기본 DB 사용
	})

	_, err := Client.Ping(Ctx).Result()
	if err != nil {
		panic("🔥 Failed to connect to Redis: " + err.Error())
	}
}

// 데이터 저장 (유효시간 포함)
func SetData(key string, value interface{}, duration time.Duration) error {
	return Client.Set(Ctx, key, value, duration).Err()
}

// 데이터 조회
func GetData(key string) (string, error) {
	return Client.Get(Ctx, key).Result()
}

// 데이터 삭제
func DeleteData(key string) error {
	return Client.Del(Ctx, key).Err()
}

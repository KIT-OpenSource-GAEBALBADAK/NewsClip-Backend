package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Env 변수를 가져오는 함수
func GetEnv(key string) string {
	return os.Getenv(key)
}

// .env 파일을 로드하는 함수
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

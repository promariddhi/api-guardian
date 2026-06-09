package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Env struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	JWTSecret string
}

func LoadEnv() Env {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using system environment")
	}

	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		db = 0
	}

	return Env{
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       db,
		JWTSecret:     os.Getenv("JWT_SECRET"),
	}
}

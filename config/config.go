package config

import (
	"os"

	"github.com/joho/godotenv"
)

var SECRETKEY string
var HEADER string
var CHANDAO_HOST string
var CHANDAO_ACCOUNT string
var CHANDAO_PASSWORD string
var AUTH_URL string

func init() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	SECRETKEY = os.Getenv("SECRETKEY")
	if SECRETKEY == "" {
		panic("SECRETKEY is not set")
	}
	HEADER = os.Getenv("HEADER")
	if HEADER == "" {
		panic("HEADER is not set")
	}
	CHANDAO_HOST = os.Getenv("CHANDAO_HOST")
	if CHANDAO_HOST == "" {
		panic("CHANDAO_HOST is not set")
	}
	CHANDAO_ACCOUNT = os.Getenv("CHANDAO_ACCOUNT")
	if CHANDAO_ACCOUNT == "" {
		panic("CHANDAO_ACCOUNT is not set")
	}
	CHANDAO_PASSWORD = os.Getenv("CHANDAO_PASSWORD")
	if CHANDAO_PASSWORD == "" {
		panic("CHANDAO_PASSWORD is not set")
	}
	AUTH_URL = os.Getenv("AUTH_URL")
	if AUTH_URL == "" {
		panic("AUTH_URL is not set")
	}
}

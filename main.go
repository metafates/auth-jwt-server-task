package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/metafates/auth-jwt-server-task/config"
	"github.com/metafates/auth-jwt-server-task/db"
	"github.com/metafates/auth-jwt-server-task/server"
)

func main() {
	ctx := context.TODO()

	cfg, err := config.Load(".")
	if err != nil {
		log.Fatal(err)
	}

	dbClient, err := db.Instance(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatal(err)
	}

	err = server.Start(net.JoinHostPort("", cfg.Port), server.Options{
		RefreshTokenDuration: 7 * 24 * time.Hour,
		AccessTokenDuration:  5 * time.Minute,
		Secret:               []byte(cfg.JWTSecret),
		SigningMethod:        jwt.SigningMethodHS512,
		DB: server.DBOptions{
			Client:   dbClient,
			Database: cfg.MongoDB,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

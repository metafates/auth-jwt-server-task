package server

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/metafates/auth-jwt-server-task/server/openapi"
	"go.mongodb.org/mongo-driver/mongo"
)

type DBOptions struct {
	Client   *mongo.Client
	Database string
}

type Options struct {
	DB DBOptions

	AccessTokenDuration,
	RefreshTokenDuration time.Duration

	Secret        []byte
	SigningMethod jwt.SigningMethod
}

func Start(addr string, options Options) error {
	server := Server{options: options}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	openapi.RegisterHandlers(e, &server)

	return e.Start(addr)
}

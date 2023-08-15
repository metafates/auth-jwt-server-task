package server

import (
	"bytes"
	"context"
	"crypto/sha512"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	echo "github.com/labstack/echo/v4"
	"github.com/metafates/auth-jwt-server-task/model"
	"github.com/metafates/auth-jwt-server-task/server/openapi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ openapi.ServerInterface = (*Server)(nil)

const (
	refreshTokenName = "refresh_token"
	accessTokenName  = "access_token"
)

type signedToken struct {
	String  string
	Expires time.Time
}

type signedTokenPair struct {
	Access, Refresh signedToken
}

type Server struct {
	options Options
}

func (s *Server) db() *mongo.Database {
	return s.options.DB.Client.Database(s.options.DB.Database)
}

func (s *Server) hashToken(token string) []byte {
	// bcrypt can't operate on passwords longer that 72 bytes which makes it
	// unsuitable for JWT tokens
	//
	// https://stackoverflow.com/questions/64860460/store-the-hashed-jwt-token-in-the-database

	hashed := sha512.Sum512([]byte(token))
	return hashed[:]
}

func (s *Server) verifyRefreshToken(GUID string, token string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	usersDB := s.db().Collection("users")

	filter := bson.M{"guid": GUID}

	result := usersDB.FindOne(ctx, filter)

	if err := result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}

		return false, err
	}

	var user model.User
	if err := result.Decode(&user); err != nil {
		return false, err
	}

	userToken := user.RefreshToken
	if userToken == nil {
		return false, nil
	}

	return bytes.Equal([]byte(*userToken), s.hashToken(token)), nil
}

func (s *Server) writeRefreshToken(GUID string, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	usersDB := s.db().Collection("users")

	hashed := s.hashToken(token)

	updateObject := primitive.M{
		refreshTokenName: string(hashed),
	}

	filter := bson.M{"guid": GUID}
	update := bson.D{
		{Key: "$set", Value: updateObject},
	}

	upsert := true
	options := &options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := usersDB.UpdateOne(ctx, filter, update, options)
	return err
}

func (s *Server) generateToken(GUID string, expirationDuration time.Duration) (token signedToken, err error) {
	token.Expires = time.Now().Add(expirationDuration)

	claims := &Claims{
		GUID: GUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(token.Expires),
		},
	}

	token.String, err = jwt.
		NewWithClaims(s.options.SigningMethod, claims).
		SignedString(s.options.Secret)

	return
}

func (s *Server) generateSignedTokenPair(GUID string) (pair signedTokenPair, err error) {
	pair.Access, err = s.generateToken(GUID, s.options.AccessTokenDuration)
	if err != nil {
		return
	}

	pair.Refresh, err = s.generateToken(GUID, s.options.RefreshTokenDuration)
	if err != nil {
		return
	}

	return
}

func (s *Server) sendTokens(ctx echo.Context, pair signedTokenPair) error {
	ctx.SetCookie(&http.Cookie{
		Name:     refreshTokenName,
		Value:    pair.Refresh.String,
		Expires:  pair.Refresh.Expires,
		HttpOnly: true,
	})

	return ctx.JSON(http.StatusOK, struct {
		AccessToken string `json:"access_token"`
		ExpiresAt   int64  `json:"expires_at"`
	}{
		AccessToken: pair.Access.String,
		ExpiresAt:   pair.Access.Expires.Unix(),
	})
}

// PostAuth implements openapi.ServerInterface.
func (s *Server) PostAuth(ctx echo.Context, params openapi.PostAuthParams) error {
	GUID := params.Guid
	if GUID == "" {
		return ctx.NoContent(http.StatusUnauthorized)
	}

	pair, err := s.generateSignedTokenPair(GUID)
	if err != nil {
		log.Print(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if err := s.writeRefreshToken(GUID, pair.Refresh.String); err != nil {
		log.Print(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return s.sendTokens(ctx, pair)
}

// PostRefresh implements openapi.ServerInterface.
func (s *Server) PostRefresh(ctx echo.Context) error {
	cookie, err := ctx.Cookie(refreshTokenName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return ctx.NoContent(http.StatusUnauthorized)
		}

		log.Print(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	var claims Claims
	refreshToken, err := jwt.ParseWithClaims(cookie.Value, &claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != s.options.SigningMethod.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}

		return s.options.Secret, nil
	})

	if !refreshToken.Valid {
		return ctx.NoContent(http.StatusUnauthorized)
	}

	isVerified, err := s.verifyRefreshToken(claims.GUID, cookie.Value)
	if err != nil {
		log.Print(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if !isVerified {
		return ctx.NoContent(http.StatusUnauthorized)
	}

	pair, err := s.generateSignedTokenPair(claims.GUID)
	if err != nil {
		log.Print(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if err := s.writeRefreshToken(claims.GUID, pair.Refresh.String); err != nil {
		log.Print(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return s.sendTokens(ctx, pair)
}

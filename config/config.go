package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Port      string `koanf:"port"`
	JWTSecret string `koanf:"jwt.secret"`
	MongoURI  string `koanf:"mongo.uri"`
	MongoDB   string `koanf:"mongo.db"`
}

func Load(path string) (config Config, err error) {
	err = godotenv.Load(filepath.Join(path, ".env"))

	// ignore if file does not exist
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return
	}

	k := koanf.New(".")

	// Default values
	err = k.Load(confmap.Provider(map[string]any{
		"port": "1234",
	}, "."), nil)

	if err != nil {
		return
	}

	const envPrefix = "SERVER_"
	err = k.Load(env.Provider(envPrefix, ".", func(s string) string {
		return strings.Replace(
			strings.ToLower(strings.TrimPrefix(s, envPrefix)),
			"_",
			".",
			-1,
		)
	}), nil)

	if err != nil {
		return
	}

	err = k.UnmarshalWithConf("", &config, koanf.UnmarshalConf{
		Tag:       "koanf",
		FlatPaths: true,
	})

	if err != nil {
		return
	}

	if config.JWTSecret == "" {
		err = errors.New("jwt secret is empty")
		return
	}

	if config.MongoDB == "" {
		err = errors.New("mongo db is empty")
		return
	}

	if config.MongoURI == "" {
		err = errors.New("mongo uri is empty")
		return
	}

	return
}

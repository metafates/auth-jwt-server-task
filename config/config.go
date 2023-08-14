package config

import (
	"os"
)

type Config struct {
	Port string `mapstructure:"PORT"`

	JWTSecret string `mapstructure:"JWT_SECRET"`

	MongoURI string `mapstructure:"MONGO_URI"`
	MongoDB  string `mapstructure:"MONGO_DB"`
}

func Load(path string) (config Config, err error) {
	config.Port = os.Getenv("PORT")
	config.JWTSecret = os.Getenv("JWT_SECRET")
	config.MongoURI = os.Getenv("MONGO_URI")
	config.MongoDB = os.Getenv("MONGO_DB")
	return
	// viper.AddConfigPath(path)
	// viper.SetConfigName("app")
	// viper.SetConfigType("env")

	// viper.AutomaticEnv()

	// if err = viper.ReadInConfig(); err != nil {
	// 	if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
	// 		return
	// 	}
	// }

	// err = viper.Unmarshal(&config)
	// return
}

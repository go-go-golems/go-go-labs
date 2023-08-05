package pkg

import (
	"fmt"
	"github.com/spf13/viper"
	"path/filepath"

	"github.com/mattn/go-mastodon"
)

func InitConfig() {
	viper.SetConfigName("config")                           // config file name without extension
	viper.SetConfigType("yaml")                             // or viper.SetConfigType("YAML")
	viper.AddConfigPath(filepath.Join("$HOME", ".mastoid")) // path to look for the config file in
	_ = viper.ReadInConfig()                                // read in config file
}

func StoreCredentials(credentials *Credentials) error {
	viper.Set("client_id", credentials.Application.ClientID)
	viper.Set("client_secret", credentials.Application.ClientSecret)
	viper.Set("auth_uri", credentials.Application.AuthURI)
	viper.Set("redirect_uri", credentials.Application.RedirectURI)
	viper.Set("grant_token", credentials.GrantToken)
	viper.Set("access_token", credentials.AccessToken)
	viper.Set("server", credentials.Server)

	err := viper.WriteConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// config file not found, create it
		err = viper.SafeWriteConfig()
		if err != nil {
			return fmt.Errorf("Error writing config: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error writing config: %w", err)
	}
	return nil
}

func LoadCredentials() (*Credentials, error) {
	clientId := viper.GetString("client_id")
	clientSecret := viper.GetString("client_secret")
	authUri := viper.GetString("auth_uri")
	redirectUri := viper.GetString("redirect_uri")
	grantToken := viper.GetString("grant_token")
	accessToken := viper.GetString("access_token")
	server := viper.GetString("server")

	// check that they are valid
	if clientId == "" || clientSecret == "" {
		return nil, fmt.Errorf("no credentials found")
	}

	app := &Credentials{
		Server:      server,
		GrantToken:  grantToken,
		AccessToken: accessToken,
		Application: &mastodon.Application{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			AuthURI:      authUri,
			RedirectURI:  redirectUri,
		},
	}
	return app, nil
}

// Credentials stores the result of an oauth flow.
type Credentials struct {
	Server      string
	GrantToken  string
	Application *mastodon.Application
	AccessToken string
}

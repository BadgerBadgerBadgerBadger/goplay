package main

type Config struct {
	Slack    SlackConfig    `json:"slack"`
	Host     string         `json:"host"`
	Database DatabaseConfig `json:"database"`
}

type SlackConfig struct {
	Oauth OauthConfig `json:"oauth"`
}

type OauthConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type DatabaseConfig struct {
	Path string `json:"path"`
}

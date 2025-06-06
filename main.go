package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	"gopkg.in/yaml.v3"
)

type Config struct {
	URL              string `yaml:"url"`
	Realm            string `yaml:"realm"`
	ClientID         string `yaml:"client_id"`
	ClientSecret     string `yaml:"client_secret"`
	Attribute        string `yaml:"attribute"`
	Debug            bool   `yaml:"debug"`
	IgnoreDisabled   bool   `yaml:"ignore_disabled"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	// Default to true if unset (zero value)
	if !cfg.IgnoreDisabled {
		cfg.IgnoreDisabled = true
	}
	return &cfg, nil
}

func main() {
	// Read username from command-line argument
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: keycloak-ssh-auth <username>")
		os.Exit(1)
	}
	loginUser := os.Args[1]

	// Load config from default path
	configPath := "/etc/keycloak-ssh-auth/config.yaml"
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client := gocloak.NewClient(cfg.URL)
	ctx := context.Background()

	token, err := client.LoginClient(ctx, cfg.ClientID, cfg.ClientSecret, cfg.Realm)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	if cfg.Debug {
		log.Printf("Access token retrieved")
	}

	users, err := client.GetUsers(ctx, token.AccessToken, cfg.Realm, gocloak.GetUsersParams{
		Username: gocloak.StringP(loginUser),
	})
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}
	if len(users) == 0 {
		log.Fatalf("No such user: %s", loginUser)
	}

	user := users[0]

	if cfg.IgnoreDisabled && (user.Enabled != nil && !*user.Enabled) {
		if cfg.Debug {
			log.Printf("User %s is disabled", loginUser)
		}
		os.Exit(0) // Exit silently without printing any keys
	}

	if user.Attributes == nil {
		log.Fatalf("No attributes found for user %s", loginUser)
	}

	attr, ok := (*user.Attributes)[cfg.Attribute]
	if !ok {
		log.Fatalf("Attribute %s not found", cfg.Attribute)
	}

	var keys []string
	for _, val := range attr {
		for _, line := range strings.Split(val, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				keys = append(keys, line)
			}
		}
	}

	if len(keys) == 0 {
		log.Fatalf("No keys found for user %s", loginUser)
	}

	for _, key := range keys {
		fmt.Println(key)
	}
}

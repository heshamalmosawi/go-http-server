package modules

import (
	"encoding/json"
	"log"
	"os"
)

type ServerConfig struct {
	Host  string `json:"host"`
	Ports []uint  `json:"ports"`
}

type DefaultServerConfig struct {
	ServerName      string            `json:"server_name"`
	ErrorPages      map[string]string `json:"error_pages"`
	ClientBodyLimit uint              `json:"client_body_limit"`
}

type RouteConfig struct {
	Root             string   `json:"root,omitempty"`
	Methods          []string `json:"methods,omitempty"`
	Redirect         string   `json:"redirect,omitempty"`
	DefaultFile      string   `json:"default_file,omitempty"`
	DirectoryListing bool     `json:"directory_listing,omitempty"`
	ClientBodyLimit  uint     `json:"client_body_limit,omitempty"`
}

type Config struct {
	Server        ServerConfig           `json:"server"`
	DefaultServer DefaultServerConfig    `json:"default_server"`
	Routes        map[string]RouteConfig `json:"routes"`
}

func LoadConfig(path string) Config {
	var config Config
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("Failed to read config file: ", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatal("Invalid config JSON format: ", err)
	}

	return config
}

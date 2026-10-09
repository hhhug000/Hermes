package main

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	Host        string
	Port        string
	DownloadDir string
}

func LoadConfig() Config {
	cfg := Config{
		Host:        "localhost",
		Port:        "2222",
		DownloadDir: ".",
	}

	file, err := os.Open("hermes.conf")
	if err != nil {
		return cfg
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch strings.ToLower(key) {
		case "host":
			cfg.Host = val
		case "port":
			cfg.Port = val
		case "download_dir":
			cfg.DownloadDir = val
		}
	}
	return cfg
}

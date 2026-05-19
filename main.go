package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	SourcePath   string
	MoviesPath   string
	TVShowPath   string
	ScanInterval time.Duration
	DryRun       bool
}

func main() {
	cfg := loadConfig()

	log.Printf("[MediaSorter] Starting")
	log.Printf("[MediaSorter] Source:        %s", cfg.SourcePath)
	log.Printf("[MediaSorter] TV Shows:      %s", cfg.TVShowPath)
	log.Printf("[MediaSorter] Movies:        %s", cfg.MoviesPath)
	log.Printf("[MediaSorter] Scan interval: %s", cfg.ScanInterval)
	if cfg.DryRun {
		log.Println("[MediaSorter] DRY RUN — no files will be moved")
	}

	sorter := NewSorter(cfg)

	// Run immediately on start then on every tick
	sorter.Scan()

	ticker := time.NewTicker(cfg.ScanInterval)
	defer ticker.Stop()
	for range ticker.C {
		sorter.Scan()
	}
}

func loadConfig() Config {
	return Config{
		SourcePath:   requireEnv("SOURCE_PATH"),
		MoviesPath:   requireEnv("MOVIES_PATH"),
		TVShowPath:   requireEnv("TVSHOW_PATH"),
		ScanInterval: envDuration("SCAN_INTERVAL", 300),
		DryRun:       os.Getenv("DRY_RUN") != "",
	}
}

func requireEnv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		log.Fatalf("[Config] Required environment variable %s is not set", key)
	}
	return v
}

func envDuration(key string, defaultSecs int) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return time.Duration(defaultSecs) * time.Second
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("[Config] Invalid %s=%q, using default %ds", key, v, defaultSecs)
		return time.Duration(defaultSecs) * time.Second
	}
	return time.Duration(n) * time.Second
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp-education/terraminogo/internal/highscore"
	"github.com/hashicorp-education/terraminogo/internal/hvs_client"
)

type TerraminoServer struct {
	highScoreManager *highscore.Manager
}

func main() {
	ctx := context.Background()

	// Get application name
	appName, envExists := os.LookupEnv("APP_NAME")
	if !envExists {
		appName = "terramino"
	}

	// Initialize high score manager
	server := &TerraminoServer{
		highScoreManager: highscore.NewManager(ctx, appName),
	}

	// Configure Redis connection
	redisHost, hasRedisHost := os.LookupEnv("REDIS_HOST")
	redisPort, hasRedisPort := os.LookupEnv("REDIS_PORT")

	if hasRedisHost && hasRedisPort {
		// Use direct Redis connection
		server.highScoreManager.ConfigureRedis(
			redisHost,
			redisPort,
			os.Getenv("REDIS_PASSWORD"),
		)
	} else {
		// Use HVS for Redis configuration
		server.highScoreManager.ConfigureHVS(hvs_client.NewHVSClient())
	}

	// Set up HTTP routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Terramino - HashiCorp Demo App\nhttps://developer.hashicorp.com/\n"))
	})
	http.HandleFunc("/env", envHandler)
	http.HandleFunc("/redis", server.redisHandler)
	http.HandleFunc("/score", server.highScoreManager.HandleHTTP)

	// Start server
	envPort, envPortExists := os.LookupEnv("TERRAMINO_PORT")
	if !envPortExists {
		envPort = "8080"
	}
	port := fmt.Sprintf(":%s", envPort)
	fmt.Printf("Terramino server is running on http://localhost%s\n", port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// DEBUG: Print all runtime environment variables that start with "HCP_"
func envHandler(w http.ResponseWriter, r *http.Request) {
	out := ""
	for _, e := range os.Environ() {
		// Split the environment variable into key and value
		pair := strings.SplitN(e, "=", 2)
		if strings.HasPrefix(pair[0], "HCP_") {
			out += fmt.Sprintf("%s\n", e)
		}
	}

	out += fmt.Sprintf("APP_NAME=%s\n", os.Getenv("APP_NAME"))

	w.Write([]byte(out))
}

func (s *TerraminoServer) redisHandler(w http.ResponseWriter, r *http.Request) {
	host, port, status := s.highScoreManager.GetRedisInfo()
	fmt.Fprintf(w, "redis_host=%s\nredis_port=%s\n\nConnection: %s", host, port, status)
}

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"

	terraminogo "github.com/hashicorp-education/terraminogo/internal"
	"github.com/redis/go-redis/v9"
)

type TerraminoData struct {
	HVSClient   *terraminogo.HVSClient
	redisClient *redis.Client
	ctx         context.Context
	appName     string
}

func main() {
	t := &TerraminoData{}
	t.HVSClient = terraminogo.NewHVSClient()
	t.redisClient = nil
	t.ctx = context.Background()

	appName, envExists := os.LookupEnv("APP_NAME")
	if !envExists {
		appName = "terramino"
	}
	t.appName = appName

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/env", envHandler)
	http.HandleFunc("/score", t.highScoreHandler)
	http.HandleFunc("/redis", t.redisHandler)
	http.HandleFunc("/{path}", pathHandler)

	port := ":8080"
	fmt.Printf("Terramino server is running on http://localhost%s\n", port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Parse and serve index template
func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("web/index.html")
	if err != nil {
		log.Fatal(err)
	}

	err = t.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Handle non-template files
func pathHandler(w http.ResponseWriter, r *http.Request) {
	filePath, err := fileLookup(r.PathValue("path"))
	if err != nil {
		// User requested a file that does not exist
		// Return 404
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(404)
			return
		} else {
			// Unknown error
			log.Fatal(err)
		}
	}

	http.ServeFile(w, r, filePath)
}

func (t *TerraminoData) highScoreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		score := t.GetHighScore()
		w.Write([]byte(strconv.Itoa(score)))
	} else if r.Method == "POST" {
		newScore, _ := io.ReadAll(r.Body)
		iNewScore, _ := strconv.Atoi(string(newScore))
		iOldScore := t.GetHighScore()
		if iNewScore > iOldScore {
			t.SetHighScore(iNewScore)
			w.Write(newScore)
		} else {
			w.Write([]byte(strconv.Itoa(iOldScore)))
		}
	}
}

func (t *TerraminoData) getRedisClient() *redis.Client {
	if t.redisClient != nil {
		// We have an existing connection, make sure it's still valid
		pingResp := t.redisClient.Ping(t.ctx)
		if pingResp.Err() == nil {
			// Connection is valid, return client
			return t.redisClient
		}
	}

	// Either we don't have a connection, or it's no longer valid
	// Create a new client

	// Check for connection info in HVS
	redisIP, err := t.HVSClient.GetSecret(t.appName, "redis_ip")
	if err != nil {
		// No Redis server is available
		t.redisClient = nil
		return nil
	}
	redisPort, _ := t.HVSClient.GetSecret(t.appName, "redis_port")
	redisPassword, _ := t.HVSClient.GetSecret(t.appName, "redis_password")
	t.redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisIP, redisPort),
		Password: redisPassword,
		DB:       0,
	})

	// Check connection
	pingResp := t.redisClient.Ping(t.ctx)
	if pingResp.Err() != nil {
		// Error connecting to the server
		log.Println(pingResp.Err())
		return nil
	}

	return t.redisClient
}

func (t *TerraminoData) GetHighScore() int {
	redisClient := t.getRedisClient()
	if redisClient != nil {
		val, err := redisClient.Get(t.ctx, "score").Result()
		if err == nil {
			iVal, _ := strconv.Atoi(val)
			return iVal
		}
	}

	return 0
}

func (t *TerraminoData) SetHighScore(score int) {
	redisClient := t.getRedisClient()
	if redisClient != nil {
		redisClient.Set(t.ctx, "score", score, 0)
	}
}

// Lookup requested file, return an error if it
// does not exist
func fileLookup(file string) (string, error) {
	fullPath := fmt.Sprintf("web/%s", file)
	_, err := os.Stat(fullPath)

	if err != nil {
		return "", err
	} else {
		return fullPath, nil
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

	w.Write([]byte(out))
}

func (t *TerraminoData) redisHandler(w http.ResponseWriter, r *http.Request) {
	redisHost, _ := t.HVSClient.GetSecret(t.appName, "redis_ip")
	redisPort, _ := t.HVSClient.GetSecret(t.appName, "redis_port")

	redisPing := "No connection"
	redisClient := t.getRedisClient()
	if redisClient != nil {
		pingResp := redisClient.Ping(t.ctx)
		redisPing = pingResp.String()
	}

	fmt.Fprintf(w, "redis_host=%s\nredis_port=%s\n\nConnection: %s", redisHost, redisPort, redisPing)
}

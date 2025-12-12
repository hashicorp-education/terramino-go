package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"

	terraminogo "github.com/hashicorp-education/terraminogo/internal"
	"github.com/redis/go-redis/v9"
)

type TerraminoData struct {
	redisClient *redis.Client
	ctx         context.Context
	appName     string
}

func main() {
	// slog.SetLogLoggerLevel(slog.LevelDebug) // Uncomment to enable debug logging

	t := &TerraminoData{}
	t.redisClient = nil
	t.ctx = context.Background()

	appName, envExists := os.LookupEnv("APP_NAME")
	if !envExists {
		appName = "terramino-go"
	}
	t.appName = appName

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/env", envHandler)
	http.HandleFunc("/score", t.highScoreHandler)
	http.HandleFunc("/redis", t.redisHandler)
	http.HandleFunc("/{path}", pathHandler)

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

// Parse and serve index template
func indexHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("REQ: /")
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
	slog.Debug("REQ: File", "path", r.PathValue("path"))
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
		slog.Debug("REQ: GET /highscore")
		score := t.GetHighScore()
		w.Write([]byte(strconv.Itoa(score)))
	} else if r.Method == "POST" {
		slog.Debug("REQ: POST /highscore")
		newScore, _ := io.ReadAll(r.Body)
		iNewScore, _ := strconv.Atoi(string(newScore))
		iOldScore := t.GetHighScore()
		if iNewScore > iOldScore {
			t.SetHighScore(iNewScore)
			w.Write(newScore)
		} else {
			w.Write([]byte(strconv.Itoa(iOldScore)))
		}
	} else if r.Method == "PUT" {
		slog.Debug("REQ: PUT /highscore")
		newScore, _ := io.ReadAll(r.Body)
		iNewScore, _ := strconv.Atoi(string(newScore))
		t.SetHighScore(iNewScore)
		w.Write(newScore)
	}
}

func (t *TerraminoData) getRedisClient() *redis.Client {
	if t.redisClient != nil {
		// We have an existing connection, make sure it's still valid
		pingResp := t.redisClient.Ping(t.ctx)
		if pingResp.Err() == nil {
			// Connection is valid, return client
			return t.redisClient
		} else {
			slog.Error("Could not ping Redis, connection lost?")
		}
	}

	// Either we don't have a connection, or it's no longer valid
	// Create a new client

	// Check for connection info in HVS
	slog.Debug("Getting Redis credentials from SSM", "secret name", t.appName+"-redis")
	creds, err := terraminogo.GetSecret(t.appName + "-redis")
	if err != nil {
		// No Redis server is available
		t.redisClient = nil
		slog.Error("Could not get Redis credentials from SSM", "error", err)
		return nil
	}
	slog.Debug("Got Redis credentials", "ip", creds.IP)
	redisIP := creds.IP
	redisPort := creds.Port
	redisPassword := creds.Password
	t.redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisIP, redisPort),
		Password: redisPassword,
		DB:       0,
	})

	// Check connection
	pingResp := t.redisClient.Ping(t.ctx)
	if pingResp.Err() != nil {
		// Error connecting to the server
		slog.Error("Could not ping Redis after getting credentials", "error", pingResp.Err())
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
	slog.Debug("Set high score", "score", score)
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
	slog.Debug("GET /env")
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

func (t *TerraminoData) redisHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("GET /redis")
	redisHost := ""
	redisPort := ""

	creds, err := terraminogo.GetSecret(t.appName + "-redis")
	if err == nil {
		redisHost = creds.IP
		redisPort = creds.Port
	}

	redisPing := "No connection"
	redisClient := t.getRedisClient()
	if redisClient != nil {
		pingResp := redisClient.Ping(t.ctx)
		redisPing = pingResp.String()
	}

	fmt.Fprintf(w, "redis_host=%s\nredis_port=%s\n\nConnection: %s", redisHost, redisPort, redisPing)
}

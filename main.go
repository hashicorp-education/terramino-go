package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"

	terraminogo "github.com/brianmmcclain/terraminogo/internal"
	"github.com/redis/go-redis/v9"
)

type TerraminoData struct {
	HVSClient   *terraminogo.HVSClient
	RedisClient *redis.Client
}

func main() {
	t := &TerraminoData{}
	t.HVSClient = terraminogo.NewHVSClient()
	t.RedisClient = nil

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/env", envHandler)
	http.HandleFunc("/score", t.highScoreHandler)
	http.HandleFunc("/{path}", pathHandler)

	err := http.ListenAndServe(":8080", nil)
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
	if t.RedisClient == nil {
		// We haven't connected to Redis yet, see if we have one available
		redisHost, err := t.HVSClient.GetSecret("terramino", "redis_host")
		if err != nil {
			// No host defined, return an error
			w.WriteHeader(500)
		}
		// Otherwise, we should have one available, get the rest of the connection info
		redisPort, _ := t.HVSClient.GetSecret("terramino", "redis_port")
		t.RedisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
			Password: "",
			DB:       0,
		})
	}

	ctx := context.Background()

	if r.Method == "GET" {
		val, err := t.RedisClient.Get(ctx, "score").Result()
		if err != nil {
			if err == redis.Nil {
				// Key does not exist, return 0
				val = "0"
			} else {
				log.Fatal(err)
			}
		}
		w.Write([]byte(val))
	} else if r.Method == "POST" {
		// Read the body
		newScore, _ := io.ReadAll(r.Body)
		// TODO: Get the current high score and see if it's higher
		t.RedisClient.Set(ctx, "score", newScore, 0)
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

// DEBUG: Print all runtime environment variables
func envHandler(w http.ResponseWriter, r *http.Request) {
	out := ""
	for _, e := range os.Environ() {
		out += fmt.Sprintf("%s\n", e)
	}

	w.Write([]byte(out))
}

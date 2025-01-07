package highscore

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp-education/terraminogo/internal/hvs_client"
	"github.com/redis/go-redis/v9"
)

// Manager handles high score operations
type Manager struct {
	HVSClient      *hvs_client.HVSClient
	redisClient    *redis.Client
	ctx            context.Context
	appName        string
	useDirectRedis bool
	redisHost      string
	redisPort      string
	redisPassword  string
}

// NewManager creates a new high score manager
func NewManager(ctx context.Context, appName string) *Manager {
	return &Manager{
		ctx:     ctx,
		appName: appName,
	}
}

// ConfigureRedis sets up Redis connection details
func (m *Manager) ConfigureRedis(host, port, password string) {
	m.useDirectRedis = true
	m.redisHost = host
	m.redisPort = port
	m.redisPassword = password
	m.HVSClient = nil
}

// ConfigureHVS sets up HVS client for Redis connection details
func (m *Manager) ConfigureHVS(hvsClient *hvs_client.HVSClient) {
	m.useDirectRedis = false
	m.HVSClient = hvsClient
}

func (m *Manager) getRedisClient() *redis.Client {
	if m.redisClient != nil {
		// We have an existing connection, make sure it's still valid
		pingResp := m.redisClient.Ping(m.ctx)
		if pingResp.Err() == nil {
			// Connection is valid, return client
			return m.redisClient
		}
	}

	// Either we don't have a connection, or it's no longer valid
	// Create a new client
	var redisIP, redisPort, redisPassword string
	var err error

	if m.useDirectRedis {
		redisIP = m.redisHost
		redisPort = m.redisPort
		redisPassword = m.redisPassword
	} else {
		// Check for connection info in HVS
		redisIP, err = m.HVSClient.GetSecret(m.appName, "redis_ip")
		if err != nil {
			// No Redis server is available
			m.redisClient = nil
			return nil
		}
		redisPort, _ = m.HVSClient.GetSecret(m.appName, "redis_port")
		redisPassword, _ = m.HVSClient.GetSecret(m.appName, "redis_password")
	}

	m.redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisIP, redisPort),
		Password: redisPassword,
		DB:       0,
	})

	// Check connection
	pingResp := m.redisClient.Ping(m.ctx)
	if pingResp.Err() != nil {
		// Error connecting to the server
		log.Println(pingResp.Err())
		return nil
	}

	return m.redisClient
}

// GetScore retrieves the current high score
func (m *Manager) GetScore() int {
	redisClient := m.getRedisClient()
	if redisClient != nil {
		val, err := redisClient.Get(m.ctx, "score").Result()
		if err == nil {
			iVal, _ := strconv.Atoi(val)
			return iVal
		}
	}
	return 0
}

// SetScore updates the high score if the new score is higher
func (m *Manager) SetScore(score int) {
	redisClient := m.getRedisClient()
	if redisClient != nil {
		redisClient.Set(m.ctx, "score", score, 0)
	}
}

// HandleHTTP handles HTTP requests for high scores
func (m *Manager) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		score := m.GetScore()
		w.Write([]byte(strconv.Itoa(score)))
	case http.MethodPost:
		newScore, _ := io.ReadAll(r.Body)
		iNewScore, _ := strconv.Atoi(string(newScore))
		iOldScore := m.GetScore()
		if iNewScore > iOldScore {
			m.SetScore(iNewScore)
			w.Write(newScore)
		} else {
			w.Write([]byte(strconv.Itoa(iOldScore)))
		}
	case http.MethodPut:
		newScore, _ := io.ReadAll(r.Body)
		iNewScore, _ := strconv.Atoi(string(newScore))
		m.SetScore(iNewScore)
		w.Write(newScore)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// GetRedisInfo returns Redis connection information
func (m *Manager) GetRedisInfo() (host, port, status string) {
	if m.useDirectRedis {
		host = m.redisHost
		port = m.redisPort
	} else if m.HVSClient != nil {
		host, _ = m.HVSClient.GetSecret(m.appName, "redis_ip")
		port, _ = m.HVSClient.GetSecret(m.appName, "redis_port")
	}

	status = "No connection"
	redisClient := m.getRedisClient()
	if redisClient != nil {
		pingResp := redisClient.Ping(m.ctx)
		status = pingResp.String()
	}

	return host, port, status
}

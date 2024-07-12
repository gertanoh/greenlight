package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/henrtytanoh/greenlight/internal/data"
	jsonlog "github.com/henrtytanoh/greenlight/internal/jsonLog"
	"github.com/henrtytanoh/greenlight/internal/mailer"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}

	limiter struct {
		windowLength int
		requestLimit int
		enabled      bool
	}

	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}

	cors struct {
		trustedOrigins []string
	}

	redis struct {
		dsn string
	}
}

// prometheus metrics config
type metrics struct {
	totalRequestReceived prometheus.Counter
	requestDuration      prometheus.Gauge
	totalResponsesSent   prometheus.Counter
	totalInternalServer  prometheus.Counter
	totalClientSideError prometheus.Counter
}
type application struct {
	config      config
	logger      *jsonlog.Logger
	models      data.Models
	mailer      mailer.Mailer
	wg          sync.WaitGroup
	m           *metrics
	redisClient *redis.Client
}

var (
	buildTime string
	version   string
)

func main() {

	var cfg config

	flag.IntVar(&cfg.port, "port", getPortFromEnv(), "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.StringVar(&cfg.redis.dsn, "redis-dsn", "", "REDIS DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.IntVar(&cfg.limiter.windowLength, "window-length", 1, "Length of window")
	flag.IntVar(&cfg.limiter.requestLimit, "request-limit", 10, "Maxmium request per window length")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.henrygtanoh.net>", "SMTP sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	displayVersion := flag.Bool("version", false, "Display version and exit")
	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	logger.PrintInfo("db dsn ", map[string]string{
		"db": cfg.db.dsn,
	})
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	logger.PrintInfo("redis dsn ", map[string]string{"redis": cfg.redis.dsn})
	redis, err := setupRedis(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer redis.Close()
	logger.PrintInfo("Redis connection established", nil)

	app := &application{
		config:      cfg,
		logger:      logger,
		models:      data.NewModels(db),
		mailer:      mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		m:           NewMetrics(),
		redisClient: redis,
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)

	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func NewMetrics() *metrics {
	m := &metrics{
		totalRequestReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "greenlight_total_request_received",
			Help: "The total number of request received",
		}),
		requestDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "greenlight_request_duration_seconds",
			Help: "The duration of each request in seconds",
		}),
		totalResponsesSent: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "greenlight_total_responses_sent",
			Help: "The total number of responses sent",
		}),
		totalInternalServer: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "greenlight_total_internal_server_error",
			Help: "The total number of internal server error",
		}),
		totalClientSideError: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "greenlight_total_client_side_error",
			Help: "The total number of client side error",
		}),
	}

	prometheus.MustRegister(m.totalRequestReceived)
	prometheus.MustRegister(m.requestDuration)
	prometheus.MustRegister(m.totalResponsesSent)
	prometheus.MustRegister(m.totalInternalServer)
	prometheus.MustRegister(m.totalClientSideError)
	return m
}

func setupRedis(cfg config) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.redis.dsn)
	if err != nil {
		return nil, err
	}
	redis := redis.NewClient(opts)
	return redis, nil
}

func getPortFromEnv() int {
	// Docker Compose will automatically set PORT for your service replicas.
	portStr := os.Getenv("PORT")
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err == nil {
			return port
		}
	}

	fmt.Println("warning: using default port value")
	// Default port if not provided in environment variable
	return 4000
}

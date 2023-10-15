package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/henrtytanoh/greenlight/internal/data"
	jsonlog "github.com/henrtytanoh/greenlight/internal/jsonLog"
	"github.com/henrtytanoh/greenlight/internal/mailer"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
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
		rps     float64
		burst   int
		enabled bool
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
}

// prometheus metrics config
type metrics struct {
	totalRequestReceived prometheus.Counter
	requestDuration      prometheus.Gauge
	totalResponsesSent   prometheus.Counter
}
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
	m      *metrics
}

var (
	buildTime string
	version   string
)

func main() {

	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "b08f06550b8060", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "f6e834f9eb7b99", "SMTP password")
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

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)

	// Metrics

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		m:      NewMetrics(),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	fmt.Println("db_dsn :", cfg.db.dsn)
	fmt.Println("err :", err)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	fmt.Println("err time_duration:", err)

	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	fmt.Println("err ping_context:", err)
	if err != nil {
		return nil, err
	}

	fmt.Println("open works fine")
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
	}

	prometheus.MustRegister(m.totalRequestReceived)
	prometheus.MustRegister(m.requestDuration)
	prometheus.MustRegister(m.totalResponsesSent)
	return m
}

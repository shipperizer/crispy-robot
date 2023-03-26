package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shipperizer/crispy-robot/pkg/echo"
	"github.com/shipperizer/crispy-robot/pkg/search"
	"github.com/shipperizer/crispy-robot/pkg/watcher"

	bleve "github.com/blevesearch/bleve/v2"
	"github.com/kelseyhightower/envconfig"
	redisotel "github.com/redis/go-redis/extra/redisotel/v9"
	redis "github.com/redis/go-redis/v9"
	"github.com/shipperizer/miniature-monkey/v2/config"
	"github.com/shipperizer/miniature-monkey/v2/core"
	monConfig "github.com/shipperizer/miniature-monkey/v2/monitoring/config"
	monCore "github.com/shipperizer/miniature-monkey/v2/monitoring/core"
	"github.com/shipperizer/miniature-monkey/v2/tracing"
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type EnvSpec struct {
	Port            string   `envconfig:"http_port" default:"8000"`
	TracingEndpoint string   `envconfig:"tracing_endpoint" default:"http://jaeger.svc.cluster.local:14268/api/traces"`
	RedisEndpoint   string   `envconfig:"redis_endpoint" default:"redis.svc.cluster.local:6379"`
	RedisPassword   string   `envconfig:"redis_password"` // TODO @shipperizer just getting going, need to remove for any real docs
	KeyPrefix       string   `envconfig:"key_prefix" default:"test"`
	ETCDEndpoints   []string `envconfig:"etcd_endpoints" default:"localhost:2379,localhost:2380"`
	ETCDPassword    string   `envconfig:"etcd_password"` // TODO @shipperizer just getting going, need to remove for any real docs
}

func main() {
	logger, err := zap.NewDevelopment()
	defer logger.Sync()

	if err != nil {
		panic(err.Error())
	}

	var specs EnvSpec
	err = envconfig.Process("", &specs)

	if err != nil {
		logger.Sugar().Fatal(err.Error())
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     specs.RedisEndpoint, // use default Addr
		Password: specs.RedisPassword,
		DB:       0, // use default DB
	})

	if err := redisotel.InstrumentTracing(rdb); err != nil {
		panic(err)
	}

	// Enable metrics instrumentation.
	if err := redisotel.InstrumentMetrics(rdb); err != nil {
		panic(err)
	}

	monitor := monCore.NewMonitor(
		monConfig.NewMonitorConfig("web", nil, logger.Sugar()),
	)
	tracer := tracing.NewTracer(
		tracing.NewTracerConfig("web", specs.TracingEndpoint, logger.Sugar()),
	)

	etcdClient, err := etcd.New(
		etcd.Config{
			Endpoints:   specs.ETCDEndpoints,
			Logger:      logger,
			DialTimeout: 20 * time.Second,
			Username:    "root",
			Password:    specs.ETCDPassword,
		},
	)

	if err != nil {
		logger.Sugar().Fatal(err)
	}
	// bleveIndex, err := bleve.New("bleve", bleve.NewIndexMapping())
	bleveIndex, err := bleve.NewMemOnly(bleve.NewIndexMapping())
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	apiCfg := config.NewAPIConfig(
		"web",
		nil,
		// nil,
		tracer,
		monitor,
		logger.Sugar(),
	)

	api := core.NewAPI(apiCfg)

	api.RegisterBlueprints(
		api.Router(),
		echo.NewBlueprint(
			echo.NewService(
				echo.NewStore(rdb, tracer),
				tracer,
			),
			tracer,
		),
		search.NewBlueprint(
			specs.KeyPrefix,
			bleveIndex,
			etcdClient,
			tracer,
			logger.Sugar(),
		),
	)

	_ = watcher.NewWatcher(specs.KeyPrefix, bleveIndex, etcdClient, tracer, logger.Sugar())
	_ = watcher.NewScanner(specs.KeyPrefix, bleveIndex, etcdClient, tracer, logger.Sugar())

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", specs.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      api.Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Sugar().Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	etcdClient.Close()
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	logger.Sugar().Info("Shutting down")
	os.Exit(0)
}

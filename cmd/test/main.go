package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/kelseyhightower/envconfig"
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type EnvSpec struct {
	KeyPrefix     string   `envconfig:"key_prefix" default:"test"`
	ETCDEndpoints []string `envconfig:"etcd_endpoints" default:"localhost:2379,localhost:2380"`
	ETCDPassword  string   `envconfig:"etcd_password"` // TODO @shipperizer just getting going, need to remove for any real docs
}

type Experiment struct {
	ID            int64      `json:"id"`
	UUID          uuid.UUID  `json:"uuid"`
	Type          string     `json:"type"`
	Status        string     `json:"status"`
	Enabled       bool       `json:"enabled"`
	CreatedAt     *time.Time `json:"createdAt"`
	LastUpdatedAt *time.Time `json:"lastUpdatedAt"`
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

	etcdClient, err := etcd.New(
		etcd.Config{
			Endpoints:   specs.ETCDEndpoints,
			DialTimeout: 5 * time.Second,
			Logger:      logger,
			Password:    specs.ETCDPassword,
			Username:    "root",
		},
	)

	if err != nil {
		logger.Sugar().Fatal(err)
	}

	go func() {
		logger.Info("startng watch chan")
		for r := range etcdClient.Watch(etcd.WithRequireLeader(context.Background()), specs.KeyPrefix, etcd.WithPrefix()) {

			logger.Info("listening watch chan")

			for _, ev := range r.Events {
				logger.Sugar().Infof("Watcher says %s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				// TODO @shipperizer add unmarshalling for value
			}
		}
	}()

	for i := 0; i < 1000; i++ {
		now := time.Now()
		e := Experiment{
			ID:            rand.Int63n(int64(i) + 1),
			UUID:          uuid.New(),
			Type:          fmt.Sprintf("test-%v", i),
			Enabled:       true,
			CreatedAt:     &now,
			LastUpdatedAt: &now,
		}

		data, _ := json.Marshal(e)
		r, err := etcdClient.Put(context.Background(), e.UUID.String(), string(data))
		logger.Sugar().Info(r, err)
		time.Sleep(5 * time.Second)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until we receive our signal.
	<-c

	etcdClient.Close()
	logger.Sugar().Info("Shutting down")
	os.Exit(0)
}

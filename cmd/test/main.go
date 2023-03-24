package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type EnvSpec struct {
	KeyPrefix     string   `envconfig:"key_prefix" default:"test"`
	ETCDEndpoints []string `envconfig:"etcd_endpoints" default:"localhost:2379,localhost:2380"`
	ETCDPassword  string   `envconfig:"etcd_password"` // TODO @shipperizer just getting going, need to remove for any real docs
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
		data, _ := json.Marshal(map[string]string{"name": fmt.Sprintf("test.%v", i), "age": fmt.Sprint(i)})
		r, err := etcdClient.Put(context.Background(), fmt.Sprintf("%s.%v", specs.KeyPrefix, i), string(data))
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

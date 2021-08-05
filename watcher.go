package main

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func watch(ctx context.Context, watch WatchConfig) {

	for {

		log.Infof("conf %s start interval %d sec", watch.Name, watch.WatchIntervalSeconds)

		select {
		case <-ctx.Done():
			fmt.Printf("%s: canceld \n", watch.Name)
			return
		case <-time.After(time.Duration(watch.WatchIntervalSeconds) * time.Second):
			continue
		}
	}

}

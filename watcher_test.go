package main

import (
	"context"
	"testing"
	"time"
)

func TestWatchWillBeDoneWhenContextCancel(t *testing.T) {

	timeout := time.After(3 * time.Second)
	done := make(chan bool)

	fixture := WatchConfig{
		WatchIntervalSeconds: 120,
		Name:                 "test",
		Namespace:            "testnamesapce",
		SecretType:           "Opaque",
		SecretPath:           "/tmp/hoge",
	}
	ctx := context.Background()
	btc, canecl := context.WithCancel(ctx)
	go func() {
		watch(btc, fixture)
		done <- true
	}()

	canecl()

	select {
	case <-timeout:
		t.Fatal("Watch method not finished by cancel")
	case <-done:
	}

}

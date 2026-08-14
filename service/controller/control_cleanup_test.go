package controller

import (
	"context"
	"testing"

	corestats "github.com/xtls/xray-core/app/stats"
)

func TestUnregisterUserCounters(t *testing.T) {
	manager, err := corestats.NewManager(context.Background(), new(corestats.Config))
	if err != nil {
		t.Fatal(err)
	}
	email := "node|user@example.com|1"
	for _, name := range []string{
		"user>>>" + email + ">>>traffic>>>uplink",
		"user>>>" + email + ">>>traffic>>>downlink",
	} {
		if _, err := manager.GetOrRegisterCounter(name); err != nil {
			t.Fatal(err)
		}
	}
	controller := &Controller{stm: manager}
	if err := controller.unregisterUserCounters(email); err != nil {
		t.Fatal(err)
	}
	if manager.GetCounter("user>>>"+email+">>>traffic>>>uplink") != nil ||
		manager.GetCounter("user>>>"+email+">>>traffic>>>downlink") != nil {
		t.Fatal("user traffic counters remain registered")
	}
}

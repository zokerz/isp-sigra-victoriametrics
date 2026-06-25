package cache

import (
	"testing"
	"time"
)

func TestReadyAfterSuccessfulPoll(t *testing.T) {
	c := New()
	if c.Ready() {
		t.Fatal("cache should not be ready before success")
	}
	c.Upsert(RouterSnapshot{Name: "r1", Address: "192.0.2.1", LastPollAt: time.Now()}, true)
	if !c.Ready() {
		t.Fatal("cache should be ready after success")
	}
}

func TestFailureKeepsLastInterfaces(t *testing.T) {
	c := New()
	c.Upsert(RouterSnapshot{Name: "r1", InterfacesTotal: 1, ExportedTotal: 1}, true)
	c.Upsert(RouterSnapshot{Name: "r1", APIUp: false, LastError: "timeout"}, false)

	got, ok := c.Router("r1")
	if !ok {
		t.Fatal("router missing")
	}
	if got.ExportedTotal != 1 || got.InterfacesTotal != 1 {
		t.Fatalf("last metrics not retained: %+v", got)
	}
	if got.LastError != "timeout" {
		t.Fatalf("last error missing: %+v", got)
	}
}

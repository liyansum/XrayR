package limiter

import (
	"context"
	"sync"
	"testing"

	"github.com/wyx2685/XrayR/api"
	"golang.org/x/time/rate"
)

func TestRemoveInboundUsersReclaimsState(t *testing.T) {
	users := []api.UserInfo{
		{UID: 1, Email: "first@example.com", SpeedLimit: 100, DeviceLimit: 1},
		{UID: 2, Email: "second@example.com", SpeedLimit: 200, DeviceLimit: 2},
	}
	limiter := New()
	if err := limiter.AddInboundLimiter("node", 0, &users, nil); err != nil {
		t.Fatal(err)
	}
	value, _ := limiter.InboundInfo.Load("node")
	inbound := value.(*InboundInfo)
	removedEmail := "node|first@example.com|1"
	inbound.BucketHub.Store(removedEmail, rate.NewLimiter(1, 1))
	inbound.UserOnlineIP.Store(removedEmail, new(sync.Map))
	inbound.OldUserOnline.Store("192.0.2.1", 1)

	if err := limiter.RemoveInboundUsers("node", []string{removedEmail}); err != nil {
		t.Fatal(err)
	}
	if _, found := (*inbound.userInfo.Load())[removedEmail]; found {
		t.Fatal("removed user remains in limiter snapshot")
	}
	if _, found := inbound.BucketHub.Load(removedEmail); found {
		t.Fatal("removed user's rate bucket remains")
	}
	if _, found := inbound.UserOnlineIP.Load(removedEmail); found {
		t.Fatal("removed user's online IP state remains")
	}
	if _, found := inbound.OldUserOnline.Load("192.0.2.1"); found {
		t.Fatal("removed user's historical IP remains")
	}
	if _, found := (*inbound.userInfo.Load())["node|second@example.com|2"]; !found {
		t.Fatal("active user was removed")
	}
}

func TestOldUserOnlineIsReplacedEachReport(t *testing.T) {
	users := []api.UserInfo{{UID: 1, Email: "user@example.com"}}
	limiter := New()
	if err := limiter.AddInboundLimiter("node", 0, &users, nil); err != nil {
		t.Fatal(err)
	}
	value, _ := limiter.InboundInfo.Load("node")
	inbound := value.(*InboundInfo)
	inbound.OldUserOnline.Store("192.0.2.1", 1)
	ipMap := new(sync.Map)
	ipMap.Store("192.0.2.2", 1)
	inbound.UserOnlineIP.Store("node|user@example.com|1", ipMap)
	if _, err := limiter.GetOnlineDevice("node"); err != nil {
		t.Fatal(err)
	}
	if _, found := inbound.OldUserOnline.Load("192.0.2.1"); found {
		t.Fatal("previous reporting period remains")
	}
	if _, found := inbound.OldUserOnline.Load("192.0.2.2"); !found {
		t.Fatal("current reporting period was not retained")
	}
}

func TestDeleteInboundLimiterClosesRedisClient(t *testing.T) {
	users := []api.UserInfo{}
	limiter := New()
	config := &GlobalDeviceLimitConfig{
		Enable:       true,
		RedisNetwork: "tcp",
		RedisAddr:    "127.0.0.1:1",
		Timeout:      1,
		Expiry:       60,
	}
	if err := limiter.AddInboundLimiter("node", 0, &users, config); err != nil {
		t.Fatal(err)
	}
	value, _ := limiter.InboundInfo.Load("node")
	client := value.(*InboundInfo).GlobalLimit.redisClient
	if err := limiter.DeleteInboundLimiter("node"); err != nil {
		t.Fatal(err)
	}
	if err := client.Ping(context.Background()).Err(); err == nil {
		t.Fatal("redis client remains usable after limiter deletion")
	}
}

package registry_test

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/mervinkid/allspark/registry"
	"github.com/mervinkid/allspark/util"
	"strconv"
	"testing"
	"time"
)

func TestRedis(t *testing.T) {
	conn, err := redis.Dial("tcp", "127.0.01:6379")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Do("set", "demo/election", "node-demo", "nx", "px", 15000); err != nil {
		t.Fatal(err)
	}
}

func TestRedisRegistry(t *testing.T) {
	registrySlice := make([]registry.Registry, 1)
	for i := 0; i < 1; i++ {
		nodeId := "demo" + strconv.Itoa(i)
		config := registry.Config{}
		config.AppId = "demo"
		config.NodeId = nodeId
		config.Url = util.ParseUrl("redis://127.0.0.1:6379")
		config.Election = func(event registry.ElectionEvent, masterId string) {
			if event == registry.MasterTake {
				fmt.Println(nodeId, "take master.")
			} else {
				fmt.Println(nodeId, "take slaver.")
			}
		}
		reg, err := registry.NewRegister(config)
		if err != nil {
			t.Fatal(err)
		}
		if err := reg.Start(); err != nil {
			t.Fatal(err)
		}
		registrySlice[i] = reg
	}
	time.Sleep(10 * time.Second)
	for _, register := range registrySlice {
		register.Stop()
	}
	time.Sleep(10 * time.Second)
}

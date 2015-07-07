package main

import (
	"fmt"
	"log"

	"github.com/GeertJohan/aerospike-discovery/asdisc"
	as "github.com/aerospike/aerospike-client-go"
	"github.com/coreos/go-etcd/etcd"
)

func main() {

	watcher := asdisc.NewWatcher(&asdisc.WatcherConfig{
		EtcdClient: etcd.NewClient([]string{"http://10.0.3.56:2379"}),
	})
	ann, err := watcher.Next()
	if err != nil {
		log.Fatalf("error getting aerospike-discovery announcement: %v", err)
	}
	watcher.Close()

	// create new as client
	asClient, err := as.NewClient(ann.IP, int(ann.ServicePort))
	if err != nil {
		log.Fatalf("error creating new client: %v", err)
	}

	nodes := asClient.GetNodes()
	for _, node := range nodes {
		stats, err := as.RequestNodeStats(node)
		if err != nil {
			log.Printf("error getting stats for node %s: %v", node.GetName(), err)
			continue
		}
		fmt.Printf("name:%s cluster-key:%s cluster-size:%s objects:%s total-bytes-memory:%s total-bytes-disk:%s\n", node.GetName(), stats["cluster_key"], stats["cluster_size"], stats["objects"], stats["total-bytes-memory"], stats["total-bytes-disk"])
	}
}

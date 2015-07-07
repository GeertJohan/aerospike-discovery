package main

import (
	"fmt"
	"log"

	"github.com/GeertJohan/aerospike-discovery/asdisc"
	as "github.com/aerospike/aerospike-client-go"
)

func main() {
	// Create a new aerospike-discovery watcher using default configuration settings
	watcher := asdisc.NewWatcher(nil)

	// Retrieve an announcement
	// This method is blocking and only returns once an announcement was found,
	// that means this application could be started before the cluster is up and it will just wait for the cluster.
	ann, err := watcher.Next()
	if err != nil {
		log.Fatalf("error getting aerospike-discovery announcement: %v", err)
	}
	// Close the watcher, we're not interested in more announcements right now.
	watcher.Close()

	// Create a new Aerospike client using the announcement information
	asClient, err := as.NewClient(ann.IP, int(ann.ServicePort))
	if err != nil {
		log.Fatalf("error creating new client: %v", err)
	}

	// Loop over cluster nodes and print some simple information
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

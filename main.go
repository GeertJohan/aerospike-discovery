package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/GeertJohan/aerospike-discovery/asdisc"
	"github.com/coreos/go-etcd/etcd"
)

func main() {
	parseFlags()

	startAnnouncer()

	runWatcher()
}

func startAnnouncer() {
	announcerConfig := &asdisc.AnnouncerConfig{
		ClusterPrefix: flags.ClusterPrefix,

		TTL:      flags.AnnounceTTL,
		Interval: flags.AnnounceInterval,

		Logger: log.New(os.Stdout, `aerospike-discovery_announcer`, log.LstdFlags),
	}
	if len(flags.EtcdAddresses) > 0 {
		announcerConfig.EtcdClient = etcd.NewClient(flags.EtcdAddresses)
	}
	announcement := &asdisc.Announcement{
		Key:         flags.LocalAerospikeName,
		IP:          flags.LocalAerospikeIP,
		ServicePort: flags.LocalAerospikeServicePort,
		MeshPort:    flags.LocalAerospikeMeshPort,
	}
	_, err := asdisc.NewAnnouncer(announcement, announcerConfig)
	if err != nil {
		log.Fatalf("error creating new announcer: %v", err)
	}
}

func runWatcher() {
	// setup etcd client
	config := &asdisc.WatcherConfig{
		ClusterPrefix: flags.ClusterPrefix,
		Logger:        log.New(os.Stdout, `aerospike-discovery_watcher`, log.LstdFlags),
	}
	if len(flags.EtcdAddresses) > 0 {
		config.EtcdClient = etcd.NewClient(flags.EtcdAddresses)
	}
	watcher := asdisc.NewWatcher(config)

	// receive on the Announcements channel and maybe tip local aerospike
	for {
		ann, err := watcher.Next()
		if err != nil {
			log.Printf("error watching for announcements: %v", err)
			return
		}
		// ignore announcements about the local instance
		if ann.Key == flags.LocalAerospikeName {
			continue
		}
		go tipLocalAerospike(ann)
	}
}

func tipLocalAerospike(ann *asdisc.Announcement) {
	fmt.Printf("tipping aerospike about remote instance %s at %s:%d\n", ann.Key, ann.IP, ann.MeshPort)
	cmd := exec.Command("asinfo", "-h", flags.LocalAerospikeIP, "-p", strconv.FormatUint(uint64(flags.LocalAerospikeServicePort), 10), "-v", fmt.Sprintf("tip:host=%s;port=%d", ann.IP, ann.MeshPort))
	err := cmd.Run()
	if err != nil {
		log.Printf("error tipping local aerospike instance: %v", err)
	}
}

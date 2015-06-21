package asdisc

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

// Announcer announces a node at regular intervals.
type Announcer struct {
	etcdClient        *etcd.Client
	announcementsPath string

	annKey   string
	annValue string
	ttl      uint64
	interval uint64

	stopCh chan struct{}

	logger *log.Logger
}

// AnnouncerConfig can be used to modify the Announcer behaviour.
type AnnouncerConfig struct {
	// EtcdClient to use in announcement calls.
	// When nil, a new client with address http://localhost:2379 is used.
	// An etcd.Client should only be used once as it is not goroutine safe.
	EtcdClient *etcd.Client

	// ClusterPrefix to use in the etcd keys
	ClusterPrefix string

	// TTL and Interval for the announcement
	TTL      uint64
	Interval uint64

	// Logger for debug/info logs.
	// When nil, logs are discarded.
	Logger *log.Logger
}

// NewAnnouncer starts and returns a new Announcer instance.
// When AnnouncerConfig is nil, DefaultAnnouncerConfig is used.
// An announcer announces a single aerospike node at regular intervals, for this a goroutine is started.
// Use the Stop method to stop announcing.
func NewAnnouncer(announcement *Announcement, config *AnnouncerConfig) (*Announcer, error) {
	// use default config is none is given
	if config == nil {
		config = DefaultAnnouncerConfig
	}

	// create new announcer
	announcer := &Announcer{
		etcdClient:        config.EtcdClient,
		announcementsPath: config.ClusterPrefix,

		annKey:   announcement.Key,
		ttl:      config.TTL,
		interval: config.Interval,

		stopCh: make(chan struct{}),

		logger: config.Logger,
	}

	// set defaults when no custom value was set
	if announcer.etcdClient == nil {
		announcer.etcdClient = etcd.NewClient(DefaultEtcdAddresses)
	}
	if announcer.logger == nil {
		announcer.logger = DefaultLogger
	}
	if len(announcer.announcementsPath) == 0 {
		announcer.announcementsPath = DefaultClusterPrefix
	}
	announcer.announcementsPath = path.Join(announcer.announcementsPath, "announcements")

	// check ttl and interval
	if announcer.interval == 0 {
		return nil, errors.New("interval must be above zero")
	}
	if announcer.interval >= announcer.ttl {
		return nil, errors.New("ttl must be greater than announce interval")
	}
	if announcer.ttl > 120 {
		announcer.logger.Println("ttl is set to a very high value (>120), please note that the ttl unit is seconds")
	}

	// marshall announcement to json
	announcementValue, err := json.Marshal(announcement)
	if err != nil {
		return nil, fmt.Errorf("error marshalling announcement to json: %v", err)
	}
	announcer.annValue = string(announcementValue)

	// create prefix dir
	_, err = announcer.etcdClient.CreateDir(announcer.announcementsPath, 0)
	if err != nil {
		// ignore errorcode 105 (dir already exists)
		if etcdErr, ok := err.(*etcd.EtcdError); !ok || etcdErr.ErrorCode != 105 {
			return nil, fmt.Errorf("error creating dir in etcd: %v", err)
		}
	}

	// start announcer goroutine
	go announcer.run()

	return announcer, nil
}

func (a *Announcer) run() {
	for {
		a.etcdClient.Set(path.Join(a.announcementsPath, a.annKey), a.annValue, a.ttl)
		a.logger.Printf("announced %s as %s\n", a.annKey, a.annValue)
		select {
		case <-time.After(time.Duration(a.interval) * time.Second):
		case <-a.stopCh:
			_, err := a.etcdClient.Delete(path.Join(a.announcementsPath, a.annKey), false)
			if err != nil {
				a.logger.Printf("error deleting announcement from etcd: %v", err)
			}
			return
		}
	}
}

// Stop stops the announcer.
// This also removes the announcement from etcd.
// It is safe to call Stop multiple times.
func (a *Announcer) Stop() {
	select {
	case <-a.stopCh:
	default:
		close(a.stopCh)
	}
}

package asdisc

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

var (
	// ErrWatcherClosed is returned by (*Watcher).Next() when the Watcher was closed.
	ErrWatcherClosed = errors.New("watcher is closed")
)

// Watcher watches aerospike-discovery announcements.
// The creator must either continiously receive from the Announcements channel or close the Watcher.
type Watcher struct {
	etcdClient        *etcd.Client
	announcementsPath string

	annCh  chan *Announcement
	stopCh chan struct{}

	known map[string]string

	err error

	logger *log.Logger
}

// WatcherConfig can be used to pass custom settings to NewWatcher
type WatcherConfig struct {
	// EtcdClient for Get and Watch calls. When nil a new etcd.Client is created with DefaultEtcdAddress as path.
	EtcdClient *etcd.Client

	// ClusterPrefix used in the etcd keys. DefaultClusterPrefix is used when empty.
	ClusterPrefix string

	// Logger to send info/debug logs on. When nil the DefaultLogger is used.
	Logger *log.Logger
}

// NewWatcher starts a new watcher. The config argument can be nil, in which case DefaultWatcherConfig will be used.
// Read all announcements and their updates by calling the Next() method.
func NewWatcher(config *WatcherConfig) *Watcher {
	if config == nil {
		config = DefaultWatcherConfig
	}
	watcher := &Watcher{
		etcdClient:        config.EtcdClient,
		announcementsPath: config.ClusterPrefix,

		annCh:  make(chan *Announcement),
		stopCh: make(chan struct{}),

		known: make(map[string]string),

		logger: config.Logger,
	}

	// set defaults when no custom value was given
	if watcher.etcdClient == nil {
		watcher.etcdClient = etcd.NewClient(DefaultEtcdAddresses)
	}
	if watcher.logger == nil {
		watcher.logger = DefaultLogger
	}
	if len(watcher.announcementsPath) == 0 {
		watcher.announcementsPath = DefaultClusterPrefix
	}
	watcher.announcementsPath = path.Join(watcher.announcementsPath, "announcements")

	go watcher.run()

	return watcher
}

// run is stared by NewWatcher as a goroutine and takes care of watching for new announcements.
func (w *Watcher) run() {
	// read existing announcements
	resp, err := w.etcdClient.Get(w.announcementsPath, false, true)
	if err != nil {
		w.err = fmt.Errorf("error getting announcements: %v", err)
		w.Close()
		return
	}
	w.logger.Printf("have %d announcements on initial load", len(resp.Node.Nodes))
	for _, node := range resp.Node.Nodes {
		w.handleResp(node.Key, node.Value)
	}

	// watch for changed/new announcements
	watchRecvCh := make(chan *etcd.Response)
	watchStopCh := make(chan bool)
	go func() {
		_, err := w.etcdClient.Watch(w.announcementsPath, resp.EtcdIndex, true, watchRecvCh, watchStopCh)
		if err != nil {
			w.err = fmt.Errorf("error watching for announcements: %v", err)
			w.Close()
		}
	}()
	for {
		select {
		case <-w.stopCh:
			watchStopCh <- true
			return
		case resp, ok := <-watchRecvCh:
			if !ok {
				// etcd watcher has stopped
				return
			}
			w.handleResp(resp.Node.Key, resp.Node.Value)
		}
	}

}

// handleResp unmarshals an announcement from an etcd.Response and sends it on the *Watcher.Announcements channel.
func (w *Watcher) handleResp(fullKey, value string) {
	key := strings.Trim(strings.TrimPrefix(fullKey, w.announcementsPath), `/`)
	if len(key) == 0 {
		return
	}
	// remove when value is empty (announcement was removed)
	if len(value) == 0 {
		delete(w.known, key)
		return
	}
	// we're not interested in announcements that have not changed
	if w.known[key] == value {
		return
	}
	w.known[key] = value

	// create Announcement object to send on channel
	ann := &Announcement{
		Key: key,
	}
	err := json.Unmarshal([]byte(value), ann)
	if err != nil {
		w.logger.Printf("error unmarshalling announcement with key '%s': %v", key, err)
		return
	}

	select {
	case <-w.stopCh:
	case w.annCh <- ann:
	}

}

// Close stops the watcher
func (w *Watcher) Close() {
	select {
	case <-w.stopCh:
	default:
		// if there was no error yet, set ErrWatcherClosed
		if w.err == nil {
			// there's a race here but it could only cause w.err to be set with ErrWatcherClosed multiple times
			w.err = ErrWatcherClosed
		}

		// indicate this watcher is stopping
		close(w.stopCh)
	}
}

// Next gives a new announcement or an update to an announcement that was returned by Next() earlier. When an error occurred during watching, that error is returned.
// When the Watcher was closed, ErrWatcherClosed is returned.
func (w *Watcher) Next() (*Announcement, error) {
	select {
	case ann := <-w.annCh:
		return ann, nil
	case <-w.stopCh:
		return nil, w.err
	}
}

package asdisc

import (
	"io/ioutil"
	"log"
)

var (
	// DefaultEtcdAddresses used when no custom *etcd.Client is given
	DefaultEtcdAddresses = []string{`http://localhost:2379`}

	// DefaultClusterPrefix is using geertjohan.net to avoid collision. Chances are tiny anyone else will use this key.
	DefaultClusterPrefix = `/geertjohan.net/aerospike-discovery/default`

	// DefaultLogger discards logs
	DefaultLogger = log.New(ioutil.Discard, ``, 0)

	// DefaultAnnouncerConfig is used by NewAnnouncer when no custom *AnnouncerConfig is given.
	DefaultAnnouncerConfig = &AnnouncerConfig{
		TTL:      60,
		Interval: 45,
	}

	// DefaultWatcherConfig is used by NewWatcher when no custom *WatcherConfig is given
	DefaultWatcherConfig = &WatcherConfig{}
)

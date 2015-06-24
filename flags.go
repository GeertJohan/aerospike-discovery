package main

import (
	"fmt"
	"os"

	goflags "github.com/jessevdk/go-flags"
)

// flags holds the flags and arguments given to this process when parseFlags() has ran.
// flags is filled with defaults from the tags and the initFlags function.
// The default EtcdPrefix is using geertjohan.net to avoid collision. Chances are tiny anyone else will use this too.
var flags struct {
	EtcdAddresses []string `long:"etcd-address" description:"etcd (proxy) address" default:"http://localhost:2379"`

	ClusterPrefix string `long:"cluster-prefix" description:"prefix (path) for the announcement keys in etcd"`

	LocalAerospikeName        string `long:"local-aerospike-name" description:"Name for the local aerospike instance, this name must be unique througout the aerospike cluster" required:"true"`
	LocalAerospikeIP          string `long:"local-aerospike-ip" description:"IP address for the local Aerospike instance" required:"true"`
	LocalAerospikeServicePort uint16 `long:"local-aerospike-service-port" description:"Service port for the local Aerospike instance" default:"3000"`
	LocalAerospikeMeshPort    uint16 `long:"local-aerospike-mesh-port" description:"Mesh port for the local Aerospike instance" default:"3002"`

	AnnounceTTL      uint64 `long:"announce-ttl" decsription:"announce time-to-live in seconds, set as ttl to the announcement value in etcd" default:"60"`
	AnnounceInterval uint64 `long:"announce-interval" description:"announce interval in seconds, this should always be a lower value than --announce-ttl" default:"45"`

	LoggingDisableTimestamp bool `long:"logging-disable-timestamp" description:"do not write timestamp before each log message"`
}

// parseFlags parses the given arguments.
// when the user asks for help (-h or --help): the application exists with status 0
// when unexpected flags is given: the application exits with status 1
func parseFlags() {
	args, err := goflags.Parse(&flags)
	if err != nil {
		// assert the err to be a flags.Error
		flagError := err.(*goflags.Error)
		if flagError.Type == goflags.ErrHelp {
			// user asked for help on flags.
			// program can exit successfully
			os.Exit(0)
		}
		if flagError.Type == goflags.ErrUnknownFlag {
			fmt.Println("Use --help to view all available options.")
			os.Exit(1)
		}
		fmt.Printf("error parsing flags: %s\n", err)
		os.Exit(1)
	}

	// check for unexpected arguments
	// when an unexpected argument is given: the application exists with status 1
	if len(args) > 0 {
		fmt.Printf("error: unknown argument '%s'.\n", args[0])
		os.Exit(1)
	}

	//++ TODO: verify IPv4 or IPv6
}

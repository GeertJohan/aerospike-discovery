## Aerospike cluster node discovery with etcd

This project provides the tools to easily setup and use an Aerospike cluster with etcd.

This project contains two parts:
 - The `aerospike-discovery` command can run as a sidekick process to each Aerospike node in a cluster and tips the nodes about each other.
 - The `asdisc` package allows you to write a custom announcement/discovery with an Aerospike cluster.

### aerospike-discovery
`aerospike-discovery` can be ran as sidekick process to an Aerospike node and performs two tasks:
 - announce the local Aerospike node to etcd.
 - watch etcd for announcements and tip the local Aerospike node about new nodes in the cluster.

#### Docker
This repository is available as automatically built docker container image in docker hub as `geertjohan/aerospike-discovery`.

#### CoreOS example (systemd, flannel)
TODO

### asdisc
The asdisc package provides a simple API to connect any application to an aerospike cluster.

TODO: usage example

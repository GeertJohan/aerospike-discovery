
## Thoughts:

### --mode=sidekick vs --mode=master
As sidekick, `aerospike-discovery` expects --local-aerospike-{ip,name} and announces the external process. starting/stopping announcements is achieved by systemd (or any other system) starting and stopping `aerospike-discovery` where `aerospike` itself crashes or stops.

As master, `aerospike-discovery` starts `asd` as child process. It asumes the hostname as --local-aerospike-name (when not given) and also figures out the local ip address. Arguments to asd can be passed after `--`. This mode makes deployment easier, but less configurable.

### Dynamic config
We have a sidekick process running next to each Aerospike node. Why not add a tool that allows to set configuration paramters through etcd, aerospike-discovery would then apply the config parameters on the local node. Would this work? What happens with errors from asinfo? How to check consistently the parameters are applied. What to do when an aerospike node is restarted?

### Possible issue? tipping before "soon there will be cake" can fail??
Had once: the tipping didn't have anny effect. This caused the setup of the cluster to fail.

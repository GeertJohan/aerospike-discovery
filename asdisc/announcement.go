package asdisc

// Announcement contains information about a single node in an Aerospike cluster.
// The Announcement is marshalled to json and stored in etcd.
type Announcement struct {
	Key         string `json:"-"`
	IP          string `json:"ip"`
	ServicePort uint16 `json:"servicePort"`
	MeshPort    uint16 `json:"meshPort"`
}

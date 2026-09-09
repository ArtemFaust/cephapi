package radosapi

import "github.com/ceph/go-ceph/rados"

type ClusterInfo struct {
	Stats           rados.ClusterStat
	Fsid            string
	InstanseID      uint64
	Addrs           string
	PublicNetwork   string
	ClusterNetwork  string
	MonHosts        string
	MaxPgByOsd      string
	MonAllowPoolDel string
	PgAutoScale     string
	STS             string
	STSKey          string
}

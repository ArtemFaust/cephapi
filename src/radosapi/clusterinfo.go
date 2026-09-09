package radosapi

import (
	"encoding/json"
	"fmt"

	"github.com/ceph/go-ceph/rados"
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// Метод получения информации о кластере
func GetClusterInfo(conn *rados.Conn, format string) error {
	clusterInfo := ClusterInfo{}
	// Получаем информацию о кластере
	stats, e := conn.GetClusterStats()
	if e != nil {
		return e
	}
	fsid, e := conn.GetFSID()
	if e != nil {
		return e
	}
	inst_id := conn.GetInstanceID()
	addrs, e := conn.GetAddrs()
	if e != nil {
		return e
	}
	mon_host, e := conn.GetConfigOption("mon_host")
	if e != nil {
		return e
	}
	pnet, e := conn.GetConfigOption("public_network")
	if e != nil {
		return e
	}
	cnet, e := conn.GetConfigOption("cluster_network")
	if e != nil {
		return e
	}
	mpbo, e := conn.GetConfigOption("mon_max_pg_per_osd")
	if e != nil {
		return e
	}
	mapd, e := conn.GetConfigOption("mon_allow_pool_delete")
	if e != nil {
		return e
	}
	pgautoscale, e := conn.GetConfigOption("osd_pool_default_pg_autoscale_mode")
	if e != nil {
		return e
	}
	sts, e := conn.GetConfigOption("rgw_s3_auth_use_sts")
	if e != nil {
		return e
	}
	stskey, e := conn.GetConfigOption("rgw_sts_key")
	if e != nil {
		return e
	}

	// Формируем структуру
	clusterInfo.Stats = stats
	clusterInfo.Fsid = fsid
	clusterInfo.InstanseID = inst_id
	clusterInfo.Addrs = addrs
	clusterInfo.MonHosts = mon_host
	clusterInfo.PublicNetwork = pnet
	clusterInfo.ClusterNetwork = cnet
	clusterInfo.MaxPgByOsd = mpbo
	clusterInfo.MonAllowPoolDel = mapd
	clusterInfo.PgAutoScale = pgautoscale
	clusterInfo.STS = sts
	clusterInfo.STSKey = stskey

	if format == "json" {
		e = jsonPrint(clusterInfo)
		return e
	} else {
		tablePrint(clusterInfo)
	}
	return nil
}

func jsonPrint(s ClusterInfo) error {
	b, e := json.Marshal(s)
	if e != nil {
		return e
	}
	fmt.Println(string(b))
	return nil
}

func tablePrint(s ClusterInfo) {
	headerFmt := color.New(color.FgGreen, color.Bold).SprintfFunc()
	columnFmt := color.New(color.FgHiYellow, color.Bold).SprintfFunc()
	tbl := table.New("ATTR", "VALUE")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithHeaderSeparatorRow('-').WithPadding(1)

	tbl.AddRow("FSID", s.Fsid)
	tbl.AddRow("ADDRS", s.Addrs)
	tbl.AddRow("C NET", s.ClusterNetwork)
	tbl.AddRow("P NET", s.PublicNetwork)
	tbl.AddRow("INSTANCE ID", s.InstanseID)
	tbl.AddRow("MON/S", s.MonHosts)
	tbl.AddRow("SIZE Kb", s.Stats.Kb)
	tbl.AddRow("AVALIABLE Kb", s.Stats.Kb_avail)
	tbl.AddRow("USED Kb", s.Stats.Kb_used)
	tbl.AddRow("OBJECTS", s.Stats.Num_objects)
	tbl.AddRow("MON MAX PG PER OSD", s.MaxPgByOsd)
	tbl.AddRow("MON ALLOW POOL DEL", s.MonAllowPoolDel)
	tbl.AddRow("PG AUTOSCALE", s.PgAutoScale)
	tbl.AddRow("STS", s.STS)
	tbl.AddRow("STS KEY", s.STSKey)

	tbl.Print()
}

package utils

import (
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// Метод печати примеров использования
func PrintUsageExamples() {
	headerFmt := color.New(color.FgGreen, color.Bold).SprintfFunc()
	columnFmt := color.New(color.FgHiYellow, color.Bold).SprintfFunc()
	tbl := table.New("EXANPLE", "DESCRIPTION")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithHeaderSeparatorRow('-').WithPadding(1)
	tbl.AddRow("-rgw -lu -format table|json -e <endpoint> -size kb|gb|tb", "Get rgw user list")
	tbl.AddRow("-rgw -lb -format table|json -e <endpoint> -size kb|gb|tb", "Get buckets list")
	tbl.AddRow(`-rgw -cu 
	-rup \"<USER ID>|<USER DISPLAY NAME>|<USER EEMAIL> or empty|<QUOTA SIZE IN Kb> or -1 or not set|<QUOTA OBJECTS> or -1 or not set|\"
	-cups \"buckets=*;users=*;usage=read;metadata=read;zone=read\"
	-e <endpoint> -format table|json`, "Create new RGW user")
	tbl.AddRow("-rgw -guk -uid <UID> -e <endpoint> -format table|json", "Get S3 user keys and cups")
	tbl.AddRow("-rgw -cuc -uid <UID> -caps \"buckets=*;users=*\" -e <endpoint>", "Change or set user caps")
	tbl.AddRow("-rgw -ruc -uid <UID> -e <endpoint>", "Remove user caps")
	tbl.AddRow("-rgw -ck -e <enpoint> -uid <uid> -kt <s3|swift> -format <json|table> -suid <subuser uid>", "Cretae new s3 key. Use suid for greate new key over subuser")
	tbl.AddRow("-rados -e <endpoint> -cf -fsname <fs_name> -placement <* or number or servers>", "Create new ceph fs")
	tbl.AddRow("-rados -e <endpoint> -cpc -pool <pool name> -crush_rule <rule name>", "Change exists pool crush rule")
	tbl.AddRow("-rados -e <endpoint> -cpa -pool <pool name> -dpa|-mpa \"off|32|none|0|0\"", "Change pool attributes")
	tbl.AddRow("-rados -e <endpoint> -cu -user <user entry name> -caps \"mon==<value>;mds==<value>;osd==<value>\"", "Create new ceph user with caps")
	tbl.AddRow("-rados -e <endpoint> -cuc -user <user entry name> -caps \"mon==<value>;mds==<value>;osd==<value>\"", "Set or update ceph user caps")
	tbl.AddRow("-rados -e <endpoint> -guk -user <user entry name>", "Get ceph user auth data")
	tbl.AddRow("-rados -e <endpoint> -gf -format table -fsname <fs_name>", "Get ceph fs info")
	tbl.AddRow("-rados -e <endpoint> -gfs -format table|json -fsname <cephfs_name>", "Get cephfs subvolumes")
	tbl.AddRow("-rados -e <endpoint> -ci -format table|json", "Get cluster info")
	tbl.Print()
}

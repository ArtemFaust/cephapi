package rgwapi

import "github.com/ceph/go-ceph/rgw/admin"

// Структура описавающая информацию о пользователе RGW
type RgwUser struct {
	Uid      string
	UserInfo struct {
		Error error
		Info  admin.User
	}
	UserQuota struct {
		Error error
		admin.QuotaSpec
	}
	Buckets []struct {
		Error  error
		Bucket admin.Bucket
	}
}

// Структура описывающая доступные подключения (rgw)
type Connection struct {
	RGW RGW `yaml:"RGW"`
}
type RGW struct {
	EndPoints []struct {
		Name       string `yaml:"Name"`
		S3Endpoint string `yaml:"S3Endpoint"`
		AcessKey   string `yaml:"AcessKey"`
		SecretKey  string `yaml:"SecretKey"`
	} `yaml:"EndPoints"`
}

package utils

// Структура описывающая доступные типы подключений
type ConnectionEndpoints struct {
	Rgw struct {
		Name       string "yaml:\"Name\""
		S3Endpoint string "yaml:\"S3Endpoint\""
		AcessKey   string "yaml:\"AcessKey\""
		SecretKey  string "yaml:\"SecretKey\""
	}
	Rados struct {
		Name                 string "yaml:\"Name\""
		Fsid                 string "yaml:\"Fsid\""
		MonHosts             string "yaml:\"MonHosts\""
		DefaultMetaCrushRule string "yaml:\"DefaultMetaCrushRule\""
		DefaultDataCrushRule string "yaml:\"DefaultDataCrushRule\""
		KeyRing              string "yaml:\"KeyRing\""
	}
}

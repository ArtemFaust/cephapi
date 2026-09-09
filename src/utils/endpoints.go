package utils

import (
	"cephapi/cephfsapi"
	"cephapi/rgwapi"
	"errors"
	"os"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
)

// Метод сичтвания файла конфигурации с описанием endpoint
func Endpoints() (struct {
	Rgw   rgwapi.Connection
	Rados cephfsapi.Connection
}, error) {
	en_rgw := rgwapi.Connection{}
	en_rados := cephfsapi.Connection{}
	endpoints := struct {
		Rgw   rgwapi.Connection
		Rados cephfsapi.Connection
	}{}

	enspointsconfig, e := FindEndpointsConfig()
	if e != nil {
		return endpoints, e
	}

	_, e = os.Stat(enspointsconfig)
	if e != nil {
		logrus.Error("endpoints not found: ", e)
		return struct {
			Rgw   rgwapi.Connection
			Rados cephfsapi.Connection
		}{}, e
	}

	b, e := os.ReadFile(enspointsconfig)
	if e != nil {
		logrus.Error("Error read endpoints from file: ", e)
		return struct {
			Rgw   rgwapi.Connection
			Rados cephfsapi.Connection
		}{}, e
	}

	// Декодируем фаил endpoints.yaml
	db, e := DecodeFile(b)
	if e != nil {
		logrus.Error("Error read endpoints from file: ", e)
		return struct {
			Rgw   rgwapi.Connection
			Rados cephfsapi.Connection
		}{}, e
	}

	// Создание endpoints rgw
	e = yaml.Unmarshal(db, &en_rgw)
	if e != nil {
		return struct {
			Rgw   rgwapi.Connection
			Rados cephfsapi.Connection
		}{}, e
	}

	// Создание endpoints rados
	e = yaml.Unmarshal(db, &en_rados)
	if e != nil {
		return struct {
			Rgw   rgwapi.Connection
			Rados cephfsapi.Connection
		}{}, e
	}

	endpoints.Rgw = en_rgw
	endpoints.Rados = en_rados

	return endpoints, e
}

// Метод поиска файла конфигурации endpoints.yaml
// 1 - ищем в рабочей директории запуска
// 2 - ищем в рабочей директории внутри ./cephapi/
// 3 - ищем в домашней директории пользоватлея $HOMEDIR/./cephapi/
func FindEndpointsConfig() (string, error) {
	// Сначала ищем фаил конфигурации в рабочей директории запуска
	_, e := os.Stat("./endpoints.yaml")
	if e == nil {
		return "./endpoints.yaml", nil
	}
	// Ищем фаил конфигурации в скрытой директории .cephapi/endpoints.yaml директории запуска
	_, e = os.Stat("./.cephapi/endpoints.yaml")
	if e == nil {
		return "./.cephapi/endpoints.yaml", nil
	}
	// Ищем в директории $home_dir/.cephapi/endpoints.yaml
	h, e := os.UserHomeDir()
	if e != nil {
		return "", e
	}
	_, e = os.Stat(h + "/.cephapi/endpoints.yaml")
	if e == nil {
		return h + "/.cephapi/endpoints.yaml", nil
	}
	// Если ни где не нашли тогда говорим что не нашли
	return "", errors.New("can not find endpoints configuration file")
}

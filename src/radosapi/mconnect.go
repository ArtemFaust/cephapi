package radosapi

import (
	"os"
	"strings"

	"github.com/ceph/go-ceph/rados"
	"github.com/sirupsen/logrus"
)

// Метод создания нового подключения к rados
func MakeNewRadosConnection(fsid string, mon_hosts string, keyring string) (*rados.Conn, error) {
	// 1. Создаем и настраиваем соединение
	conn, e := rados.NewConn()
	if e != nil {
		logrus.Error("Ошибка создания нового подключения к rados: ", e)
		return nil, e
	}

	// Чистим конфиг от паразитных строк
	keyring = strings.TrimSpace(keyring)
	keyring = strings.TrimRight(keyring, "\r")
	keyring = strings.TrimRight(keyring, "\r\n")
	// Добавляем перенос строки в конец так как иначе возникает ошибка парсинга
	keyring = keyring + "\n"

	// Создаем ключ для подключения
	e = os.WriteFile("/tmp/ceph.client.admin.keyring", []byte(keyring), 0600)
	if e != nil {
		logrus.Error("Ошибка записи административного ключа: ", e)
		return nil, e
	}

	// Устанавливаем параметры подключения
	conn.SetConfigOption("fsid", fsid)
	conn.SetConfigOption("mon_host", mon_hosts)
	conn.SetConfigOption("keyring", "/tmp/ceph.client.admin.keyring")

	// Gопдключение к rados
	e = conn.Connect()
	if e != nil {
		logrus.Error("Ошибка создания нового подключения к rados: ", e)
		return nil, e
	}

	return conn, nil
}

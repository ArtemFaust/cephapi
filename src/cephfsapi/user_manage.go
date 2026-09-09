package cephfsapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ceph/go-ceph/rados"
	"github.com/sirupsen/logrus"
)

// Создание нового cepf пользователя
func CreateNewCephUser(conn *rados.Conn, userentry string, caps string) error {
	if userentry == "" {
		logrus.Error("userentry is empty!")
		return errors.New("userentry is empty!")
	}
	// Проверяем наличие client. в имени создаваемого пользователя
	if !strings.Contains(userentry, "client.") {
		userentry = "client." + userentry
	}

	// Создание команды к MON серверу
	cmd, e := func() ([]byte, error) {
		// Если caps не переданны
		if caps == "" {
			return json.Marshal(map[string]interface{}{
				"prefix": "auth get-or-create",
				"entity": userentry,
			})
		}

		// Если caps переданны
		usercaps, e := cupsParser(caps)
		fmt.Println(usercaps[1])
		if e != nil {
			return nil, e
		}

		return json.Marshal(map[string]interface{}{
			"prefix": "auth get-or-create",
			"entity": userentry,
			"caps":   usercaps,
		})

	}()

	if e != nil {
		logrus.Error("Ошибка исполнения:", e)
		return e
	}

	// Отправляем команду монитору
	response, info, e := conn.MonCommand(cmd)
	if e != nil {
		logrus.Errorf("Ошибка выполнения команды:\nОшибка: %v\nДетали: %s\nОтвет: %s",
			e, info, string(response))
		return e
	}

	// Проверяем наличие пользователя
	e = GetUserInfo(conn, userentry)
	if e != nil {
		return e
	}

	logrus.Info("Выполненно создание пользователя \n", string(response), "\n", info)

	return nil
}

// Метод устанвки/изменения caps для пользователя
// ceph auth caps client.john mon 'allow r' osd 'allow rw pool=mypool'
func SetCephUserCaps(conn *rados.Conn, userentry string, caps string) error {
	if userentry == "" || caps == "" {
		logrus.Error("error set user caps - atributes is empty!")
		return errors.New("error set user caps - atributes is empty!")
	}

	usercaps, e := cupsParser(caps)
	if e != nil {
		logrus.Error("error parse user caps!")
		return e
	}

	// Проверяем наличие пользователя
	e = GetUserInfo(conn, userentry)
	if e != nil {
		return e
	}

	// Создание команды к MON серверу
	cmd, e := json.Marshal(map[string]interface{}{
		"prefix": "auth caps",
		"entity": userentry,
		"caps":   usercaps,
	})

	if e != nil {
		logrus.Error("Ошибка исполнения:", e)
		return e
	}

	// Отправляем команду монитору
	response, info, e := conn.MonCommand(cmd)
	if e != nil {
		logrus.Errorf("Ошибка выполнения команды:\nОшибка: %v\nДетали: %s\nОтвет: %s",
			e, info, string(response))
		return e
	}
	logrus.Info("Выполненно обновление caps пользователя \n", string(response), "\n", info)

	return nil
}

// Метод удаления пользователя
func DeleteCephUser(conn *rados.Conn, userentry string) error {
	if userentry == "" || !strings.Contains(userentry, "client.") {
		logrus.Error("Ошибка выполнения - имя пользователя переданно не верно")
		return errors.New("userentry is incorrect")
	}

	// Проверяем наличие пользователя
	e := GetUserInfo(conn, userentry)
	if e != nil {
		return e
	}

	// Создание команды к MON серверу
	cmd, e := json.Marshal(map[string]interface{}{
		"prefix": "auth rm",
		"entity": userentry,
	})

	if e != nil {
		logrus.Error("Ошибка исполнения ", e)
		return e
	}

	// Отправляем команду монитору
	response, info, e := conn.MonCommand(cmd)
	if e != nil {
		logrus.Errorf("Ошибка выполнения команды:\nОшибка: %v\nДетали: %s\nОтвет: %s",
			e, info, string(response))
		return e
	}
	logrus.Info("Выполненно удаление пользователя - ", userentry, "\n", string(response), "\n", info)

	return nil
}

// Метод получения информации о пользователе
func GetUserInfo(conn *rados.Conn, userentry string) error {
	if userentry == "" || !strings.Contains(userentry, "client.") {
		logrus.Error("Ошибка выполнения - имя пользователя переданно не верно")
		return errors.New("userentry is incorrect")
	}

	// Создание команды к MON серверу
	cmd, e := json.Marshal(map[string]interface{}{
		"prefix": "auth get",
		"entity": userentry,
	})

	if e != nil {
		logrus.Error("Ошибка исполнения ", e)
		return e
	}

	// Отправляем команду монитору
	response, info, e := conn.MonCommand(cmd)
	if e != nil {
		logrus.Errorf("Ошибка выполнения команды:\nОшибка: %v\nДетали: %s\nОтвет: %s",
			e, info, string(response))
		return e
	}
	logrus.Info("Выполненно получение данных пользователя - ", userentry, "\n", string(response), "\n", info)

	return nil
}

// внутренний метод парсинга caps
// ожидаем на входе caps вида "mon==allow r;mds==allow rw fsname=cephfs01; osd==allow rwx tag cephfs data=cephfs01"
func cupsParser(caps string) ([]string, error) {
	itter_caps := []string{}
	if caps == "" {
		return nil, errors.New("error caps is empty!")
	}

	s_caps := strings.Split(caps, ";")
	if len(s_caps) == 0 {
		return nil, errors.New("error parse caps!")
	}

	for _, cap := range s_caps {
		cap = strings.TrimSpace(cap)
		if cap != "" {
			v := strings.Split(cap, "==")[0]
			c := strings.Split(cap, "==")[1]
			itter_caps = append(itter_caps, v)
			itter_caps = append(itter_caps, c)
		}
	}
	return itter_caps, nil
}

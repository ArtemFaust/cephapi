package rgwapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ceph/go-ceph/rgw/admin"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/sirupsen/logrus"
)

// Метод создания нового пользователя
func CreateRgwUser(co *admin.API, user string, format string, cups string, hsd string, refresh bool, nocache bool) error {
	// Парсим параметры пользователя
	if len(strings.Split(user, "|")) < 3 {
		return errors.New("failed parse user params!")
	}

	deadline := time.Now().Add(20 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// Создаем пользователя
	cu, e := co.CreateUser(ctx, admin.User{
		ID:          strings.Split(user, "|")[0],
		DisplayName: strings.Split(user, "|")[1],
		Email:       strings.Split(user, "|")[2],
		//DefaultPlacement: ,
	})
	if e != nil {
		return e
	}

	// Парсим параметры пользователя
	if len(strings.Split(user, "|")) > 3 {
		q_max_size_kb := int(-1)
		q_max_objects := int64(-1)

		if strings.Split(user, "|")[3] == "" {
			q_max_size_kb = -1
		} else {
			q_max_size_kb, e = strconv.Atoi(strings.Split(user, "|")[3]) // Размер квоты в kb
			if e != nil {
				return errors.New("error parse quota params!")
			}
		}

		if strings.Split(user, "|")[4] == "" {
			q_max_objects = -1
		} else {
			q_max_objects, e = strconv.ParseInt(strings.Split(user, "|")[4], 10, 64) // Размер квоты для объектов
			if e != nil {
				return errors.New("error parse quota params!")
			}
		}

		enabled := true
		if q_max_size_kb == -1 && q_max_objects == -1 {
			enabled = false
		}

		// Устанавливаем квоты для создаваемого пользователя
		e = co.SetUserQuota(ctx, admin.QuotaSpec{
			UID:        strings.Split(user, "|")[0],
			Enabled:    &enabled,
			QuotaType:  "user",
			MaxSizeKb:  &q_max_size_kb,
			MaxObjects: &q_max_objects,
		})
	}

	// Если переданны CUPS для пользователя
	// то вызываем метод их установки
	if cups != "" {
		e = SetUserCaps(co, cu.ID, cups, "")
		if e != nil {
			// Ошибка не критична для операции создания пользователя
			// поэтому только печаем сообщение об ошибке
			logrus.Error("failed set user cups! ", e)
		}
	}

	// параметры печати результата
	if format == "table" {
		u, e := GetUserList(co, refresh, nocache)
		if e != nil {
			return e
		}
		UserTprint(*u, strings.Split(user, "|")[0], hsd)
		return nil
	}
	if format == "json" {
		u, e := GetUserList(co, refresh, nocache)
		if e != nil {
			return e
		}
		e = UserJPrint(*u, strings.Split(user, "|")[0])
		if e != nil {
			return e
		}
		return nil
	}
	return nil
}

// Метод получения ключей пользователя
func GetUserKeys(co *admin.API, uid string, format string) error {
	deadline := time.Now().Add(20 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	u, e := co.GetUser(ctx, admin.User{
		ID: uid,
	})
	if e != nil {
		return e
	}

	if format == "json" {
		b, e := json.Marshal(u)
		if e != nil {
			return e
		}
		fmt.Println(string(b))
		return nil
	}
	headerFmt := color.New(color.FgGreen, color.Bold).SprintfFunc()
	columnFmt := color.New(color.FgHiYellow, color.Bold).SprintfFunc()

	tbl := table.New("UID", "DISPLAY NAME", "USER TYPE", "ACCESS KEY", "SECRET KEY", "KEY TYPE", "CAPS")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithHeaderSeparatorRow('-').WithPadding(2)

	// Перебираем ключи основного пользователя
	for _, key := range u.Keys {
		tbl.AddRow(key.User, u.DisplayName, u.Type, key.AccessKey, key.SecretKey, "S3", u.Caps)
	}

	// Перебираем subusers основного пользователя и берем его ключи
block:
	for _, su := range u.Subusers {
		// SWIFT ключи если для subuser найден swift ключ
		// значит это swift subuser добавляем его в список вывода
		// и переходим к следующей итерации внешнего блока
		for _, sf_k := range u.SwiftKeys {
			if su.Name == sf_k.User {
				tbl.AddRow(sf_k.User, "", "", "", sf_k.SecretKey, "SWIFT", su.Access)
			}
			continue block // Переходит к следующей итерации внешнего цикла
		}
		// Добавляем пользователя в вывод
		tbl.AddRow(su.Name, "", "", su.AccessKey, su.SecretKey, "S3", su.Access)
	}

	tbl.Print()

	return nil
}

// Метод установки CAPS для пользователя
// "type:value|type:value"
func SetUserCaps(co *admin.API, uid string, cups string, format string) error {
	deadline := time.Now().Add(20 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// Если переданные параметры не парсятся
	if len(strings.Split(cups, "|")) == 0 {
		return errors.New("failed parse caps!")
	}

	// Сначала удалем user caps
	e := RemoveUserCaps(co, uid)
	if e != nil {
		return e
	}

	// Модифицируем параметры CUPS пользователя
	_, e = co.AddUserCap(ctx, uid, cups)
	if e != nil {
		return e
	}
	return nil
}

// Метод удаление caps пользователя
func RemoveUserCaps(co *admin.API, uid string) error {
	deadline := time.Now().Add(20 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	u, e := GetUserInfo(co, ctx, uid)
	if e != nil {
		return e
	}

	// Если caps пустой значит удалять нечего
	if len(u.Caps) == 0 {
		return nil
	}

	// Парсим user caps для формирования строки
	uc := ""
	for _, caps := range u.Caps {
		uc += caps.Type + "=" + caps.Perm + ";"
	}
	uc = uc[0 : len(uc)-1] // отрезаем последний сисмвол

	_, e = co.RemoveUserCap(ctx, uid, uc)
	return e
}

// Метод создания нового ключа пользователя
func CreateKey(co *admin.API, uid string, keytype string, format string, subuser string) error {
	generatekey := true
	deadline := time.Now().Add(20 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// Для идентификации нового ключа сначала получаем список существующих ключей пользователя
	old_keys := map[string]string{}
	u, e := co.GetUser(ctx, admin.User{
		ID: uid,
	})
	if e != nil {
		return e
	}
	for _, key := range u.Keys {
		old_keys[key.AccessKey] = key.SecretKey
	}

	// Создаем новый ключ
	uk := admin.UserKeySpec{
		UID:         uid,
		KeyType:     keytype,
		GenerateKey: &generatekey,
	}
	if subuser != "" {
		uk.SubUser = subuser
	}
	_, e = co.CreateKey(ctx, uk)
	if e != nil {
		return e
	}

	// Получаем списко новых ключей
	new_keys := map[string]string{}
	u, e = co.GetUser(ctx, admin.User{
		ID: uid,
	})
	if e != nil {
		return e
	}
	for _, key := range u.Keys {
		new_keys[key.AccessKey] = key.SecretKey
	}

	// Сраниваем старые и новые ключи для получения созданного ключа
	newkey_for_user := map[string]map[string]string{}
	for k := range new_keys {
		if _, ok := old_keys[k]; !ok {
			// Если новый ключ найден через различие между old и new мапами
			// то выводим данный ключ - подсвечиваем его
			a := map[string]string{}
			a[k] = new_keys[k]
			newkey_for_user[uid] = a
			// Выводим в json
			if format == "json" {
				b, e := json.Marshal(newkey_for_user)
				if e != nil {
					return e
				}
				fmt.Println(string(b))
				return nil
			}
			// Выводим в table
			if format == "table" {
				headerFmt := color.New(color.FgGreen, color.Bold).SprintfFunc()
				columnFmt := color.New(color.FgHiYellow, color.Bold).SprintfFunc()

				tbl := table.New("UID", "ACCESS KEY", "SECRET KEY")
				tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithHeaderSeparatorRow('-').WithPadding(2)
				tbl.AddRow(uid, k, new_keys[k])
				tbl.Print()
			}
			return nil
		}
	}
	return fmt.Errorf("new key for user: %s not found!", uid)
	/*
			e = GetUserKeys(co, uid, format)
			if e != nil {
				return e
			}
		return nil
	*/
}

// Метод создания SubUser для основного пользователя
func CreateSubUser(co *admin.API, uid string, suidparam string, format string, hsd string, refresh bool, nocache bool) error {
	generatekey := true
	deadline := time.Now().Add(20 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// Парсим параметры subuser
	if len(strings.Split(suidparam, "|")) < 2 {
		return fmt.Errorf("failed parse suid user params - example \"<suid|[read or write or readwrite or full or none]\"")
	}
	suid := strings.Split(suidparam, "|")[0]
	access := strings.Split(suidparam, "|")[1]

	// Создаем subuser с переданными параметрами
	e := co.CreateSubuser(ctx, admin.User{ID: uid}, admin.SubuserSpec{
		Name:        suid,
		GenerateKey: &generatekey,
		Access: func() admin.SubuserAccess {
			switch access {
			case "read":
				return admin.SubuserAccessRead
			case "write":
				return admin.SubuserAccessWrite
			case "readwrite":
				return admin.SubuserAccessReadWrite
			case "full":
				return admin.SubuserAccessFull
			default:
				return admin.SubuserAccessNone
			}
		}(),
	})
	if e != nil {
		return e
	}

	// параметры печати результата
	if format == "table" {
		u, e := GetUserList(co, refresh, nocache)
		if e != nil {
			return e
		}
		UserTprint(*u, uid, hsd)
		return nil
	}
	if format == "json" {
		u, e := GetUserList(co, refresh, nocache)
		if e != nil {
			return e
		}
		e = UserJPrint(*u, uid)
		if e != nil {
			return e
		}
		return nil
	}

	return nil
}

// Метод удалению subusera из пользователя
func RemoveSubUser(co *admin.API, u_uid string, s_uid string) error {
	deadline := time.Now().Add(20 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// Получение данных пользователя и subuser
	user, e := GetUserInfo(co, ctx, u_uid)
	if e != nil {
		return e
	}
	subusers, e := func() (admin.SubuserSpec, error) {
		for _, su := range user.Subusers {
			if su.Name == s_uid {
				return su, nil
			}
		}
		return admin.SubuserSpec{}, fmt.Errorf("subuser not found")
	}()

	// Eдаление subusera из пользователя
	e = co.RemoveSubuser(ctx, user, subusers)
	if e != nil {
		return e
	}

	return nil
}

// Метод удаления пользователя S3
func DeleteUser(co *admin.API, uid string) error {
	deadline := time.Now().Add(20 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	// Получение данных пользователя и subuser
	user, e := GetUserInfo(co, ctx, uid)
	if e != nil {
		return e
	}

	e = co.RemoveUser(ctx, user)
	if e != nil {
		return e
	}
	return nil
}

// Метод смены user placement и user storage class
// -rgw -e ceph-single -uid a4f4073f-e80a-40bc-b416-1ac8fe5ae2b3 -cup -ptags HDD_PLACEMENT -dp STANDARD_IA -ds STANDARD_IA
func ChangeUserPlacementAndStorageClass(co *admin.API, uid string, tags string, placement string, storageclass string, fomat string, humanitySizeDisplay string, refresh bool, nocache bool) error {

	deadline := time.Now().Add(20 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// Получение данных пользователя и subuser
	user, e := GetUserInfo(co, ctx, uid)
	if e != nil {
		logrus.Error(e)
		return e
	}

	// Смена placement tags для пользователя
	// На текучий момент не поддерживается в ceph api
	//tags_slice := strings.Split(tags, ",")
	//user.PlacementTags = []interface{}{}
	//user.PlacementTags = append(user.PlacementTags, tags_slice)

	// Смена defaultPlacement и storageclass для пользователя
	user.DefaultPlacement = placement
	// storageclass не поддерживается на текучий момент в go-ceph v0.39.0 поддерджка начиная с v0.39.1-0.20260511125330-98f6cb17b27c
	user.DefaultStorageClass = storageclass

	u, e := co.ModifyUser(ctx, user)
	if e != nil {
		logrus.Error(e)
		return e
	}

	logrus.Info("Выполненна смена атрибутов пользователя")

	rgwusers, e := GetUserList(co, refresh, nocache)
	if e != nil {
		logrus.Error(e)
	}

	// Печать в виде json
	if fomat == "json" {
		e = UserJPrint(*rgwusers, u.ID)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
	}

	// Печать в табличном виде
	if fomat == "table" {
		UserTprint(*rgwusers, u.ID, humanitySizeDisplay)
	}
	return nil
}

package cephfsapi

import (
	"cephapi/db"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/ceph/go-ceph/rados"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/rodaine/table"
	"github.com/sirupsen/logrus"
)

// Метод создания новой CephFS
func CreateNewCephFS(conn *rados.Conn, fsName string, placement string, dmc string, ddc string, mpa string, dpa string) error {
	defer conn.Shutdown()
	if dmc == "" || ddc == "" || mpa == "" || dpa == "" || fsName == "" {
		return errors.New("ошибка обработки аргументов - аргументы переданны как пустые строки")
	}

	// Проверка на допустимые символы fsName
	matched, err := regexp.MatchString("^[a-zA-Z0-9_.-]+$", fsName)
	if err != nil || !matched {
		return fmt.Errorf("имя файловой системы '%s' содержит недопустимые символы. "+
			"Разрешены только буквы, цифры, точки, дефисы и подчеркивания", fsName)
	}

	// Проверка на максимальную длину fsName
	if len(fsName) > 256 {
		return fmt.Errorf("имя файловой системы слишком длинное: %d символов, максимум 256", len(fsName))
	}

	// Создание команды к MON серверу для создания новой CephFS
	cmd, e := json.Marshal(map[string]interface{}{
		"prefix":    "fs volume create",
		"name":      fsName,    // Имя создаваемой ФС CephFS
		"placement": placement, // Параметры размещения - по умолчанию размещается на 3 узлах
		// "pool_name": "custom_pool", // имя пула (по умолчанию cephfs.data и cephfs.meta)
	})

	if e != nil {
		logrus.Error("Ошибка исполнения:", e)
		return e
	}

	// Отправляем команду монитору
	response, info, e := conn.MonCommand(cmd)
	if e != nil {
		logrus.Errorf("Ошибка выполнения команды:\nОшибка: %v\nДетали: %s\nОтвет: %s",
			e, info, response)
		return e
	}
	logrus.Info("Выполненно создание cephfs \nResponce: ", response, "\nInfo: ", info)

	// Применяем правила размещения crush_rule для meta и data пуллов создаваемой CephFS
	e = ChangePoolCrushRule("cephfs."+fsName+".meta", dmc, conn)
	if e != nil {
		logrus.Error("Ошибка cмены crush_rule для pool : ", " cephfs."+fsName+".meta", " crush_rule: ", dmc)
		return e
	}
	e = ChangePoolCrushRule("cephfs."+fsName+".data", ddc, conn)
	if e != nil {
		logrus.Error("Ошибка cмены crush_rule для pool : ", " cephfs."+fsName+".data", " crush_rule: ", ddc)
		return e
	}

	// Применяем параметры атрибутов для meta и data пулов созданной ceph fs
	e = ChagePoolAttributes("cephfs."+fsName+".meta", mpa, conn)
	if e != nil {
		logrus.Error("Ошибка установки атрбутов для pool : ", " cephfs."+fsName+".meta", " атрибуты: ", mpa)
		return e
	}
	e = ChagePoolAttributes("cephfs."+fsName+".data", dpa, conn)
	if e != nil {
		logrus.Error("Ошибка установки атрбутов для pool : ", " cephfs."+fsName+".meta", " атрибуты: ", dpa)
		return e
	}
	return nil

}

// Метод смены crush_rule для pool
func ChangePoolCrushRule(pool string, crush_rule string, conn *rados.Conn) error {
	cmd, e := json.Marshal(map[string]interface{}{
		"prefix": "osd pool set",
		"pool":   pool,
		"var":    "crush_rule",
		"val":    crush_rule,
	})
	if e != nil {
		return e
	}
	response, info, e := conn.MonCommand(cmd)
	if e != nil {
		logrus.Error("Ошибка смены crush_rule: ", e, " response: ", response, " info: ", info)
		return e
	}
	logrus.Info("Смена crush_rule для pool ", pool, " выполненна. Response: ", response, "Info: ", info)
	return nil
}

// Метод установки параметров на cephfs pool
// attr - атрибуты pool "pg_autoscale|pg_num|compresion|quota_bytes|quota_objects"
func ChagePoolAttributes(pool string, attr string, conn *rados.Conn) error {
	// Парсим аргументы
	if attr == "" {
		return errors.New("Ошибка парсинга параметров. Парамтеры не переданны")
	}
	args := strings.Split(attr, "|")
	if len(args) < 5 {
		return errors.New("Ошибка парсинга параметров")
	}

	// Установка параметров autoscale ceph osd pool set {pool-name} pg_autoscale_mode off|on
	cmd, e := json.Marshal(map[string]interface{}{
		"prefix": "osd pool set",
		"pool":   pool,
		"var":    "pg_autoscale_mode",
		"val":    args[0],
	})
	if e != nil {
		logrus.Error("Ошибка исполнения:", e)
		return e
	}
	response, info, e := conn.MonCommand(cmd)
	if e != nil {
		logrus.Error("Ошибка смены autoscale: ", e, " response: ", response, " info: ", info)
		return e
	}
	logrus.Info("Смена autoscale для pool ", pool, " выполненна. Response: ", response, "Info: ", info)

	// Смена pg_num ceph osd pool set {pool-name} pg_num {new-pg-num}
	cmd, e = json.Marshal(map[string]interface{}{
		"prefix": "osd pool set",
		"pool":   pool,
		"var":    "pg_num",
		"val":    args[1],
	})
	if e != nil {
		logrus.Error("Ошибка исполнения:", e)
		return e
	}
	response, info, e = conn.MonCommand(cmd)
	if e != nil {
		logrus.Error("Ошибка смены pg_num: ", e, " response: ", response, " info: ", info)
		return e
	}
	logrus.Info("Смена pg_num для pool ", pool, " выполненна. Response: ", response, "Info: ", info)

	// Смена compresion ceph osd pool set {pool-name} compression_algorithm none|
	cmd, e = json.Marshal(map[string]interface{}{
		"prefix": "osd pool set",
		"pool":   pool,
		"var":    "compression_algorithm",
		"val":    args[2],
	})
	if e != nil {
		logrus.Error("Ошибка исполнения:", e)
		return e
	}
	response, info, e = conn.MonCommand(cmd)
	if e != nil {
		logrus.Error("Ошибка смены compression_algorithm: ", e, " response: ", response, " info: ", info)
		return e
	}
	logrus.Info("Смена compresion для pool ", pool, " выполненна. Response: ", response, "Info: ", info)

	// Установка квот на pool max_bytes
	cmd, e = json.Marshal(map[string]interface{}{
		"prefix": "osd pool set-quota",
		"pool":   pool,
		"field":  "max_bytes",
		"val":    args[3],
	})
	response, info, e = conn.MonCommand(cmd)
	if e != nil {
		logrus.Error("Ошибка смены max_bytes: ", e, " response: ", response, " info: ", info)
		return e
	}
	logrus.Info("Смена max_bytes для pool ", pool, " выполненна. Response: ", response, "Info: ", info)

	// Установка квот на pool max_objects
	cmd, e = json.Marshal(map[string]interface{}{
		"prefix": "osd pool set-quota",
		"pool":   pool,
		"field":  "max_objects",
		"val":    args[4],
	})
	response, info, e = conn.MonCommand(cmd)
	if e != nil {
		logrus.Error("Ошибка смены max_objects: ", e, " response: ", response, " info: ", info)
		return e
	}
	logrus.Info("Смена max_objects для pool ", pool, " выполненна. Response: ", response, "Info: ", info)

	return nil
}

// Метод получения списка cephfs
func GetCephFsList(conn *rados.Conn, format string, fsname string, nocache bool, refresh bool) error {
	// Получаем список всех pool
	cmd, e := json.Marshal(map[string]interface{}{
		"prefix": "osd pool ls",
		"detail": "detail",
		"format": "json",
	})

	if e != nil {
		logrus.Error("Ошибка исполнения:", e)
		return e
	}

	// Генерация ключа для сохранения в БД
	fsid, e := conn.GetFSID()
	if e != nil {
		return e
	}

	var response []byte
	var info string
	key := uuid.NewSHA1(uuid.NameSpaceX500, []byte(fsid+"GetCephFsList"))

	// тут реализуем выгрузку из БД с проверкой по времени
	// результатами выгрузки будет response
	// если данные в БД не актуальны или не найдены то идем к API
	value, updated_at, e := db.GetResponse(key.String())
	if e != nil || refresh {
		response, info, e = conn.MonCommand(cmd)
		if e != nil {
			logrus.Error("Ошибка получения списка pool: ", e, " response: ", response, " info: ", info)
			return e
		}
		if !nocache {
			if e != nil {
				logrus.Warn("Failed marshal data: ", e)
			} else {
				db.SaveResponse(key.String(), response)
			}
		}
	} else {
		if time.Now().After(updated_at.Add(2 * time.Minute)) {
			response, info, e = conn.MonCommand(cmd)
			if e != nil {
				logrus.Error("Ошибка получения списка pool: ", e, " response: ", response, " info: ", info)
				return e
			}
			if !nocache {
				if e != nil {
					logrus.Warn("Failed marshal data: ", e)
				} else {
					db.SaveResponse(key.String(), response)
				}
			}
		} else {
			e = json.Unmarshal(value.([]byte), &response)
			if e != nil {
				response, info, e = conn.MonCommand(cmd)
				if e != nil {
					logrus.Error("Ошибка получения списка pool: ", e, " response: ", response, " info: ", info)
					return e
				}
				// Если не передан флаг не кешировать запрос то кешируем в БД
				if !nocache {
					if e != nil {
						logrus.Warn("Failed marshal data: ", e)
					} else {
						db.SaveResponse(key.String(), response)
					}
				}
			}
		}
	}

	// Определяем volume относячиеся к cephfs
	var pools []PoolInfo
	e = json.Unmarshal(response, &pools)
	if e != nil {
		logrus.Error("error unmarshal json data: ", e)
		return e
	}

	if format == "json" {
		if fsname != "" {
			// Находим нужную CephFS и выводим ее в виде json
			out := []PoolInfo{}
			for _, p := range pools {
				// Если в имение pool нет fsname то сразу пропускаем
				if !strings.Contains(p.PoolName, fsname) {
					continue
				}

				if p.ApplicationMetadata.Cephfs.Data == fsname {
					// Опердеяем наличие subvulumes для обрабатываемой cephfs
					var subvolumes []Subvolume
					cmd, e := json.Marshal(map[string]interface{}{
						"prefix":   "fs subvolume ls",
						"vol_name": p.ApplicationMetadata.Cephfs.Data,
					})
					if e != nil {
						logrus.Error("Ошибка исполнения:", e)
						return e
					}
					response, info, e := conn.MonCommand(cmd)
					if e != nil {
						logrus.Error("Ошибка смены crush_rule: ", e, " response: ", response, " info: ", info)
						return e
					}
					e = json.Unmarshal(response, &subvolumes)
					if e != nil {
						logrus.Error("Ошибка исполнения:", e)
						return e
					}
					p.Subvolumes = subvolumes
					out = append(out, p)
				}
				if p.ApplicationMetadata.Cephfs.MetaData == fsname {
					out = append(out, p)
				}
			}

			b, e := json.Marshal(out)
			if e != nil {
				log.Println("error marchal json data")
				return e
			}
			fmt.Println(string(b))
			return nil
		}
		fmt.Println(string(response))
	} else if format == "table" {
		headerFmt := color.New(color.FgGreen, color.Bold).SprintfFunc()
		columnFmt := color.New(color.FgHiYellow, color.Bold).SprintfFunc()
		tbl := table.New("FSNAME", "SUBVOLUMES", "POOLS", "CACHE MODE", "PGNUM", "PG PENDING", "PG TARGET", "PGAUTOSCALE", "CRUSH RULE ID", "FLAGS", "MAX BYTES", "SIZE", "CREATE TIME")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithHeaderSeparatorRow('-').WithPadding(1)
		// еребираем все pool и выбираем только с идентифкатором CephFS
		for _, p := range pools {
			if fsname != "" {
				if !strings.Contains(p.PoolName, fsname) {
					continue
				}
			}

			if p.ApplicationMetadata.Cephfs.Data != "" {
				// Опердеяем наличие subvulumes для обрабатываемой cephfs
				var subvolumes []Subvolume
				cmd, e := json.Marshal(map[string]interface{}{
					"prefix":   "fs subvolume ls",
					"vol_name": p.ApplicationMetadata.Cephfs.Data,
				})
				if e != nil {
					logrus.Error("Ошибка исполнения:", e)
					return e
				}
				response, info, e := conn.MonCommand(cmd)
				if e != nil {
					logrus.Error("Ошибка смены crush_rule: ", e, " response: ", response, " info: ", info)
					return e
				}
				e = json.Unmarshal(response, &subvolumes)
				if e != nil {
					logrus.Error("Ошибка исполнения:", e)
					return e
				}

				tbl.AddRow(p.ApplicationMetadata.Cephfs.Data, func() string {
					str := ""
					for _, s := range subvolumes {
						str = str + s.Name + ";"
					}
					if str == "" {
						return str
					}
					str = str[:len(str)-1]
					if len(str) > 10 {
						str = str[:10] + "..."
					}
					return str
				}(), p.PoolName, p.CacheMode,
					p.PGNum, p.PGNumPending, p.PGNumTarget, p.PGAutoscaleMode, p.CrushRule, p.FlagsNames, p.QuotaMaxBytes, p.Size,
					func() string {
						t, e := time.Parse("2006-01-02T15:04:05.999999-0700", p.CreateTime)
						if e != nil {
							return ""
						}
						return t.Format("2006-01-02 15:04")
					}(),
				)
			}
			if p.ApplicationMetadata.Cephfs.MetaData != "" {
				tbl.AddRow("", "", p.PoolName, p.CacheMode,
					p.PGNum, p.PGNumPending, p.PGNumTarget, p.PGAutoscaleMode, p.CrushRule, p.FlagsNames, p.QuotaMaxBytes, p.Size,
					func() string {
						t, e := time.Parse("2006-01-02T15:04:05.999999-0700", p.CreateTime)
						if e != nil {
							return ""
						}
						return t.Format("2006-01-02 15:04")
					}())
			}

		}
		tbl.Print()
	}

	return nil
}

// Метод получения списка subvolume указанной cephfs
func GetCephfsSubvolumeList(conn *rados.Conn, format string, fsname string) error {
	// Получаем список всех subvolumes указанной cephfs
	cmd, e := json.Marshal(map[string]interface{}{
		"prefix":   "fs subvolume ls",
		"vol_name": fsname,
	})

	if e != nil {
		logrus.Error("Ошибка исполнения:", e)
		return e
	}

	response, info, e := conn.MonCommand(cmd)
	if e != nil {
		logrus.Error("Ошибка получения списка subvolumes: ", e, " response: ", response, " info: ", info)
		return e
	}

	var subvolumes []Subvolume
	e = json.Unmarshal(response, &subvolumes)
	if e != nil {
		logrus.Error("error unmarchal data!: ", e)
		return e
	}

	// олучаем подробную информацию о каждом usbvolume
	for i, subvolume := range subvolumes {
		cmd, e := json.Marshal(map[string]interface{}{
			"prefix":   "fs subvolume info",
			"vol_name": fsname,
			"sub_name": subvolume.Name,
		})

		if e != nil {
			logrus.Error("Ошибка исполнения:", e)
			return e
		}

		response, info, e := conn.MonCommand(cmd)
		if e != nil {
			logrus.Error("Ошибка получения списка subvolumes: ", e, " response: ", response, " info: ", info)
			return e
		}
		var subvolumeinfo SubvolumeInfo
		e = json.Unmarshal(response, &subvolumeinfo)
		if e != nil {
			logrus.Error("Ошибка исполнения:", e)
			return e
		}
		subvolumes[i].Info = subvolumeinfo
	}

	if format == "table" {
		// Формируем таблицу с информацией о subvolumes
		headerFmt := color.New(color.FgGreen, color.Bold).SprintfFunc()
		columnFmt := color.New(color.FgHiYellow, color.Bold).SprintfFunc()
		tbl := table.New("SUB NAME", "BYTES QUOTA", "BYTES USED", "USED PERCENT", "DATA POOL", "GID", "ATIME", "MTIME")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithHeaderSeparatorRow('-').WithPadding(1)
		for _, subvolume := range subvolumes {
			tbl.AddRow(subvolume.Name, subvolume.Info.BytesQuota, subvolume.Info.BytesUsed,
				subvolume.Info.BytesPcent, subvolume.Info.DataPool, subvolume.Info.Gid, subvolume.Info.Atime, subvolume.Info.Mtime,
			)
		}
		tbl.Print()
	} else if format == "json" {
		b, e := json.Marshal(subvolumes)
		if e != nil {
			logrus.Error("error unmarchal data!: ", e)
			return e
		}
		fmt.Println(string(b))
	}
	return nil
}

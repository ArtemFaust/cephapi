package main

import (
	"cephapi/cephfsapi"
	"cephapi/db"
	"cephapi/radosapi"
	"cephapi/rgwapi"
	"cephapi/utils"
	"flag"
	"fmt"
	"os"

	"github.com/ceph/go-ceph/rados"
	"github.com/sirupsen/logrus"
)

// Внешние глобальные переменные
var (
	Akey          string // AccessKey rgw
	SKey          string // SecretKey rgw
	S3Endpoint    string // Endpoint rgw
	EndpointsPath string // Путь к файлу конфигурации endpoints.yaml
)

var (
	// Ключи связанные с шифрование и дешифрованием файла endpoints
	Encode *bool // Зашифровать фаил подключений
	Decode *bool // Расшифровать фаил подключений
	// Верхнии ключи запуска
	ManageMenu          *bool   // Ключ запуска менеджера конфигурации
	Interactive         *bool   // Ключ перехода в интерактивный режим
	RGW                 *bool   // Ключ api RGW
	RADOS               *bool   // Ключ api RADOS
	HumanitySizeDisplay *string // Ключ переключения отображения size (kb gb tb)
	// Обющие ключи
	UserCups       *string // Ключ описания CUPS пользователя
	ChangeUserCaps *bool   // Ключ изменения пользовательских CAPS
	EndPoint       *string // Ключ описания endpoint
	CreateUser     *bool   // Ключ создания нового пользователя
	RemoveUser     *bool   // Ключ удаления пользователя S3/Ceph
	GetUser        *bool   // Ключ получения информации о пользователе
	Refresh        *bool   // Ключ указывающий на необходимость принудительного обновления данных через API минуя обращения к БД
	NoCache        *bool   // Ключ отключения кеширования результата выполнения запроса
	// Ключи операций RGW
	GetUserList    *bool   // Ключ операции получения списка пользователей
	Format         *string // Ключ формата вывода
	User           *string // Ключ uid пользователя
	SUser          *string // Ключ uid subuser основного пользователя
	Examples       *bool   // Ключ выводв примеров использования
	GetBucketList  *bool   // Ключ операции получения списка бакетов
	CreateSubUser  *bool   // Ключ создания subuser
	RgwUserParam   *string // Параметры создаваемого пользователя
	RemoveUserCaps *bool   // Ключ удаление пользовательских CAPS
	RemoveSubuser  *bool   // Ключ удаления subuser у пользователя
	CreateKey      *bool   // Ключ создания нового ключа пользователя
	KeyType        *string // Ключ указывающий на тип ключа
	BucketName     *string // Ключ имени бакета
	// Ключи RGW связанные с placement
	ChangeUserPlacement *bool   // Ключ вызова операции смены placement для rgw пользователя
	PlacementTags       *string // Ключ указания placement tags для пользователя
	DefaultPlacement    *string // Ключ указания placement по умолчанию для пользователя
	DefaultStorageClass *string // Ключ укзания storage_class по умолчанию для пользователя
	// Ключи операций RADOS
	// Операции работы с кластером
	GetCluInfo *bool // Получение информации о кластере
	// Операции CEPHFS
	CreateNewCepfFS      *bool   // Ключ создания новой cephfs
	ChangePoolCrushRule  *bool   // Ключ оперции смены crush_rule для pool
	ChangePoolAttributes *bool   // Ключ смены атрибутов для pool
	Pool                 *string // Имя пула для выполнения операций
	CrushRule            *string // Название crush rule
	Placement            *string // Ключ указания placement размечения cephfs
	CephfsName           *string // Имя создаваемой cephfs
	DefaultMetaCrushRule *string // Название crush rule для размещения meta пула новой cephFS
	DefaultDataCrushRule *string // Название crush rule для размещения data пула новой cephFS
	MetaPoolAttributes   *string // Атрибуты пула meta
	DataPoolAttributes   *string // Атрибуты пула data
	// Операции с пользовтелями CEPH
	CephUserEntryName *string // Имя пользователя ceph
	GetFsList         *bool   // Ключ получения списка cephfs
	GetSubvolumeList  *bool   // Ключ операции получения списка subvolume указанной cephfs
)

// Инициализация парсинга аргументов командной строки
func init() {
	// Создание БД хранения результатов запроса
	db.CreateKVdatabase()
	// Парсинг аргементов
	Encode = flag.Bool("encode", false, "Encode endpoints.yaml")
	Decode = flag.Bool("decode", false, "Decode and print endpoints.yaml")
	GetUserList = flag.Bool("lu", false, "List RGW user - for json print +buckets info")
	GetUser = flag.Bool("guk", false, "Get S3 user keys or RADOS user info")
	GetBucketList = flag.Bool("lb", false, "List RGW buckets - only table print")
	Format = flag.String("format", "json", "Print format json,table")
	RGW = flag.Bool("rgw", false, "RGW API operations")
	RADOS = flag.Bool("rados", false, "RADOS API operations")
	User = flag.String("uid", "", "User uid - use only table format print")
	SUser = flag.String("suid", "", "Subuser uid")
	EndPoint = flag.String("e", "", "Selected endpoint")
	Examples = flag.Bool("examples", false, "Print usage examples")
	CreateUser = flag.Bool("cu", false, "Create new user")
	CreateSubUser = flag.Bool("cus", false, "Create new subuser for rgw user")
	RgwUserParam = flag.String("rup", "", "RGW user parameters")
	UserCups = flag.String("caps", "", `User cups by example:
		 for RGW "buckets=*;users=*;usage=read;metadata=read;zone=read"
		 for RADOS "mon==allow *;mds==allow *; osd==allow *"`)
	ChangeUserCaps = flag.Bool("cuc", false, "Change user cups")
	RemoveUserCaps = flag.Bool("ruc", false, "Remove user caps")
	RemoveSubuser = flag.Bool("rsu", false, "Remove S3 subuser")
	RemoveUser = flag.Bool("ru", false, "Remove S3/Ceph user")
	CreateKey = flag.Bool("ck", false, "Create new user key paire")
	KeyType = flag.String("kt", "s3", "Key type s3 or swift")
	Placement = flag.String("placement", "3", "Placement group")
	CephfsName = flag.String("fsname", "", "CephFS name")
	CreateNewCepfFS = flag.Bool("cf", false, "Create new cephfs")
	DefaultMetaCrushRule = flag.String("dmc", "", "CephFS meta pool crush rule")
	DefaultDataCrushRule = flag.String("ddc", "", "CephFS data pool crush rule")
	MetaPoolAttributes = flag.String("mpa", "off|16|none|0|0", "CephFS meta pool attributes - see examples")
	DataPoolAttributes = flag.String("dpa", "off|32|none|0|0", "CephFS data pool crush rule - see examples")
	ChangePoolCrushRule = flag.Bool("cpc", false, "Change pool crush rule")
	Pool = flag.String("pool", "", "Pool name - example cephfs.fs.data or rbd .... etc")
	CrushRule = flag.String("crush_rule", "", "Crush rule")
	ChangePoolAttributes = flag.Bool("cpa", false, "Change pool attributes")
	CephUserEntryName = flag.String("user", "", "Ceph user entry name")
	GetFsList = flag.Bool("gf", false, "Get cephfs list")
	GetSubvolumeList = flag.Bool("gfs", false, "get cephfs subvolume list")
	GetCluInfo = flag.Bool("ci", false, "Get cluster info")
	HumanitySizeDisplay = flag.String("size", "kb", "Display size in kb|gb|tb")
	Interactive = flag.Bool("i", false, "launch interactive mod")
	ManageMenu = flag.Bool("manage", false, "manage menu")
	BucketName = flag.String("bucketname", "", "Bucket name")
	ChangeUserPlacement = flag.Bool("cup", false, "Change RGW user placement")
	PlacementTags = flag.String("ptags", "", "rgw user placement tags")
	DefaultPlacement = flag.String("dp", "", "rgw user default placement")
	DefaultStorageClass = flag.String("ds", "", "rgw user default storage class")
	Refresh = flag.Bool("refresh", false, "refresh data over api")
	NoCache = flag.Bool("nocache", false, "disable cache request result to local db")
	flag.Parse()
}

func main() {
	// Параметры логера
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		TimestampFormat:  "2006-01-02 15:04:05.000",
		ForceColors:      true, // Цвета в терминале
		PadLevelText:     true, // Выравнивание уровней
		QuoteEmptyFields: true, // Кавычки для пустых полей
	})

	// Если передан ключ Encode шифруем фаил endpoints.yaml
	if *Encode {
		EndpointsPath, e := utils.FindEndpointsConfig()
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		_, e = os.Stat(EndpointsPath)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		b, e := os.ReadFile(EndpointsPath)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		eb, e := utils.EncodeFile(b)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		e = os.WriteFile(EndpointsPath, eb, 0755)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		return
	}

	// Если передан ключ декодирования файла endpoints.yaml
	if *Decode {
		EndpointsPath, e := utils.FindEndpointsConfig()
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		_, e = os.Stat(EndpointsPath)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		b, e := os.ReadFile(EndpointsPath)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		db, e := utils.DecodeFile(b)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		fmt.Println(string(db))
		return
	}

	// Запуск меню управления подлкючениями
	if *ManageMenu {
		utils.ManageMenu()
	}

	// Если передан ключ перехода в интерактивный режим
	if *Interactive {
		utils.InteractiveLaunch(*Refresh, *NoCache)
	}

	// Считывание файла описания endpoints
	endpoints, e := utils.Endpoints()
	if e != nil {
		logrus.Error("Ошибка считывания конфигурации endpoints: ", e)
		os.Exit(1)
	}

	switch {
	// Операции связанные с RGW API
	case *RGW:
		{
			// Определяем параметры подключения по переданному endpoint
			for _, en := range endpoints.Rgw.RGW.EndPoints {
				if en.Name != *EndPoint {
					continue
				}
				Akey = en.AcessKey
				SKey = en.SecretKey
				S3Endpoint = en.S3Endpoint
			}

			// Если ключи не найдены то выход
			if Akey == "" && SKey == "" && S3Endpoint == "" {
				logrus.Warn("Endpoint not found in configuration!")
				os.Exit(1)
			}

			// Операция получения списка пользователей RGW
			if *GetUserList {
				c, e := rgwapi.MakeNewConnection(S3Endpoint, Akey, SKey)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				rgwusers, e := rgwapi.GetUserList(c, *Refresh, *NoCache)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				// Печать в виде json
				if *Format == "json" {
					e = rgwapi.UserJPrint(*rgwusers, *User)
					if e != nil {
						logrus.Error(e)
						os.Exit(1)
					}
				}
				// Печать в табличном виде
				if *Format == "table" {
					rgwapi.UserTprint(*rgwusers, *User, *HumanitySizeDisplay)
				}
				return
			}
			// Операция получения списка buckets RGW
			if *GetBucketList {
				c, e := rgwapi.MakeNewConnection(S3Endpoint, Akey, SKey)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				rgwusers, e := rgwapi.GetUserList(c, *Refresh, *NoCache)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				// Печать в виде json
				if *Format == "json" {
					rgwapi.PrintJBuckets(c, *rgwusers, *User)
				}
				// Печать в табличном виде
				if *Format == "table" {
					rgwapi.PrintTBuckets(c, *rgwusers, *User, *HumanitySizeDisplay, *BucketName)
				}
				return
			}
			// Создание нового пользователя
			if *CreateUser && *RgwUserParam != "" {
				c, e := rgwapi.MakeNewConnection(S3Endpoint, Akey, SKey)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				e = rgwapi.CreateRgwUser(c, *RgwUserParam, *Format, *UserCups, *HumanitySizeDisplay, *Refresh, *NoCache)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}
			// Получение информации о пользоватееле
			if *GetUser && *User != "" {
				c, e := rgwapi.MakeNewConnection(S3Endpoint, Akey, SKey)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				e = rgwapi.GetUserKeys(c, *User, *Format)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}
			// Изменение CAPS пользователя
			if *ChangeUserCaps && *User != "" {
				c, e := rgwapi.MakeNewConnection(S3Endpoint, Akey, SKey)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				e = rgwapi.SetUserCaps(c, *User, *UserCups, *Format)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}
			// Удаление CAPS пользоватея
			if *RemoveUserCaps && *User != "" {
				c, e := rgwapi.MakeNewConnection(S3Endpoint, Akey, SKey)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				e = rgwapi.RemoveUserCaps(c, *User)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}
			// Создание нового ключа пользователя
			if *CreateKey && *KeyType != "" && *User != "" {
				c, e := rgwapi.MakeNewConnection(S3Endpoint, Akey, SKey)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				e = rgwapi.CreateKey(c, *User, *KeyType, *Format, *SUser)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}
			// Создание subuser для пользователя
			if *CreateSubUser && *User != "" && *RgwUserParam != "" {
				c, e := rgwapi.MakeNewConnection(S3Endpoint, Akey, SKey)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				e = rgwapi.CreateSubUser(c, *User, *RgwUserParam, *Format, *HumanitySizeDisplay, *Refresh, *NoCache)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}
			// Удаление subuser пользовптеля
			if *RemoveSubuser && *User != "" && *SUser != "" {
				c, e := rgwapi.MakeNewConnection(S3Endpoint, Akey, SKey)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				e = rgwapi.RemoveSubUser(c, *User, *SUser)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				rgwusers, e := rgwapi.GetUserList(c, *Refresh, *NoCache)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				// Печать в виде json
				if *Format == "json" {
					e = rgwapi.UserJPrint(*rgwusers, *User)
					if e != nil {
						logrus.Error(e)
						os.Exit(1)
					}
				}
				// Печать в табличном виде
				if *Format == "table" {
					rgwapi.UserTprint(*rgwusers, *User, *HumanitySizeDisplay)
				}
				return
			}
			// Удаление пользователя S3
			if *RemoveUser && *User != "" {
				c, e := rgwapi.MakeNewConnection(S3Endpoint, Akey, SKey)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				e = rgwapi.DeleteUser(c, *User)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}
			// Смена placement и storageclass пользователя
			if *ChangeUserPlacement && *User != "" {
				c, e := rgwapi.MakeNewConnection(S3Endpoint, Akey, SKey)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				e = rgwapi.ChangeUserPlacementAndStorageClass(c, *User, *PlacementTags, *DefaultPlacement, *DefaultStorageClass, *Format, *HumanitySizeDisplay, *Refresh, *NoCache)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}
		}
		// Операции связанные с RADOS API
	case *RADOS:
		// кземпляр подключения используемый в операциях с rados
		var conn *rados.Conn
		defer func() {
			if conn != nil {
				conn.Shutdown()
			}
		}()
		{
			// Определяем параметры подключения по переданному endpoint
			for i, en := range endpoints.Rados.RADOS.EndPoints {
				if en.Name != *EndPoint {
					if i+1 == len(endpoints.Rados.RADOS.EndPoints) {
						logrus.Warn("Endpoint not found in configuration!")
						os.Exit(1)
					}
					continue
				}

				// Определяем параметры crush_rule для пулов
				if *DefaultMetaCrushRule == "" {
					DefaultMetaCrushRule = &en.DefaultMetaCrushRule
				}
				if *DefaultDataCrushRule == "" {
					DefaultDataCrushRule = &en.DefaultDataCrushRule
				}

				// Создаем новое подключение
				conn, e = radosapi.MakeNewRadosConnection(en.Fsid, en.MonHosts, en.KeyRing)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
			}

			// Создание новой cephfs
			if *CreateNewCepfFS && *CephfsName != "" {
				e = cephfsapi.CreateNewCephFS(conn, *CephfsName, *Placement, *DefaultMetaCrushRule, *DefaultDataCrushRule, *MetaPoolAttributes, *DataPoolAttributes)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}

			// Смена crush rule для существующего pool
			if *ChangePoolCrushRule && *Pool != "" && *CrushRule != "" {
				e = cephfsapi.ChangePoolCrushRule(*Pool, *CrushRule, conn)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}

			// Смена атрибутов существующего pool
			if *ChangePoolAttributes && *Pool != "" {
				flag.Visit(func(f *flag.Flag) {
					// Если передан флаг mpa
					if f.Name == "mpa" {
						e := cephfsapi.ChagePoolAttributes(*Pool, *MetaPoolAttributes, conn)
						if e != nil {
							logrus.Error(e)
							os.Exit(1)
						}
					}
					// Если передан флаг dpa
					if f.Name == "dpa" {
						e := cephfsapi.ChagePoolAttributes(*Pool, *DataPoolAttributes, conn)
						if e != nil {
							logrus.Error(e)
							os.Exit(1)
						}
					}
				})
				return
			}

			// Создание пользователя ceph
			if *CreateUser && *CephUserEntryName != "" {
				e = cephfsapi.CreateNewCephUser(conn, *CephUserEntryName, *UserCups)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}

			// Установка/обновление caps пользователя
			if *ChangeUserCaps && *UserCups != "" && *CephUserEntryName != "" {
				e = cephfsapi.SetCephUserCaps(conn, *CephUserEntryName, *UserCups)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}

			// Удаление пользователя ceph
			if *RemoveUser && *CephUserEntryName != "" {
				e = cephfsapi.DeleteCephUser(conn, *CephUserEntryName)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}

			// Получение информации о пользователе
			if *GetUser && *CephUserEntryName != "" {
				e = cephfsapi.GetUserInfo(conn, *CephUserEntryName)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}

			// Получение скписка cephfs
			if *GetFsList {
				e = cephfsapi.GetCephFsList(conn, *Format, *CephfsName, *NoCache, *Refresh)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}

			// Получения списка subvolume указанной cephfs
			if *GetSubvolumeList && *CephfsName != "" {
				e = cephfsapi.GetCephfsSubvolumeList(conn, *Format, *CephfsName)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}

			// Получение информации о кластере
			if *GetCluInfo {
				e = radosapi.GetClusterInfo(conn, *Format)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
				return
			}
		}
	case *Examples:
		{
			utils.PrintUsageExamples()
			return
		}
	}
}

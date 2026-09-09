package utils

import (
	"cephapi/cephfsapi"
	"cephapi/radosapi"
	"cephapi/rgwapi"
	"errors"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"

	// Данная библиотека go-yaml позволяет сохронить порядок полей при сериализации
	"github.com/goccy/go-yaml"
	"github.com/google/uuid"
	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
)

// Метод инициалиазции интерактивного режима работы
func InteractiveLaunch(refresh bool, nocache bool) {
	printMinHelp()
	enpoint_value := selectEndpointMenu()
	if enpoint_value.Operation == "RGW" {
		placeholder := fmt.Sprintf("\033[35m%s >\033[0m", enpoint_value.Endpoints.Rgw.Name)
		rgwenpoint_operation := rgwOperationsMenu(placeholder)

		// Создание экземпляра RGW подключения
		c, e := rgwapi.MakeNewConnection(enpoint_value.Endpoints.Rgw.S3Endpoint, enpoint_value.Endpoints.Rgw.AcessKey, enpoint_value.Endpoints.Rgw.SecretKey)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		if rgwenpoint_operation == "Получение списка пользователей RGW или отдельного пользователя по его UID" {
			rgwusers, e := rgwapi.GetUserList(c, refresh, nocache)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Выбор метода печати
			print_methods := []string{"table", "json"}
			print_method, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(print_methods)).
				WithDefaultText(placeholder + " Выбор метода печати").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Вввод uid пользователя
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = placeholder + " Введите uid пользователя для фильтрации или оставьте пустым "
			user_uid, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Печать в виде json
			if print_method == "json" {
				e = rgwapi.UserJPrint(*rgwusers, user_uid)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
			}

			// Печать в табличном виде
			if print_method == "table" {
				// Вввод отображения size
				size, e := pterm.DefaultInteractiveSelect.
					WithOptions([]string{"kb", "gb", "tb"}).
					WithDefaultText(placeholder + " Выбор отображения поля size").
					Show()
				if e != nil {
					logrus.Error("Ошибка: ", e)
					os.Exit(1)
				}
				rgwapi.UserTprint(*rgwusers, user_uid, size)
			}
			return
		}
		if rgwenpoint_operation == "Получение списка бакетов" {
			rgwusers, e := rgwapi.GetUserList(c, refresh, nocache)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Выбор метода печати
			print_methods := []string{"table", "json"}
			print_method, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(print_methods)).
				WithDefaultText(placeholder + " Выбор метода печати").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Вввод uid пользователя
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = placeholder + " Введите uid пользователя или оставьте пустым"
			user_uid, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Вввод названия бакета
			inputBucketName := pterm.DefaultInteractiveTextInput
			inputBucketName.DefaultText = placeholder + " Введите название бакета или оставьте пустым"
			bucketname, e := inputBucketName.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Печать в виде json
			if print_method == "json" {
				rgwapi.PrintJBuckets(c, *rgwusers, user_uid)
			}
			// Печать в табличном виде
			if print_method == "table" {
				// Вввод отображения size
				size, e := pterm.DefaultInteractiveSelect.
					WithOptions([]string{"kb", "gb", "tb"}).
					WithDefaultText("Выбор отображения поля size").
					Show()
				if e != nil {
					logrus.Error("Ошибка: ", e)
					os.Exit(1)
				}
				rgwapi.PrintTBuckets(c, *rgwusers, user_uid, size, bucketname)
			}
			return
		}
		if rgwenpoint_operation == "Создание нового пользователя S3" {
			rgwUserParam := ""
			// Вввод RgwUserParam пользователя
			uuidInput := pterm.DefaultInteractiveTextInput
			uuidInput.DefaultText = "    └─UUID пользователя (оставьте пустым для генерации)"
			userUuid, e := uuidInput.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			displayNameInput := pterm.DefaultInteractiveTextInput
			displayNameInput.DefaultText = "    └─Отображаемое имя пользователя (обязательное поле)"
			displayName, e := displayNameInput.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if displayName == "" {
				logrus.Error("Отображаемое имя не может быть пустым!")
				os.Exit(1)
			}

			userEmailInput := pterm.DefaultInteractiveTextInput
			userEmailInput.DefaultText = "    └─Email пользователя"
			userEmail, e := userEmailInput.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			userQuotaInput := pterm.DefaultInteractiveTextInput
			userQuotaInput.DefaultText = "    └─Квота пользователя в kb (пусто если ковота не устанавливается)"
			userQuota, e := userQuotaInput.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			objectQuotaInput := pterm.DefaultInteractiveTextInput
			objectQuotaInput.DefaultText = "    └─Квота объектов (пусто если ковота не устанавливается)"
			objectQuota, e := objectQuotaInput.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if userUuid == "" {
				u, e := uuid.NewV7()
				if e != nil {
					logrus.Error("Ошибка генерации UUID пользователя: ", e)
					os.Exit(1)
				}
				userUuid = u.String()
			}

			rgwUserParam = fmt.Sprintf("%s|%s|%s|%s|%s|", userUuid, displayName, userEmail, userQuota, objectQuota)

			// Вввод UserCups пользователя
			userCupsInput := pterm.DefaultInteractiveTextInput
			userCupsInput.DefaultText = "    └─CUPS создаваемого пользователя "
			userCups, e := userCupsInput.
				WithDefaultValue("buckets=*;users=*;usage=read;metadata=read;zone=read").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Вввод отображения size
			size, e := pterm.DefaultInteractiveSelect.
				WithOptions([]string{"kb", "gb", "tb"}).
				WithDefaultText(placeholder + " Выбор отображения поля size").
				Show()
			if e != nil {
				logrus.Error("Ошибка: ", e)
				os.Exit(1)
			}

			// Выбор метода печати
			print_methods := []string{"table", "json"}
			print_method, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(print_methods)).
				WithDefaultText(placeholder + " Выбор метода печати").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			e = rgwapi.CreateRgwUser(c, rgwUserParam, print_method, userCups, size, refresh, nocache)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			e = rgwapi.GetUserKeys(c, userUuid, print_method)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			return
		}
		if rgwenpoint_operation == "Создание subuser S3 пользователя" {
			// Вввод uid пользователя
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = "    └─Введите uid пользователя для которого создается subuser (обязательное поле)"
			user_uid, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if user_uid == "" {
				logrus.Error("UUID пользователя не может быть пустым!")
				os.Exit(1)
			}

			// Вввод RgwUserParam пользователя
			rgwUserParam := ""
			subuserIDInput := pterm.DefaultInteractiveTextInput
			subuserIDInput.DefaultText = "    └─Subuser ID (числовой номер по порядку) "
			subuserID, e := subuserIDInput.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if subuserID == "" {
				logrus.Error("ID пользователя не может быть пустым!")
				os.Exit(1)
			}

			// Выбор разрешений пользовтеля
			subuserPerm, e := pterm.DefaultInteractiveSelect.
				WithOptions([]string{"read", "write", "readwrite", "full", "none"}).
				WithDefaultText("    └─Разрешения пользователя").
				Show()
			if e != nil {
				logrus.Error("Ошибка: ", e)
				os.Exit(1)
			}

			rgwUserParam = fmt.Sprintf("%s|%s", subuserID, subuserPerm)

			// Выбор метода печати
			print_methods := []string{"table", "json"}
			print_method, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(print_methods)).
				WithDefaultText(placeholder + " Выбор метода печати").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Вввод отображения size
			size, e := pterm.DefaultInteractiveSelect.
				WithOptions([]string{"kb", "gb", "tb"}).
				WithDefaultText(placeholder + " Выбор отображения поля size").
				Show()
			if e != nil {
				logrus.Error("Ошибка: ", e)
				os.Exit(1)
			}

			e = rgwapi.CreateSubUser(c, user_uid, rgwUserParam, print_method, size, refresh, nocache)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			return
		}
		if rgwenpoint_operation == "Получение списка ключей и caps пользователя" {
			// Вввод uid пользователя
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = "    └─Введите uid пользователя (обязательное поле)"
			user_uid, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if user_uid == "" {
				logrus.Error("Не указан uuid пользователя")
				os.Exit(1)
			}

			// Выбор метода печати
			print_methods := []string{"table", "json"}
			print_method, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(print_methods)).
				WithDefaultText(placeholder + " Выбор метода печати").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			e = rgwapi.GetUserKeys(c, user_uid, print_method)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
		}
		if rgwenpoint_operation == "Установка или изменение пользовательских CUPS" {
			// Вввод uid пользователя
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = "    └─Введите uid пользователя (обязательное поле)"
			user_uid, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			if user_uid == "" {
				logrus.Error("Не указан uuid пользователя")
				os.Exit(1)
			}

			// Вввод UserCups пользователя
			userCupsInput := pterm.DefaultInteractiveTextInput
			userCupsInput.DefaultText = "    └─CUPS пользователя (или оставьте пустым): "
			userCups, e := userCupsInput.
				WithDefaultValue("buckets=*;users=*;usage=read;metadata=read;zone=read").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Выбор метода печати
			print_methods := []string{"table", "json"}
			print_method, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(print_methods)).
				WithDefaultText(placeholder + " Выбор метода печати").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			e = rgwapi.SetUserCaps(c, user_uid, userCups, print_method)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			return
		}
		if rgwenpoint_operation == "Удаление CAPS пользователя" {
			// Вввод uid пользователя
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = "    └─Введите uid пользователя (обязательное поле)"
			user_uid, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			if user_uid == "" {
				logrus.Error("Не указан uuid пользователя")
				os.Exit(1)
			}

			e = rgwapi.RemoveUserCaps(c, user_uid)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
		}
		if rgwenpoint_operation == "Создания нового S3 ключа для пользователя" {
			// Вввод uid пользователя
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = "    └─Введите uid пользователя (обязательное поле)"
			user_uid, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			if user_uid == "" {
				logrus.Error("Не указан uuid пользователя")
				os.Exit(1)
			}

			// Выбор метода печати
			key_types := []string{"s3", "swift"}
			key_type, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(key_types)).
				WithDefaultText(placeholder + " Тип ключа").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Вввод uid пользователя
			input = pterm.DefaultInteractiveTextInput
			input.DefaultText = "    └─Введите subuser id если ключ создается для subuser"
			suser_uid, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Выбор метода печати
			print_methods := []string{"table", "json"}
			print_method, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(print_methods)).
				WithDefaultText(placeholder + " Выбор метода печати").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			e = rgwapi.CreateKey(c, user_uid, key_type, print_method, suser_uid)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
		}
		if rgwenpoint_operation == "Удаление subusers" {
			// Вввод uid пользователя
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = "    └─Введите uid пользователя (обязательное поле)"
			user_uid, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			if user_uid == "" {
				logrus.Error("Не указан uuid пользователя")
				os.Exit(1)
			}

			// Вввод uid пользователя
			input = pterm.DefaultInteractiveTextInput
			input.DefaultText = "    └─Введите subuser id (обязательное поле)"
			suser_uid, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			if suser_uid == "" {
				logrus.Error("Не указан id subuser")
				os.Exit(1)
			}

			e = rgwapi.RemoveSubUser(c, user_uid, suser_uid)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			rgwusers, e := rgwapi.GetUserList(c, refresh, nocache)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Выбор метода печати
			print_methods := []string{"table", "json"}
			print_method, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(print_methods)).
				WithDefaultText(placeholder + " Выбор метода печати").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Печать в виде json
			if print_method == "json" {
				e = rgwapi.UserJPrint(*rgwusers, user_uid)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
			}
			// Печать в табличном виде
			if print_method == "table" {
				// Вввод отображения size
				size, e := pterm.DefaultInteractiveSelect.
					WithOptions([]string{"kb", "gb", "tb"}).
					WithDefaultText(placeholder + " Выбор отображения поля size").
					Show()
				if e != nil {
					logrus.Error("Ошибка: ", e)
					os.Exit(1)
				}
				rgwapi.UserTprint(*rgwusers, user_uid, size)
			}
			return
		}
		if rgwenpoint_operation == "Удаление S3 пользователя" {
			c, e := rgwapi.MakeNewConnection(enpoint_value.Endpoints.Rgw.S3Endpoint, enpoint_value.Endpoints.Rgw.AcessKey, enpoint_value.Endpoints.Rgw.SecretKey)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Вввод uid пользователя
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = "    └─Введите uid пользователя (обязательное поле)"
			user_uid, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			if user_uid == "" {
				logrus.Error("Не указан uuid пользователя")
				os.Exit(1)
			}

			e = rgwapi.DeleteUser(c, user_uid)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
		}
		if rgwenpoint_operation == "Смена placement и storageclass" {
			c, e := rgwapi.MakeNewConnection(enpoint_value.Endpoints.Rgw.S3Endpoint, enpoint_value.Endpoints.Rgw.AcessKey, enpoint_value.Endpoints.Rgw.SecretKey)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Вввод uid пользователя
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = "    └─Введите uid пользователя (обязательное поле)"
			user_uid, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			if user_uid == "" {
				logrus.Error("Не указан uuid пользователя")
				os.Exit(1)
			}

			// Ввод placement
			input_placement := pterm.DefaultInteractiveTextInput
			input_placement.DefaultText = "    └─Введите placement (обязательное поле)"
			placement, e := input_placement.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			if placement == "" {
				logrus.Error("Не указан placement")
				os.Exit(1)
			}

			// Ввод storage class
			input_storageclass := pterm.DefaultInteractiveTextInput
			input_storageclass.DefaultText = "    └─Введите storage class (обязательное поле)"
			storage_class, e := input_storageclass.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			if storage_class == "" {
				logrus.Error("Не указан storage class")
				os.Exit(1)
			}

			// Выбор метода печати
			print_methods := []string{"table", "json"}
			print_method, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(print_methods)).
				WithDefaultText(placeholder + " Выбор метода печати").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			size := ""
			// Печать в табличном виде
			if print_method == "table" {
				// Вввод отображения size
				size, e = pterm.DefaultInteractiveSelect.
					WithOptions([]string{"kb", "gb", "tb"}).
					WithDefaultText(placeholder + " Выбор отображения поля size").
					Show()
				if e != nil {
					logrus.Error("Ошибка: ", e)
					os.Exit(1)
				}
			}

			rgwapi.ChangeUserPlacementAndStorageClass(c, user_uid, "", placement, storage_class, print_method, size, refresh, nocache)
			return
		}

	} else if enpoint_value.Operation == "RADOS" {
		placeholder := fmt.Sprintf("\033[35m%s >\033[0m", enpoint_value.Endpoints.Rados.Name)
		radosenpoint_operation := radosOperationsMenu(placeholder)
		// Создания экземпляра RADOS подключения
		conn, e := radosapi.MakeNewRadosConnection(enpoint_value.Endpoints.Rados.Fsid,
			enpoint_value.Endpoints.Rados.MonHosts, enpoint_value.Endpoints.Rados.KeyRing)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}

		if radosenpoint_operation == "Получить информацию о кластере" {
			// Выбор метода печати
			print_methods := []string{"table", "json"}
			print_method, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(print_methods)).
				WithDefaultText(placeholder + " Выбор метода печати").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			e = radosapi.GetClusterInfo(conn, print_method)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			return
		}
		if radosenpoint_operation == "Получить список всех subvolume указанной CephfFS" {

			// Ввод названия cephfs
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = placeholder + " Введите название cephFS (обязательное поле) "
			cepf_fsname, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if cepf_fsname == "" {
				logrus.Error("CephFS Name - не может быть пустым!")
				os.Exit(1)
			}

			// Выбор метода печати
			print_methods := []string{"table", "json"}
			print_method, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(print_methods)).
				WithDefaultText(placeholder + " Выбор метода печати").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			e = cephfsapi.GetCephfsSubvolumeList(conn, print_method, cepf_fsname)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			return
		}
		if radosenpoint_operation == "Получить список существующих CephFS" {
			// Ввод названия cephfs
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = placeholder + " Введите название cephFS для фильтрации или оставьте поле пустым "
			cepf_fsname, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			// Выбор метода печати
			print_methods := []string{"table", "json"}
			print_method, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(print_methods)).
				WithDefaultText(placeholder + " Выбор метода печати").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			e = cephfsapi.GetCephFsList(conn, print_method, cepf_fsname, nocache, refresh)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			return
		}
		if radosenpoint_operation == "Получение информации о пользователе CEPH" {
			// Ввод имени пользователя cephfs
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = placeholder + " Введите имя пользователя ceph (обязательное поле) "
			username, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if username == "" {
				logrus.Error("Ceph username - не может быть пустым!")
				os.Exit(1)
			}

			e = cephfsapi.GetUserInfo(conn, username)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			return
		}
		if radosenpoint_operation == "Изменение или установка caps пользователя CEPH" {
			// Ввод имени пользователя cephfs
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = placeholder + " Введите имя пользователя ceph (обязательное поле) "
			username, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if username == "" {
				logrus.Error("Ceph username - не может быть пустым!")
				os.Exit(1)
			}

			// Ввод caps
			// mon
			input_caps_mon := pterm.DefaultInteractiveTextInput
			input_caps_mon.DefaultText = "    └─MON CAPS "
			caps_mon, e := input_caps_mon.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if caps_mon == "" {
				logrus.Error("MON CAPS - не может быть пустым!")
				os.Exit(1)
			}
			// mds
			input_caps_mds := pterm.DefaultInteractiveTextInput
			input_caps_mds.DefaultText = "    └─MDS CAPS "
			caps_mds, e := input_caps_mds.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if caps_mds == "" {
				logrus.Error("MDS CAPS - не может быть пустым!")
				os.Exit(1)
			}
			// osd
			input_caps_osd := pterm.DefaultInteractiveTextInput
			input_caps_osd.DefaultText = "    └─OSD CAPS "
			caps_osd, e := input_caps_osd.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if caps_osd == "" {
				logrus.Error("MDS CAPS - не может быть пустым!")
				os.Exit(1)
			}

			caps := fmt.Sprintf("mon==%s;mds==%s;osd==%s", caps_mon, caps_mds, caps_osd)

			if caps == "" {
				logrus.Error("Ошибка - caps не может быть пустым!")
				os.Exit(1)
			}

			e = cephfsapi.SetCephUserCaps(conn, username, caps)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			return
		}
		if radosenpoint_operation == "Создание нового CEPH USER" {
			// Ввод имени пользователя cephfs
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = placeholder + " Введите имя пользователя ceph (обязательное поле) "
			username, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if username == "" {
				logrus.Error("Ceph username - не может быть пустым!")
				os.Exit(1)
			}

			// Выбор назначения caps
			select_caps_setup := []string{"yes", "no"}
			caps_setup, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(select_caps_setup)).
				WithDefaultText(placeholder + " Задать CAPS для создаваемого пользователя?").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			caps := ""
			if caps_setup == "yes" {
				fmt.Println(`
	--- Пример CAPS для пользователя ---

	Минимальные права:
		* mon 'allow r'
	RBD с правами на запись:
		* mon 'allow r'
		* osd 'allow rwx pool=<rbd_pool_name'
	Для доступа к CephFS:
		- Общие права (доступ на чтения ко всем CephFS)
			* mon 'allow r'
			* osd 'allow rw pool=cephfs_data'
			* mds 'allow rw'
		- Точечные права (доступ только к определенной CephFS)
			* mon 'allow r'
			* mds 'allow rw fsname=<fs_name>'
			* osd 'allow rwx tag cephfs data=<fs_name>'
	`)
				// Ввод caps
				// mon
				input_caps_mon := pterm.DefaultInteractiveTextInput
				input_caps_mon.DefaultText = "    └─MON CAPS "
				caps_mon, e := input_caps_mon.
					WithDefaultValue("").
					Show()

				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}

				if caps_mon == "" {
					logrus.Error("MON CAPS - не может быть пустым!")
					os.Exit(1)
				}
				// mds
				input_caps_mds := pterm.DefaultInteractiveTextInput
				input_caps_mds.DefaultText = "    └─MDS CAPS "
				caps_mds, e := input_caps_mds.
					WithDefaultValue("").
					Show()

				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}

				if caps_mds == "" {
					logrus.Error("MDS CAPS - не может быть пустым!")
					os.Exit(1)
				}
				// osd
				input_caps_osd := pterm.DefaultInteractiveTextInput
				input_caps_osd.DefaultText = "    └─OSD CAPS "
				caps_osd, e := input_caps_osd.
					WithDefaultValue("").
					Show()

				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}

				if caps_osd == "" {
					logrus.Error("MDS CAPS - не может быть пустым!")
					os.Exit(1)
				}

				caps = fmt.Sprintf("mon==%s;mds==%s;osd==%s", caps_mon, caps_mds, caps_osd)
			}

			e = cephfsapi.CreateNewCephUser(conn, username, caps)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			return
		}
		if radosenpoint_operation == "Смена атрибутов существующего pool" {
			// Ввод имени pool
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = "   └─Название pool (обязательное поле) "
			pool_name, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if pool_name == "" {
				logrus.Error(placeholder + " Название pool - не может быть пустым!")
				os.Exit(1)
			}

			// Выбор типа pool
			select_pool_type := []string{"data pool", "meta pool"}
			pool_type, e := pterm.DefaultInteractiveSelect.
				WithOptions(slices.Compact(select_pool_type)).
				WithDefaultText("Укажите тип pool").
				Show()
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			if pool_type == "meta pool" {
				// атрибуты pool "pg_autoscale|pg_num|compresion|quota_bytes|quota_objects"

				e := cephfsapi.ChagePoolAttributes(pool_name, poolAttributeGenerate(pool_type), conn)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
			}
			if pool_type == "data pool" {
				e := cephfsapi.ChagePoolAttributes(pool_name, poolAttributeGenerate(pool_type), conn)
				if e != nil {
					logrus.Error(e)
					os.Exit(1)
				}
			}
			return
		}
		if radosenpoint_operation == "Смена crush_rule для существующего pool" {
			// Ввод имени pool
			input := pterm.DefaultInteractiveTextInput
			input.DefaultText = "   └─Название pool (обязательное поле) "
			pool_name, e := input.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if pool_name == "" {
				logrus.Error("Название pool - не может быть пустым!")
				os.Exit(1)
			}

			// Ввод имени crush_rule
			input_crush_rule := pterm.DefaultInteractiveTextInput
			input_crush_rule.DefaultText = "   └─Название crush_rule (обязательное поле) "
			crush_rule_name, e := input_crush_rule.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if crush_rule_name == "" {
				logrus.Error("Название crush_rule - не может быть пустым!")
				os.Exit(1)
			}

			e = cephfsapi.ChangePoolCrushRule(pool_name, crush_rule_name, conn)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			return
		}
		if radosenpoint_operation == "Создание новой cephfs" {

			// Ввод имени cephfs
			input_cephfs_name := pterm.DefaultInteractiveTextInput
			input_cephfs_name.DefaultText = "   └─Название cephfs (обязательное поле) "
			cephfs_name, e := input_cephfs_name.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if cephfs_name == "" {
				logrus.Error("Название pool - не может быть пустым!")
				os.Exit(1)
			}

			// Ввод имени cephfs
			input_placement := pterm.DefaultInteractiveTextInput
			input_placement.DefaultText = "   └─Placement (обязательное поле) "
			placement, e := input_placement.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if placement == "" {
				logrus.Error("Placement - не может быть пустым!")
				os.Exit(1)
			}

			// Ввод названия DefaultMetaCrushRule
			input_DefaultMetaCrushRule := pterm.DefaultInteractiveTextInput
			input_DefaultMetaCrushRule.DefaultText = "   └─Crush rule для meta pool (если пусто будет взят из конфигурации endpoint) "
			defaultMetaCrushRule, e := input_DefaultMetaCrushRule.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if defaultMetaCrushRule == "" {
				defaultMetaCrushRule = enpoint_value.Endpoints.Rados.DefaultMetaCrushRule
			}

			// Ввод названия DefaultDataCrushRule
			input_DefaultDataCrushRule := pterm.DefaultInteractiveTextInput
			input_DefaultDataCrushRule.DefaultText = "   └─Crush rule для data pool (если пусто будет взят из конфигурации endpoint) "
			defaultDetaCrushRule, e := input_DefaultDataCrushRule.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if defaultDetaCrushRule == "" {
				defaultDetaCrushRule = enpoint_value.Endpoints.Rados.DefaultDataCrushRule
			}

			e = cephfsapi.CreateNewCephFS(conn, cephfs_name, placement,
				defaultMetaCrushRule, defaultDetaCrushRule,
				poolAttributeGenerate("meta"), poolAttributeGenerate("data"))
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			return
		}
		if radosenpoint_operation == "Удаление пользователя ceph" {
			// Ввод имени cephfs
			input_cephfs_username := pterm.DefaultInteractiveTextInput
			input_cephfs_username.DefaultText = "   └─ID пользователя (обязательное поле) "
			cephfs_username, e := input_cephfs_username.
				WithDefaultValue("").
				Show()

			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}

			if cephfs_username == "" {
				logrus.Error("ID пользователя - не может быть пустым!")
				os.Exit(1)
			}

			e = cephfsapi.DeleteCephUser(conn, cephfs_username)
			if e != nil {
				logrus.Error(e)
				os.Exit(1)
			}
			return
		}
	}
}

// Метод отображения операция для RGW
func rgwOperationsMenu(placeholder string) string {
	options := []string{
		"Получение списка пользователей RGW или отдельного пользователя по его UID",
		"Получение списка бакетов",
		"Создание нового пользователя S3",
		"Создание subuser S3 пользователя",
		"Получение списка ключей и caps пользователя",
		"Установка или изменение пользовательских CUPS",
		"Удаление CAPS пользователя",
		"Создания нового S3 ключа для пользователя",
		"Удаление subusers",
		"Удаление S3 пользователя",
		"Смена placement и storageclass",
	}
	selectedOption, e := pterm.DefaultInteractiveSelect.
		WithOptions(slices.Compact(options)).
		WithDefaultText(placeholder + " Выбор операции RGW").
		Show()
	if e != nil {
		logrus.Error(e)
		os.Exit(1)
	}
	return selectedOption
}

// Метод отображения операция для RADOS
func radosOperationsMenu(placeholder string) string {
	options := []string{"Создание новой cephfs", "Смена crush_rule для существующего pool",
		"Смена атрибутов существующего pool", "Создание нового CEPH USER",
		"Изменение или установка caps пользователя CEPH", "Получение информации о пользователе CEPH",
		"Получить список существующих CephFS", "Получить список всех subvolume указанной CephfFS",
		"Получить информацию о кластере", "Удаление пользователя ceph",
	}
	selectedOption, e := pterm.DefaultInteractiveSelect.
		WithOptions(slices.Compact(options)).
		WithDefaultText(placeholder + " Выбор операции RADOS").
		Show()
	if e != nil {
		logrus.Error(e)
		os.Exit(1)
	}
	return selectedOption
}

func printMinHelp() {
	fmt.Println(`
--- cephapi - инструмент имплиментации над api CEPH ---

Типы операций:
	* RGW	- операции API c API RGW
	* RADOS	- операции API c API RADOS
Как пользоваться:
	* Выбери тип подключения
	* Выбери операцию и следуйте посказкам

--- Рабработчик Боканёв Артём ---
	`)
}

// Метод выбора endpoint для подключения
func selectEndpointMenu() struct {
	Operation string
	Enpoint   string
	Endpoints ConnectionEndpoints
} {

	// Все доступные endpoints
	endpoints, e := Endpoints()
	if e != nil {
		logrus.Error("Ошибка считывания конфигурации endpoints: ", e)
		os.Exit(1)
	}

	// Список enpoints names для вывода в меню выбора
	endpoints_names := func(endpoints struct {
		Rgw   rgwapi.Connection
		Rados cephfsapi.Connection
	}) []string {
		buf := []string{}
		for _, erados := range endpoints.Rados.RADOS.EndPoints {
			buf = append(buf, "RADOS: "+erados.Name)
		}
		for _, ergw := range endpoints.Rgw.RGW.EndPoints {
			buf = append(buf, "RGW:   "+ergw.Name)
		}
		return buf
	}(endpoints)
	if len(endpoints_names) == 0 {
		logrus.Error("Endpoints not defined")
		os.Exit(1)
	}

	// Вывод менюб выбора endpoint
	selectedEndpoint, e := pterm.DefaultInteractiveSelect.
		WithOptions(slices.Compact(endpoints_names)).
		WithDefaultText("Выбор endpoint подключения").
		Show()
	if e != nil {
		logrus.Error("Ошибка считывания конфигурации endpoints: ", e)
		os.Exit(1)
	}

	// Шаблон возвращаемых данных
	value := struct {
		Operation string
		Enpoint   string
		Endpoints ConnectionEndpoints
	}{}

	if len(strings.Split(selectedEndpoint, ":")) == 2 {
		value.Operation = strings.TrimSpace(strings.Split(selectedEndpoint, ":")[0])
		value.Enpoint = strings.TrimSpace(strings.Split(selectedEndpoint, ":")[1])
		if value.Operation == "RGW" {
			for _, endpoint := range endpoints.Rgw.RGW.EndPoints {
				if endpoint.Name == value.Enpoint {
					value.Endpoints.Rgw = endpoint
				}
			}
		} else if value.Operation == "RADOS" {
			for _, endpoint := range endpoints.Rados.RADOS.EndPoints {
				if endpoint.Name == value.Enpoint {
					value.Endpoints.Rados = endpoint
				}
			}
		}

	} else {
		logrus.Error("Error parse endpoint value!")
		os.Exit(1)
	}

	return value
}

// Метод задания парампетров атрибутов пула
func poolAttributeGenerate(t string) string {
	pterm.Printf("   +─Задание атрибутов для %s pool\n", t)
	// Выбор режима автосколирования pool
	pg_autoscale_select := []string{"off", "on"}
	pg_autoscale, e := pterm.DefaultInteractiveSelect.
		WithOptions(slices.Compact(pg_autoscale_select)).
		WithDefaultText("   └─Использовать pg_autoscale?").
		Show()
	if e != nil {
		logrus.Error(e)
		os.Exit(1)
	}
	// Ввод pg_num
	input_pg_num := pterm.DefaultInteractiveTextInput
	input_pg_num.DefaultText = "   └─Кол-во PG на pool "
	pg_num, e := input_pg_num.
		WithDefaultValue("32").
		Show()

	if e != nil {
		logrus.Error(e)
		os.Exit(1)
	}

	if pg_num == "" {
		logrus.Error("Кол-во PG - не может быть пустым!")
		os.Exit(1)
	}
	// Ввод quota_size
	input_quota_size := pterm.DefaultInteractiveTextInput
	input_quota_size.DefaultText = "   └─Квота в kb "
	quota_size, e := input_quota_size.
		WithDefaultValue("0").
		Show()

	if e != nil {
		logrus.Error(e)
		os.Exit(1)
	}
	// Ввод quota_objects
	input_quota_objects := pterm.DefaultInteractiveTextInput
	input_quota_objects.DefaultText = "   └─Квота в kb "
	quota_objects, e := input_quota_objects.
		WithDefaultValue("0").
		Show()

	if e != nil {
		logrus.Error(e)
		os.Exit(1)
	}
	// Выбор режима сжатия
	select_compression := []string{"none", "zlib", "lz4", "snappy", "zstd"}
	compression, e := pterm.DefaultInteractiveSelect.
		WithOptions(slices.Compact(select_compression)).
		WithDefaultText("   └─Выбор метода сжатия данных").
		Show()
	if e != nil {
		logrus.Error(e)
		os.Exit(1)
	}

	return fmt.Sprintf("%s|%s|%s|%s|%s", pg_autoscale, pg_num, compression, quota_size, quota_objects)
}

// Метод меню внесения изменений в endpoints
func ManageMenu() {
	endpoints, e := Endpoints()
	if e != nil {
		logrus.Error(e)
		os.Exit(1)
	}
	// Выбор операции
	operations := []string{
		"Добавление entrypoint",
		"Редактирование entrypoint",
		"Удаление entrypoint"}
	operation, e := pterm.DefaultInteractiveSelect.
		WithOptions(slices.Compact(operations)).
		WithDefaultText("Выбор операции").
		Show()
	if e != nil {
		logrus.Error(e)
		os.Exit(1)
	}

	switch operation {
	case "Добавление entrypoint":
		addEntryPoint(endpoints)
		return
	case "Редактирование entrypoint":
		editEntryPoint(endpoints)
		return
	case "Удаление entrypoint":
		deleteEntryPoint(endpoints)
		return
	}
}

// Метод добавления нового endpoint
func addEntryPoint(endpoints struct {
	Rgw   rgwapi.Connection
	Rados cephfsapi.Connection
}) {
	// Выбор типа добавляемого endpoint
	entrytypes := []string{
		"RGW",
		"RADOS"}
	entrytype, e := pterm.DefaultInteractiveSelect.
		WithOptions(slices.Compact(entrytypes)).
		WithDefaultText("Тип entry").
		Show()
	if e != nil {
		logrus.Error(e)
		os.Exit(1)
	}

	switch entrytype {
	case "RGW":
		// Название подключения
		inputentryname := pterm.DefaultInteractiveTextInput
		inputentryname.DefaultText = "    └─Введите название подключения"
		entryname, e := inputentryname.
			WithDefaultValue("").
			Show()

		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		if entryname == "" {
			logrus.Error("entryname - не может быть пустым!")
			os.Exit(1)
		}

		// URL подключения
		inputS3Endpoint := pterm.DefaultInteractiveTextInput
		inputS3Endpoint.DefaultText = "    └─Введите url подключения"
		s3Endpoint, e := inputS3Endpoint.
			WithDefaultValue("").
			Show()

		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		if s3Endpoint == "" {
			logrus.Error("s3Endpoint - не может быть пустым!")
			os.Exit(1)
		}
		_, e = url.ParseRequestURI(s3Endpoint)
		if e != nil {
			logrus.Error("s3Endpoint - не является корректным URL")
			os.Exit(1)
		}

		// AcessKey подключения
		inputAcessKey := pterm.DefaultInteractiveTextInput
		inputAcessKey.DefaultText = "    └─Введите AcessKey"
		acessKey, e := inputAcessKey.
			WithDefaultValue("").
			Show()

		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		if acessKey == "" {
			logrus.Error("acessKey - не может быть пустым!")
			os.Exit(1)
		}

		// SecretKey подключения
		inputSecretKey := pterm.DefaultInteractiveTextInput
		inputSecretKey.DefaultText = "    └─Введите SecretKey"
		secretKey, e := inputSecretKey.
			WithDefaultValue("").
			Show()

		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		if secretKey == "" {
			logrus.Error("secretKey - не может быть пустым!")
			os.Exit(1)
		}

		endpoints.Rgw.RGW.EndPoints = append(endpoints.Rgw.RGW.EndPoints, struct {
			Name       string `yaml:"Name"`
			S3Endpoint string `yaml:"S3Endpoint"`
			AcessKey   string `yaml:"AcessKey"`
			SecretKey  string `yaml:"SecretKey"`
		}{
			Name:       entryname,
			S3Endpoint: s3Endpoint,
			AcessKey:   acessKey,
			SecretKey:  secretKey,
		})

		// Структура новой конфигурации
		сonfig := struct {
			RGW   rgwapi.RGW      `yaml:"RGW"`
			RADOS cephfsapi.RADOS `yaml:"RADOS"`
		}{
			RGW:   endpoints.Rgw.RGW,
			RADOS: endpoints.Rados.RADOS,
		}
		e = saveEditedEntryPoints(сonfig)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		logrus.Info("RGW entrypoint успешно добавлен")

	case "RADOS":
		// Название подключения
		inputentryname := pterm.DefaultInteractiveTextInput
		inputentryname.DefaultText = "    └─Введите название подключения"
		entryname, e := inputentryname.
			WithDefaultValue("").
			Show()

		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		if entryname == "" {
			logrus.Error("entryname - не может быть пустым!")
			os.Exit(1)
		}

		// FSID кластера
		inputfsid := pterm.DefaultInteractiveTextInput
		inputfsid.DefaultText = "    └─Введите FSID кластера"
		fsid, e := inputfsid.
			WithDefaultValue("").
			Show()

		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		if fsid == "" {
			logrus.Error("fsid - не может быть пустым!")
			os.Exit(1)
		}

		// MON
		inputmon := pterm.DefaultInteractiveTextInput
		inputmon.DefaultText = "    └─Введите MON HOSTS кластера"
		monHosts, e := inputmon.
			WithDefaultValue("").
			Show()

		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		if monHosts == "" {
			logrus.Error("MonHosts - не может быть пустым!")
			os.Exit(1)
		}

		// DefaultMetaCrushRule
		inputdefaultMetaCrushRule := pterm.DefaultInteractiveTextInput
		inputdefaultMetaCrushRule.DefaultText = "    └─Введите DefaultMetaCrushRule кластера"
		defaultMetaCrushRule, e := inputdefaultMetaCrushRule.
			WithDefaultValue("").
			Show()

		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		if defaultMetaCrushRule == "" {
			logrus.Error("DefaultMetaCrushRule - не может быть пустым!")
			os.Exit(1)
		}

		// DefaultDataCrushRule
		inputdefaultDataCrushRule := pterm.DefaultInteractiveTextInput
		inputdefaultDataCrushRule.DefaultText = "    └─Введите DefaultDataCrushRule кластера"
		defaultDataCrushRule, e := inputdefaultDataCrushRule.
			WithDefaultValue("").
			Show()

		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		if defaultDataCrushRule == "" {
			logrus.Error("DefaultDataCrushRule - не может быть пустым!")
			os.Exit(1)
		}

		// DefaultDataCrushRule
		inputkeyRing := pterm.DefaultInteractiveTextInput
		inputkeyRing.DefaultText = "    └─Введите KeyRing административной УЗ"
		keyRing, e := inputkeyRing.
			WithDefaultValue("").
			Show()

		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		if keyRing == "" {
			logrus.Error("KeyRing - не может быть пустым!")
			os.Exit(1)
		}

		endpoints.Rados.RADOS.EndPoints = append(endpoints.Rados.RADOS.EndPoints, struct {
			Name                 string `yaml:"Name"`
			Fsid                 string `yaml:"Fsid"`
			MonHosts             string `yaml:"MonHosts"`
			DefaultMetaCrushRule string `yaml:"DefaultMetaCrushRule"`
			DefaultDataCrushRule string `yaml:"DefaultDataCrushRule"`
			KeyRing              string `yaml:"KeyRing"`
		}{
			Name:                 entryname,
			Fsid:                 fsid,
			MonHosts:             monHosts,
			DefaultMetaCrushRule: defaultMetaCrushRule,
			DefaultDataCrushRule: defaultDataCrushRule,
			KeyRing:              keyRing,
		})

		// Структура новой конфигурации
		сonfig := struct {
			RGW   rgwapi.RGW      `yaml:"RGW"`
			RADOS cephfsapi.RADOS `yaml:"RADOS"`
		}{
			RGW:   endpoints.Rgw.RGW,
			RADOS: endpoints.Rados.RADOS,
		}
		e = saveEditedEntryPoints(сonfig)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		logrus.Info("RADOS entrypoint успешно добавлен")
	}
}

// Метод удаления entrypoint
func deleteEntryPoint(endpoints struct {
	Rgw   rgwapi.Connection
	Rados cephfsapi.Connection
}) {
	// Выбор типа удалчемого endpoint
	entrytypes := []string{
		"RGW",
		"RADOS"}
	entrytype, e := pterm.DefaultInteractiveSelect.
		WithOptions(slices.Compact(entrytypes)).
		WithDefaultText("Тип entry").
		Show()
	if e != nil {
		logrus.Error(e)
		os.Exit(1)
	}
	switch entrytype {
	case "RGW":
		// выбор подключения из доступных
		entrynames, e := func(endpoints struct {
			Rgw   rgwapi.Connection
			Rados cephfsapi.Connection
		}) ([]string, error) {
			en := []string{}
			if len(endpoints.Rgw.RGW.EndPoints) == 0 {
				return en, errors.New("нет подключений для редактирования")
			}
			for _, e := range endpoints.Rgw.RGW.EndPoints {
				en = append(en, e.Name)
			}
			return en, nil
		}(endpoints)
		if e != nil {
			logrus.Warn(e)
			os.Exit(1)
		}

		entryname, e := pterm.DefaultInteractiveSelect.
			WithOptions(slices.Compact(entrynames)).
			WithDefaultText("Выбор RGW подключения").
			Show()
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}

		if entryname == "" {
			logrus.Error("entryname - не может быть пустым!")
			os.Exit(1)
		}

		if endpoints.Rgw.RGW.EndPoints == nil {
			logrus.Error("Ошибка исполнения. Список подключений не задан")
			os.Exit(1)
		}

		if len(endpoints.Rgw.RGW.EndPoints) == 0 {
			logrus.Warn("Список подключений пуст. Операпция отменена.")
			return
		}
		// Удаляем элемент из массива
		for i, endpoint := range endpoints.Rgw.RGW.EndPoints {
			if endpoint.Name == entryname {
				if len(endpoints.Rgw.RGW.EndPoints) < 2 {
					endpoints.Rgw.RGW = rgwapi.RGW{}
				} else {
					endpoints.Rgw.RGW.EndPoints = slices.Delete(endpoints.Rgw.RGW.EndPoints, i, i+1)
				}
				logrus.Info("Удалено RGW подключение: ", entryname)
				break
			}
			logrus.Warn("Entrypoint не найден")
		}

		// Структура новой конфигурации
		сonfig := struct {
			RGW   rgwapi.RGW      `yaml:"RGW"`
			RADOS cephfsapi.RADOS `yaml:"RADOS"`
		}{
			RGW:   endpoints.Rgw.RGW,
			RADOS: endpoints.Rados.RADOS,
		}
		e = saveEditedEntryPoints(сonfig)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		logrus.Info("Операция выполненна")

	case "RADOS":
		// выбор подключения из доступных
		entrynames, e := func(endpoints struct {
			Rgw   rgwapi.Connection
			Rados cephfsapi.Connection
		}) ([]string, error) {
			en := []string{}
			if len(endpoints.Rados.RADOS.EndPoints) == 0 {
				return en, errors.New("нет подключений для редактирования")
			}
			for _, e := range endpoints.Rados.RADOS.EndPoints {
				en = append(en, e.Name)
			}
			return en, nil
		}(endpoints)
		if e != nil {
			logrus.Info(e)
			os.Exit(1)
		}

		entryname, e := pterm.DefaultInteractiveSelect.
			WithOptions(slices.Compact(entrynames)).
			WithDefaultText("Выбор RADOS подключения").
			Show()
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}

		if entryname == "" {
			logrus.Error("entryname - не может быть пустым!")
			os.Exit(1)
		}

		if endpoints.Rados.RADOS.EndPoints == nil {
			logrus.Error("Ошибка исполнения. Список подключений не задан")
			os.Exit(1)
		}

		if len(endpoints.Rados.RADOS.EndPoints) == 0 {
			logrus.Warn("Список подключений пуст. Операпция отменена.")
			return
		}
		// Удаляем элемент из массива
		for i, endpoint := range endpoints.Rados.RADOS.EndPoints {
			if endpoint.Name == entryname {
				if len(endpoints.Rados.RADOS.EndPoints) < 2 {
					endpoints.Rados.RADOS = cephfsapi.RADOS{}
				} else {
					endpoints.Rados.RADOS.EndPoints = slices.Delete(endpoints.Rados.RADOS.EndPoints, i, i+1)
				}
				logrus.Info("Удалено RADOS подключение: ", entryname)
				break
			}
			logrus.Warn("Entrypoint не найден")
		}

		// Структура новой конфигурации
		сonfig := struct {
			RGW   rgwapi.RGW      `yaml:"RGW"`
			RADOS cephfsapi.RADOS `yaml:"RADOS"`
		}{
			RGW:   endpoints.Rgw.RGW,
			RADOS: endpoints.Rados.RADOS,
		}
		e = saveEditedEntryPoints(сonfig)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		logrus.Info("Операция выполненна")
	}
}

// Метод редактирования Entrypoint
func editEntryPoint(endpoints struct {
	Rgw   rgwapi.Connection
	Rados cephfsapi.Connection
}) {
	// Выбор типа добавляемого endpoint
	entrytypes := []string{
		"RGW",
		"RADOS"}
	entrytype, e := pterm.DefaultInteractiveSelect.
		WithOptions(slices.Compact(entrytypes)).
		WithDefaultText("Тип entry").
		Show()
	if e != nil {
		logrus.Error(e)
		os.Exit(1)
	}
	switch entrytype {
	case "RGW":
		// выбор подключения из доступных
		entrynames, e := func(endpoints struct {
			Rgw   rgwapi.Connection
			Rados cephfsapi.Connection
		}) ([]string, error) {
			en := []string{}
			if len(endpoints.Rgw.RGW.EndPoints) == 0 {
				return en, errors.New("нет подключений для редактирования")
			}
			for _, e := range endpoints.Rgw.RGW.EndPoints {
				en = append(en, e.Name)
			}
			return en, nil
		}(endpoints)
		if e != nil {
			logrus.Warn(e)
			os.Exit(1)
		}

		entryname, e := pterm.DefaultInteractiveSelect.
			WithOptions(slices.Compact(entrynames)).
			WithDefaultText("Выбор RGW подключения").
			Show()
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}

		if entryname == "" {
			logrus.Error("entryname - не может быть пустым!")
			os.Exit(1)
		}

		// Редактирование выбранного подключения
		// URL подключения
		inputS3Endpoint := pterm.DefaultInteractiveTextInput
		inputS3Endpoint.DefaultText = "    └─Введите url подключения"
		s3Endpoint, e := inputS3Endpoint.
			WithDefaultValue("").
			Show()

		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		if s3Endpoint == "" {
			logrus.Error("s3Endpoint - не может быть пустым!")
			os.Exit(1)
		}
		_, e = url.ParseRequestURI(s3Endpoint)
		if e != nil {
			logrus.Error("s3Endpoint - не является корректным URL")
			os.Exit(1)
		}

		// AcessKey подключения
		inputAcessKey := pterm.DefaultInteractiveTextInput
		inputAcessKey.DefaultText = "    └─Введите AcessKey"
		acessKey, e := inputAcessKey.
			WithDefaultValue("").
			Show()

		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		if acessKey == "" {
			logrus.Error("acessKey - не может быть пустым!")
			os.Exit(1)
		}

		// SecretKey подключения
		inputSecretKey := pterm.DefaultInteractiveTextInput
		inputSecretKey.DefaultText = "    └─Введите SecretKey"
		secretKey, e := inputSecretKey.
			WithDefaultValue("").
			Show()

		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		if secretKey == "" {
			logrus.Error("secretKey - не может быть пустым!")
			os.Exit(1)
		}

		// Редактируем сузествующее подключение задавая в нем новые атрибуты
		for i, entry := range endpoints.Rgw.RGW.EndPoints {
			if entry.Name == entryname {
				entry.S3Endpoint = s3Endpoint
				entry.AcessKey = acessKey
				entry.SecretKey = secretKey
				endpoints.Rgw.RGW.EndPoints[i] = entry // Выполняем замену структуры в массиве
				break
			}
		}

		// Структура новой конфигурации
		сonfig := struct {
			RGW   rgwapi.RGW      `yaml:"RGW"`
			RADOS cephfsapi.RADOS `yaml:"RADOS"`
		}{
			RGW:   endpoints.Rgw.RGW,
			RADOS: endpoints.Rados.RADOS,
		}
		e = saveEditedEntryPoints(сonfig)
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}
		logrus.Info("Операция выполненна")
		// Недоделал редактирвоание RADOS entrypoint
	case "RADOS":
		// выбор подключения из доступных
		entrynames, e := func(endpoints struct {
			Rgw   rgwapi.Connection
			Rados cephfsapi.Connection
		}) ([]string, error) {
			en := []string{}
			if len(endpoints.Rados.RADOS.EndPoints) == 0 {
				return en, errors.New("нет подключений для редактирования")
			}
			for _, e := range endpoints.Rados.RADOS.EndPoints {
				en = append(en, e.Name)
			}
			return en, nil
		}(endpoints)
		if e != nil {
			logrus.Info(e)
			os.Exit(1)
		}

		entryname, e := pterm.DefaultInteractiveSelect.
			WithOptions(slices.Compact(entrynames)).
			WithDefaultText("Выбор RADOS подключения").
			Show()
		if e != nil {
			logrus.Error(e)
			os.Exit(1)
		}

		if entryname == "" {
			logrus.Error("entryname - не может быть пустым!")
			os.Exit(1)
		}
	}
}

// Метод сохранения обновленной конфигурации entrypoints
func saveEditedEntryPoints(сonfig struct {
	RGW   rgwapi.RGW      `yaml:"RGW"`
	RADOS cephfsapi.RADOS `yaml:"RADOS"`
}) error {
	b, e := yaml.Marshal(сonfig)
	if e != nil {
		return e
	}
	eb, e := EncodeFile(b)
	if e != nil {
		return e
	}
	file, e := FindEndpointsConfig()
	if e != nil {
		return e
	}
	e = os.WriteFile(file, eb, 0755)
	if e != nil {
		return e
	}
	return nil
}

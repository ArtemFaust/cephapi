package rgwapi

import (
	"cephapi/db"
	"cephapi/uutils"
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/ceph/go-ceph/rgw/admin"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/rodaine/table"
	"github.com/sirupsen/logrus"
)

// Метод получения списка пользователей RGW
func GetUserList(co *admin.API, refresh bool, nocache bool) (*[]RgwUser, error) {
	rgwusers := []RgwUser{}
	deadline := time.Now().Add(480 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	var users *[]string

	key := uuid.NewSHA1(uuid.NameSpaceX500, []byte(co.Endpoint+co.AccessKey+co.SecretKey+"GetUserList"))
	value, updated_at, e := db.GetResponse(key.String())
	if e != nil || refresh {
		// Получаем списко пользователей
		users, e = co.GetUsers(ctx)
		if e != nil {
			logrus.Error(e)
			return &rgwusers, e
		}
		if !nocache {
			value_b, e := json.Marshal(users)
			if e != nil {
				logrus.Warn("Failed marshal users data: ", e)
			} else {
				db.SaveResponse(key.String(), value_b)
			}
		}
	} else {
		if time.Now().After(updated_at.Add(2 * time.Minute)) {
			// Получаем списко пользователей
			users, e = co.GetUsers(ctx)
			if e != nil {
				logrus.Error(e)
				return &rgwusers, e
			}
			if !nocache {
				value_b, e := json.Marshal(users)
				if e != nil {
					logrus.Warn("Failed marshal users data: ", e)
				} else {
					db.SaveResponse(key.String(), value_b)
				}
			}

		} else {
			e := json.Unmarshal(value.([]byte), &users)
			if e != nil {
				// Получаем списко пользователей
				users, e = co.GetUsers(ctx)
				if e != nil {
					logrus.Error(e)
					return &rgwusers, e
				}
				if !nocache {
					value_b, e := json.Marshal(users)
					if e != nil {
						logrus.Warn("Failed marshal users data: ", e)
					} else {
						db.SaveResponse(key.String(), value_b)
					}
				}

			}
		}
	}

	chunked_users := uutils.ChunkBy(users, runtime.NumCPU())

	// Потокобезапасные горутины
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, users_array := range chunked_users {
		wg.Add(1)
		go func(users *[]string, wg *sync.WaitGroup, mu *sync.Mutex) {
			// Получаем информацию о пользователях
			for _, user := range *users {
				u := RgwUser{}

				// Получаем подробную информацию о пользователе
				var ui admin.User

				key := uuid.NewSHA1(uuid.NameSpaceX500, []byte(co.Endpoint+co.AccessKey+co.SecretKey+user+"GetUserInfo"))
				value, updated_at, e := db.GetResponse(key.String())
				if e != nil || refresh {
					// Получние через API
					ui, e = GetUserInfo(co, ctx, user)
					if e != nil {
						logrus.Error(e)
						continue
					}
					if !nocache {
						// Генерируем уникальный ключ
						key = uuid.NewSHA1(uuid.NameSpaceX500, []byte(co.Endpoint+co.AccessKey+co.SecretKey+user+"GetUserInfo"))
						value_b, e := json.Marshal(ui)
						if e != nil {
							logrus.Warn("Failed marshal users data: ", e)
						} else {
							db.SaveResponse(key.String(), value_b)
						}
					}

				} else {
					// Првоеряем время хранениея данных БД
					// Позволительно 2 минуты
					if time.Now().After(updated_at.Add(2 * time.Minute)) {
						ui, e = GetUserInfo(co, ctx, user)
						if e != nil {
							logrus.Error(e)
							continue
						}
						if !nocache {
							// Генерируем уникальный ключ
							key = uuid.NewSHA1(uuid.NameSpaceX500, []byte(co.Endpoint+co.AccessKey+co.SecretKey+user+"GetUserInfo"))
							value_b, e := json.Marshal(ui)
							if e != nil {
								logrus.Warn("Failed marshal users data: ", e)
							} else {
								db.SaveResponse(key.String(), value_b)
							}
						}

					} else {
						e := json.Unmarshal(value.([]byte), &ui)
						if e != nil {
							logrus.Error(e)
							// Получние через API
							ui, e = GetUserInfo(co, ctx, user)
							if !nocache {
								// Генерируем уникальный ключ
								key = uuid.NewSHA1(uuid.NameSpaceX500, []byte(co.Endpoint+co.AccessKey+co.SecretKey+user+"GetUserInfo"))
								value_b, e := json.Marshal(ui)
								if e != nil {
									logrus.Warn("Failed marshal users data: ", e)
								} else {
									db.SaveResponse(key.String(), value_b)
								}
							}

						}
					}
				}

				u.UserInfo = struct {
					Error error
					Info  admin.User
				}{
					Error: e,
					Info:  ui,
				}

				// Получаем квоты пользователя
				var q admin.QuotaSpec

				key = uuid.NewSHA1(uuid.NameSpaceX500, []byte(co.Endpoint+co.AccessKey+co.SecretKey+user+"GetUserQuota"))
				value, updated_at, e = db.GetResponse(key.String())
				if e != nil || refresh {
					q, e = GetUserQuota(co, ctx, user)
					u.UserQuota = struct {
						Error error
						admin.QuotaSpec
					}{
						QuotaSpec: q,
						Error:     e,
					}
					if !nocache {
						value_b, e := json.Marshal(q)
						if e != nil {
							logrus.Warn("Failed marshal users data: ", e)
						} else {
							db.SaveResponse(key.String(), value_b)
						}
					}

				} else {
					// Првоеряем время хранениея данных БД
					// Позволительно 2 минуты
					if time.Now().After(updated_at.Add(2 * time.Minute)) {
						q, e = GetUserQuota(co, ctx, user)
						u.UserQuota = struct {
							Error error
							admin.QuotaSpec
						}{
							QuotaSpec: q,
							Error:     e,
						}
						if !nocache {
							value_b, e := json.Marshal(q)
							if e != nil {
								logrus.Warn("Failed marshal users data: ", e)
							} else {
								db.SaveResponse(key.String(), value_b)
							}
						}

					} else {
						e := json.Unmarshal(value.([]byte), &q)
						if e != nil {
							logrus.Error(e)
							q, e = GetUserQuota(co, ctx, user)
							u.UserQuota = struct {
								Error error
								admin.QuotaSpec
							}{
								QuotaSpec: q,
								Error:     e,
							}
							if !nocache {
								value_b, e := json.Marshal(q)
								if e != nil {
									logrus.Warn("Failed marshal users data: ", e)
								} else {
									db.SaveResponse(key.String(), value_b)
								}
							}

						} else {
							u.UserQuota = struct {
								Error error
								admin.QuotaSpec
							}{
								QuotaSpec: q,
								Error:     e,
							}
						}
					}
				}

				// Получаем бакеты пользователя
				var b []string

				key = uuid.NewSHA1(uuid.NameSpaceX500, []byte(co.Endpoint+co.AccessKey+co.SecretKey+user+"GetUserBuckets"))
				value, updated_at, e = db.GetResponse(key.String())
				if e != nil || refresh {
					b, e = GetUserBuckets(co, user, ctx)
					if e != nil {
						logrus.Error(e)
						continue
					}
					if !nocache {
						value_b, e := json.Marshal(b)
						if e != nil {
							logrus.Warn("Failed marshal users data: ", e)
						} else {
							db.SaveResponse(key.String(), value_b)
						}
					}

				} else {
					// Првоеряем время хранениея данных БД
					// Позволительно 2 минуты
					if time.Now().After(updated_at.Add(2 * time.Minute)) {
						b, e = GetUserBuckets(co, user, ctx)
						if e != nil {
							logrus.Error(e)
							continue
						}
						if !nocache {
							value_b, e := json.Marshal(b)
							if e != nil {
								logrus.Warn("Failed marshal users data: ", e)
							} else {
								db.SaveResponse(key.String(), value_b)
							}
						}

					} else {
						e := json.Unmarshal(value.([]byte), &b)
						if e != nil {
							logrus.Error(e)
							b, e = GetUserBuckets(co, user, ctx)
							if e != nil {
								logrus.Error(e)
								continue
							}
							if !nocache {
								value_b, e := json.Marshal(b)
								if e != nil {
									logrus.Warn("Failed marshal users data: ", e)
								} else {
									db.SaveResponse(key.String(), value_b)
								}
							}

						}
					}
				}

				// Получаем информацию о бакетах
				for _, bucket := range b {
					bc, e := GetBucketInfo(co, ctx, bucket)
					u.Buckets = append(u.Buckets, struct {
						Error  error
						Bucket admin.Bucket
					}{
						Error:  e,
						Bucket: bc,
					})
				}
				u.Uid = user
				mu.Lock()
				rgwusers = append(rgwusers, u)
				mu.Unlock()
			}
			wg.Done()
		}(users_array, &wg, &mu)
	}
	wg.Wait() // Ожидаем завершения всех горутин
	return &rgwusers, nil
}

// Метод получения списка бакетов пользователя
func GetUserBuckets(co *admin.API, uuid string, ctx context.Context) ([]string, error) {
	buckets, e := co.ListUsersBuckets(ctx, uuid)
	return buckets, e
}

// Метод получения квоты пользователя
func GetUserQuota(co *admin.API, ctx context.Context, uid string) (admin.QuotaSpec, error) {
	q, e := co.GetUserQuota(ctx, admin.QuotaSpec{
		UID: uid,
	})
	return q, e
}

// Метод получения информации о бакете
func GetBucketInfo(co *admin.API, ctx context.Context, bucket string) (admin.Bucket, error) {
	b, e := co.GetBucketInfo(ctx, admin.Bucket{
		Bucket: bucket,
	})
	return b, e
}

// Метод получения детальной информации о пользователе
func GetUserInfo(co *admin.API, ctx context.Context, uid string) (admin.User, error) {
	u, e := co.GetUser(ctx, admin.User{
		ID: uid,
	})
	return u, e
}

// Метод печати в виде json
func UserJPrint(rgwusers []RgwUser, rgwuser string) error {
	if rgwuser != "" {
		for _, u := range rgwusers {
			if u.Uid == rgwuser {
				b, e := json.MarshalIndent(u, " ", " ")
				if e != nil {
					return e
				}
				fmt.Println(string(b))
				return nil
			}
		}
	}
	b, e := json.MarshalIndent(rgwusers, " ", " ")
	if e != nil {
		return e
	}
	fmt.Println(string(b))
	return nil
}

// Метод табличной печати
func UserTprint(rgwusers []RgwUser, rgwuser string, hsd string) {
	headerFmt := color.New(color.FgGreen, color.Bold).SprintfFunc()
	columnFmt := color.New(color.FgHiYellow, color.Bold).SprintfFunc()

	tbl := table.New("UID", "DISPLAY NAME", "USER TYPE", "SUSPEND", "MAX BUCKETS", "QUOTA STATUS", "QUOTA MAXSIZE", "QUOTA MAX OBJECTS", "BUCKETS COUNT", "BUKETS ACTUAL SIZE", "PLACEMENT", "STORAGE CLASS")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithHeaderSeparatorRow('-').WithPadding(2)

	for _, user := range rgwusers {
		if rgwuser != "" && rgwuser != user.Uid {
			continue
		}
		tbl.AddRow(user.Uid,
			user.UserInfo.Info.DisplayName, user.UserInfo.Info.Type,
			func() string {
				if *user.UserInfo.Info.Suspended == 1 {
					return "🔻 YES"
				}
				return "🟢 NO"
			}(),
			func() string {
				if *user.UserInfo.Info.MaxBuckets == -1 {
					return "❌ disabled"
				}
				return strconv.Itoa(*user.UserInfo.Info.MaxBuckets)
			}(),
			func() string {
				if *user.UserQuota.Enabled {
					return "✅ enabled"
				}
				return "❌ disabled"
			}(),
			func() string {
				if *user.UserQuota.MaxSizeKb == 0 {
					return "❌ disabled"
				}
				if hsd == "kb" {
					return strconv.Itoa(*user.UserQuota.MaxSizeKb) + " kb"
				}
				if hsd == "gb" {
					return strconv.Itoa(*user.UserQuota.MaxSizeKb/1_048_576) + " gb"
				}
				if hsd == "tb" {
					return strconv.Itoa(*user.UserQuota.MaxSizeKb/1073741824) + " tb"
				}
				return strconv.Itoa(*user.UserQuota.MaxSizeKb) + " kb"
			}(),
			func() string {
				if *user.UserQuota.MaxObjects == -1 {
					return "❌ disabled"
				}
				return strconv.FormatInt(*user.UserQuota.MaxObjects, 10)
			}(),
			len(user.Buckets),
			func() string {
				var size uint64 = 0
				for _, b := range user.Buckets {
					if b.Bucket.Usage.RgwMain.SizeKbActual != nil {
						size += *b.Bucket.Usage.RgwMain.SizeKbActual
					}
				}
				if hsd == "kb" {
					return strconv.FormatUint(size, 10) + " kb"
				}
				if hsd == "gb" {
					return strconv.FormatUint(size/1_048_576, 10) + " gb"
				}
				if hsd == "tb" {
					return strconv.FormatUint(size/1073741824, 10) + " tb"
				}
				return strconv.FormatUint(size, 10) + " kb"
			}(),
			user.UserInfo.Info.DefaultPlacement,
			user.UserInfo.Info.DefaultStorageClass,
		)
		// Добавляем subusers
		for _, su := range user.UserInfo.Info.Subusers {
			tbl.AddRow(
				su.Name,
				"➖ SU",
				"➖ SU",
				"➖ SU",
				"➖ SU",
				"➖ SU",
				"➖ SU",
				"➖ SU",
				"➖ SU",
				"➖ SU",
				"➖ SU",
				"➖ SU",
			)
		}
	}
	tbl.Print()
}

// Метод печати информации обо всех бакетах table формат
func PrintTBuckets(co *admin.API, rgwusers []RgwUser, rgwuser string, hsd string, bucketname string) {
	headerFmt := color.New(color.FgGreen, color.Bold).SprintfFunc()
	columnFmt := color.New(color.FgHiYellow, color.Bold).SprintfFunc()
	tbl := table.New("OWNER", "BUCKET", "ID", "TENANT", "QUOTA STATUS", "QUOTA MAXSIZE",
		"QUOTA MAX OBJECTS", "BUKET ACTUAL SIZE", "OBJECTS", "C TIME", "M TIME", "VERSIONED", "VERSIONING", "OBJECT LOCK",
		"DATA POOL", "INDEX POOL", "SHARDS",
	)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithHeaderSeparatorRow('-').WithPadding(2)

	for _, user := range rgwusers {
		// Проверка условия соответствия UUID пользователя для печати
		if rgwuser != "" && rgwuser != user.Uid {
			continue
		}
		for _, bucket := range user.Buckets {
			if bucket.Error != nil {
				logrus.Error(bucket.Error)
				continue
			}
			// Проверка соответствия имени бакета для печати
			if bucketname != "" && bucket.Bucket.Bucket != bucketname {
				continue
			}
			tbl.AddRow(
				bucket.Bucket.Owner,
				bucket.Bucket.Bucket,
				bucket.Bucket.ID,
				bucket.Bucket.Tenant,
				func() string {
					if *bucket.Bucket.BucketQuota.Enabled {
						return "✅ enabled"
					}
					return "❌ disabled"
				}(),
				func() string {
					if *bucket.Bucket.BucketQuota.MaxSizeKb == 0 {
						return "❌ disabled"
					}
					if hsd == "kb" {
						return strconv.Itoa(*bucket.Bucket.BucketQuota.MaxSizeKb) + " kb"
					}
					if hsd == "gb" {
						return strconv.Itoa(*bucket.Bucket.BucketQuota.MaxSizeKb/1_048_576) + " gb"
					}
					if hsd == "tb" {
						return strconv.Itoa(*bucket.Bucket.BucketQuota.MaxSizeKb/1073741824) + " tb"
					}
					return strconv.Itoa(*bucket.Bucket.BucketQuota.MaxSizeKb) + " kb"
				}(),
				func() string {
					if *bucket.Bucket.BucketQuota.MaxObjects == -1 {
						return "❌ disabled"
					}
					return strconv.FormatInt(*bucket.Bucket.BucketQuota.MaxObjects, 10)
				}(),
				func() string {
					if bucket.Bucket.Usage.RgwMain.SizeKbActual == nil {
						return "0 kb"
					}
					if hsd == "kb" {
						return strconv.FormatUint(*bucket.Bucket.Usage.RgwMain.SizeKbActual, 10) + " kb"
					}
					if hsd == "gb" {
						return strconv.FormatUint(*bucket.Bucket.Usage.RgwMain.SizeKbActual/1_048_576, 10) + " gb"
					}
					if hsd == "tb" {
						return strconv.FormatUint(*bucket.Bucket.Usage.RgwMain.SizeKbActual/1073741824, 10) + " tb"
					}
					return strconv.FormatUint(*bucket.Bucket.Usage.RgwMain.SizeKbActual, 10) + " kb"
				}(),
				func() string {
					if bucket.Bucket.Usage.RgwMain.NumObjects == nil {
						return "0"
					}
					return strconv.FormatUint(*bucket.Bucket.Usage.RgwMain.NumObjects, 10)
				}(),
				bucket.Bucket.CreationTime.Format("2006-01-02 15:04:05"),
				func() string {
					t, e := time.Parse("2006-01-02T15:04:05.000000Z", bucket.Bucket.Mtime)
					if e != nil {
						return "-"
					}
					return t.Format("2006-01-02 15:04:05")
				}(),
				func() string {
					if bucket.Bucket.Versioned == nil {
						return "-"
					}
					if *bucket.Bucket.Versioned {
						return "✅ enabled"
					}
					return "❌ disabled"
				}(),
				func() string {
					if bucket.Bucket.Versioning == nil {
						return "-"
					}
					if *bucket.Bucket.Versioning == "enabled" {
						return "✅ enabled"
					}
					return "❌ disabled"
				}(),
				func() string {
					if bucket.Bucket.ObjectLockEnabled {
						return "✅ enabled"
					}
					return "❌ disabled"
				}(),
				bucket.Bucket.ExplicitPlacement.DataPool,
				bucket.Bucket.ExplicitPlacement.IndexPool,
				*bucket.Bucket.NumShards,
			)
		}
	}
	tbl.Print()
}

// Метод печати информации обо всех бакетах json формат
func PrintJBuckets(co *admin.API, rgwusers []RgwUser, rgwuser string) {
	var buckets []struct {
		Error  error
		Bucket admin.Bucket
	}
	for _, user := range rgwusers {
		if rgwuser != "" && rgwuser != user.Uid {
			continue
		}
		for _, bucket := range user.Buckets {
			if bucket.Error != nil {
				logrus.Error(bucket.Error)
				continue
			}
			buckets = append(buckets, bucket)
		}
	}
	b, e := json.Marshal(buckets)
	if e != nil {
		logrus.Error(e)
		return
	}
	fmt.Println(string(b))
}

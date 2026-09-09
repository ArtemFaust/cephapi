package db

/*
Методы работы с локальной БД для кеширования запросов
*/

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

// Метод создания БД если она не существует
func CreateKVdatabase() {
	// Подключение к БД
	db, e := sql.Open("sqlite3", "./.kv.db")
	if e != nil {
		logrus.Error("Error create db connection: ", e)
	}
	defer db.Close()

	// Создание таблицы если ее нет
	_, e = db.Exec(`
        CREATE TABLE IF NOT EXISTS kv_store (
			key TEXT PRIMARY KEY NOT NULL,
			value BLOB,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
    `)
	if e != nil {
		logrus.Error("Error create response table: ", e)
	}
}

// Метод установки значения
func SaveResponse(key string, value []byte) {
	// Подключение к БД
	db, e := sql.Open("sqlite3", "./.kv.db")
	if e != nil {
		logrus.Error("Error create db connection: ", e)
	}
	defer db.Close()

	_, e = db.Exec(`
		INSERT OR REPLACE INTO kv_store (key, value) values (?, ?);
	`, key, value)
	if e != nil {
		logrus.Error("Error inser data to table for cache response: ", e)
	}
}

// Метод получения значения
// Возвращает поля value (BLOB) и updated_at time.Time
func GetResponse(key string) (any, time.Time, error) {
	// Подключение к БД
	file_db, e := sql.Open("sqlite3", "./.kv.db")
	if e != nil {
		logrus.Error("Error create db connection: ", e)
	}
	defer file_db.Close()

	mem_db, e := sql.Open("sqlite3", ":memory:?_journal_mode=WAL")
	if e != nil {
		logrus.Error("Error create db connection: ", e)
	}
	defer mem_db.Close()

	e = restoreDatabase(mem_db, file_db)

	var key_db any
	var value_db any
	var updated_at time.Time

	// Если не удалось восстановить bd в ram то читаем из файла
	if e != nil {
		e = file_db.QueryRow(`
		SELECT * FROM kv_store WHERE key = ?
	`, key).Scan(&key_db, &value_db, &updated_at)

		if e != nil {
			return nil, time.Now(), e
		}
		// Если восстановили в ram успешно то читаем из ram
	} else {
		e = mem_db.QueryRow(`
		SELECT * FROM kv_store WHERE key = ?
	`, key).Scan(&key_db, &value_db, &updated_at)

		if e != nil {
			return nil, time.Now(), e
		}
	}

	return value_db, updated_at, nil
}

// Метод восстановления БД из источника в БД назначения
// Реализация in memory db на уровне sqlite
func restoreDatabase(dstDB, srcDB *sql.DB) error {
	ctx := context.Background()

	// Подключаемся к БД назначения
	dstConn, err := dstDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get destination connection: %w", err)
	}
	defer dstConn.Close()

	// Подключаемся к исходной БД
	srcConn, err := srcDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get source connection: %w", err)
	}
	defer srcConn.Close()

	// Use Raw to bypass Go's sql wrapper and tap directly into the SQLite driver
	return dstConn.Raw(func(dstRaw any) error {
		dstSQLiteConn, ok := dstRaw.(*sqlite3.SQLiteConn)
		if !ok {
			return fmt.Errorf("destination is not a mattn/go-sqlite3 connection")
		}

		return srcConn.Raw(func(srcRaw any) error {
			srcSQLiteConn, ok := srcRaw.(*sqlite3.SQLiteConn)
			if !ok {
				return fmt.Errorf("source is not a mattn/go-sqlite3 connection")
			}

			// Initialize the Backup sequence: Backup(destination_schema, source_connection, source_schema)
			// SQLite schemas defaults to "main"
			backup, err := dstSQLiteConn.Backup("main", srcSQLiteConn, "main")
			if err != nil {
				return fmt.Errorf("failed to initialize backup API: %w", err)
			}

			// -1 copies the entire database in one single step
			_, err = backup.Step(-1)
			if err != nil {
				backup.Finish() // Clean up memory allocation even on failure
				return fmt.Errorf("failed during restore step: %w", err)
			}

			return backup.Finish()
		})
	})
}

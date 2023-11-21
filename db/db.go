package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/Dizzrt/dgo-torrent/dlog"
)

const (
	dbFilePath = ".data"
)

var db *sql.DB

func isTableExists(tableName string) bool {
	sql := fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name='%s';", tableName)

	rows, err := db.Query(sql)
	if err != nil {
		dlog.Fatal(err)
	}

	if rows.Next() {
		return true
	}

	return false
}

func initTables() {
	// tasks
	if !isTableExists("tasks") {
		_, err := db.Exec(_SQL_CREATE_TABLE_TASKS)
		if err != nil {
			dlog.Fatal(err)
		}
	}
}

func Init() {
	_ = DB()
	initTables()
}

func DB() *sql.DB {
	if db == nil || db.Ping() != nil {
		_db, err := sql.Open("sqlite3", dbFilePath)
		if err != nil {
			dlog.Fatal(err)
		}

		err = _db.Ping()
		if err != nil {
			dlog.Fatal(err)
		}

		db = _db
	}

	return db
}

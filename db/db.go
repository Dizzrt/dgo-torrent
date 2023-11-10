package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/Dizzrt/dgo-torrent/dlog"
)

const (
	path = ".data"

	_CREATE_TABLE_WORKS = `
		CREATE TABLE works (
			id INT PRIMARY KEY NOT NULL,
		);
	`
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
	} else {
		return false
	}
}

func initTables() {
	// works
	if !isTableExists("works") {
		// _, err := db.Exec(_CREATE_TABLE_WORKS)
		// if err != nil {
		// 	dlog.Fatal(err)
		// }
		// TODO init db
	}
}

func Init() {
	_ = DB()
	initTables()
}

func DB() *sql.DB {
	if db == nil || db.Ping() != nil {
		_db, err := sql.Open("sqlite3", path)
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

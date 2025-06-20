// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"
	"os"

	_ "embed"

	"github.com/mattn/go-sqlite3"
)

var sqlite3conn *sqlite3.SQLiteConn

func getMaxBlobSize() int64 {
	return int64(sqlite3conn.GetLimit(sqlite3.SQLITE_LIMIT_LENGTH))
}

func touch() (db *sql.DB, err error) {
	var exists = true

	if _, errStat := os.Stat(args.Path); errStat != nil {
		logger.info.Printf("file does not exist: creating;; %s\n", args.Path)
		exists = false
	}

	sql.Register("sqlite3_2", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			sqlite3conn = conn
			return nil
		},
	})

	db, err = sql.Open("sqlite3_2", args.Path)
	if err == nil {
		if !exists {
			_, err = db.Exec(schema)

			if err != nil {
				db.Close()
				db = nil

				err1 := os.Remove(args.Path)
				if err1 != nil {
					logger.err.Printf("created file is corrupted and could not be deleted: delete and do not use")
				}
			}
		}
	}

	return
}

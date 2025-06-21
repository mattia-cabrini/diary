// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "embed"
)

//go:embed res/head_dump_day.html
var pheadDumpDay string

func cmdDump(db *sql.DB) (err error) {
	years, err := querySingleInt64Array(db, "SELECT DISTINCT strftime('%Y', datetime(init, 'unixepoch')) from entries where deleted = 0")

	for _, yx := range years {
		dir := fmt.Sprintf("%d", yx)

		// | 100 at least the user should be able to x
		err = os.Mkdir(dir, os.FileMode(args.OutputPerm|100))

		if err == nil {
			logger.info.Printf("created directory %s", dir)
			err = dumpSingleYear(db, yx, dir)
		}
	}

	return
}

func dumpSingleYear(db *sql.DB, year int64, dir string) (err error) {
	months, err := querySingleInt64Array(db, "SELECT DISTINCT strftime('%m', datetime(init, 'unixepoch')) from entries where deleted = 0 AND strftime('%Y', datetime(init, 'unixepoch')) = cast(? as TEXT)", year)

	for _, mx := range months {
		dir = fmt.Sprintf("%s/%02d", dir, mx)

		// | 100 at least the user should be able to x
		err = os.Mkdir(dir, os.FileMode(args.OutputPerm|100))

		if err == nil {
			logger.info.Printf("created directory %s", dir)
			err = dumpSingleMonth(db, year, mx, dir)
		}
	}

	return
}

func dumpSingleMonth(db *sql.DB, year int64, month int64, dir string) (err error) {
	days, err := querySingleInt64Array(db, "SELECT DISTINCT strftime('%d', datetime(init, 'unixepoch')) from entries where deleted = 0 AND strftime('%Y', datetime(init, 'unixepoch')) = cast(? as TEXT) AND CAST(strftime('%m', datetime(init, 'unixepoch')) AS INTEGER) = ?", year, month)

	for _, dx := range days {
		var dirX = fmt.Sprintf("%s/%-2d", dir, dx)

		// | 100 at least the user should be able to x
		err = os.Mkdir(dirX, os.FileMode(args.OutputPerm|100))

		if err == nil {
			logger.info.Printf("created directory %s", dirX)
			err = dumpSingleDay(db, year, month, dx, dirX)
		}
	}

	return
}

func dumpSingleDay(db *sql.DB, year int64, month int64, day int64, dir string) (err error) {
	args.DateInit = time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.Now().Location())

	tmpWD, err := os.Getwd()
	if err == nil {

		err = os.Chdir(dir)
		if err == nil {

			err = cmdDumpDay(db)
			if err == nil {

				err = os.Chdir(tmpWD)
			}
		}
	}
	return
}

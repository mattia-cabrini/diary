// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "embed"
)

//go:embed res/head_dump_index.html
var headDumpIndex string

func cmdDump(db *sql.DB) (err error) {
	years, err := querySingleInt64Array(db, "SELECT DISTINCT strftime('%Y', datetime(init, 'unixepoch')) from entries where deleted = 0")
	if err != nil {
		return
	}

	fp, err := os.OpenFile("index.html", os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.FileMode(args.OutputPerm))
	if err != nil {
		return
	}

	defer fp.Close()
	_, err = fmt.Fprintf(fp, "<html>"+headDumpIndex+"<body><ul>", "Diary Dump")
	if err != nil {
		return
	}

	for _, yx := range years {
		dir := fmt.Sprintf("%d", yx)

		_, err = fmt.Fprintf(fp, "<li><a href=\"%s/index.html\">%s</a>\n", dir, dir)

		if err == nil {
			if args.Force {
				err = rmR(dir, true)
			}
		}

		if err == nil {
			err = createDirectoryIfNE(dir)
		}

		if err == nil {
			err = dumpSingleYear(db, yx, dir)
		}

		if err != nil {
			break
		}
	}

	if err == nil {
		_, err = fmt.Fprintf(fp, "</ul></body></html>")
	}

	return
}

func dumpSingleYear(db *sql.DB, year int64, dir string) (err error) {
	months, err := querySingleInt64Array(db, "SELECT DISTINCT strftime('%m', datetime(init, 'unixepoch')) from entries where deleted = 0 AND strftime('%Y', datetime(init, 'unixepoch')) = cast(? as TEXT)", year)
	if err != nil {
		return
	}

	fp, err := os.OpenFile(dir+"/index.html", os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.FileMode(args.OutputPerm))
	if err != nil {
		return
	}

	defer fp.Close()
	_, err = fmt.Fprintf(fp, "<html>"+headDumpIndex+"<body><ul>", dir)
	if err != nil {
		return
	}

	for _, mx := range months {
		var dirX = fmt.Sprintf("%s/%02d", dir, mx)

		_, err = fmt.Fprintf(fp, "<li><a href=\"../%s/index.html\">%s</a>\n", dirX, dirX)

		if err == nil {
			err = createDirectoryIfNE(dirX)
		}

		if err == nil {
			err = dumpSingleMonth(db, year, mx, dirX)
		}

		if err != nil {
			break
		}
	}

	if err == nil {
		_, err = fmt.Fprintf(fp, "</ul></body></html>")
	}

	return
}

func dumpSingleMonth(db *sql.DB, year int64, month int64, dir string) (err error) {
	days, err := querySingleInt64Array(db, "SELECT DISTINCT strftime('%d', datetime(init, 'unixepoch')) from entries where deleted = 0 AND strftime('%Y', datetime(init, 'unixepoch')) = cast(? as TEXT) AND CAST(strftime('%m', datetime(init, 'unixepoch')) AS INTEGER) = ?", year, month)
	if err != nil {
		return
	}

	fp, err := os.OpenFile(dir+"/index.html", os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.FileMode(args.OutputPerm))
	if err != nil {
		return
	}

	defer fp.Close()
	_, err = fmt.Fprintf(fp, "<html>"+headDumpDay+"<body><ul>", dir)
	if err != nil {
		return
	}

	for _, dx := range days {
		var dirX = fmt.Sprintf("%s/%02d", dir, dx)

		_, err = fmt.Fprintf(fp, "<li><a href=\"../../%s/index.html\">%s</a>\n", dirX, dirX)

		if err == nil {
			err = createDirectoryIfNE(dirX)
		}

		if err == nil {
			err = dumpSingleDay(db, year, month, dx, dirX)
		}

		if err != nil {
			break
		}
	}

	if err == nil {
		_, err = fmt.Fprintf(fp, "</ul></body></html>")
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

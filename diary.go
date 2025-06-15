// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"os/exec"
	"strings"
	"time"

	_ "embed"

	"github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schema string

var sqlite3conn *sqlite3.SQLiteConn

var args struct {
	Path    string
	Command string
	Help    bool

	Id         int64
	DateInit   time.Time
	DateEnd    time.Time
	Note       string
	NoAttach   bool
	OutputFile *os.File
	OutputPerm int

	// unchecked input
	OutputFileStr string
	OutputPermStr string
	DateInitStr   string
	DateEndStr    string
	TimeInitStr   string
	TimeEndStr    string
}

var logger struct {
	info *log.Logger
	err  *log.Logger
	warn *log.Logger
}

// BEGIN UTILITY FUNCTIONS

func getMaxBlobSize() int64 {
	return int64(sqlite3conn.GetLimit(sqlite3.SQLITE_LIMIT_LENGTH))
}

func getRandomString() (s string) {
	const a = int('a')
	const span = int('z') - a

	for range 24 {
		x := rand.Int() % span
		s = s + string(rune(x+a))
	}

	return s
}

func readAllFileContent(path string) (text string, err error) {
	fp, err := os.OpenFile(path, os.O_RDONLY, 0444)

	if err != nil {
		return
	}

	var sb = strings.Builder{}
	var buf = make([]byte, 65536)
	var n = -1

	for n, err = fp.Read(buf); err == nil && n != 0; n, err = fp.Read(buf) {
		sb.Write(buf)
	}

	if err.Error() == "EOF" {
		err = nil
	}

	text = sb.String()
	return
}

func editor() (text string, err error) {
	fileName := "diary_" + getRandomString()
	cmd := exec.Command("vim", fileName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err == nil {
		text, err = readAllFileContent(fileName)
		err1 := os.Remove(fileName)
		if err1 != nil {
			logger.err.Printf("could not remove temp file: %s", fileName)
		}
	}

	return
}

var molt = []string{"", "k", "M", "G"}

func sizeNorm[T int | int32 | int64](s T) string {
	var sz = float64(s)
	var i = 0

	for sz > 1000 && i < len(molt) {
		sz /= 1000.0
		i++
	}

	return fmt.Sprintf("%.3f %sB", sz, molt[i])
}

func printLine(n int, ch rune) {
	for range n {
		print(string(ch))
	}
	println()
}

func permStrToInt(perm string) (permInt int, err error) {
	const v0 = int('0')

	if len(perm) != 3 {
		err = fmt.Errorf("invalid permissions: \"%s\"", perm)
		return
	}

	for i, rx := range perm {
		if rxInt := int(rx); rxInt >= v0 && rxInt <= int('9') {
			permInt += rxInt - v0
			permInt *= 8 // octal shift
		} else {
			err = fmt.Errorf("invalid permissions, char '%s' is not allawed", string(perm[i]))
			return
		}
	}

	permInt /= 8
	return
}

func myerr(err error, fatal bool) {
	if err != nil {
		log.Println(err)

		if fatal {
			os.Exit(1)
		}
	}
}

// END UTILITY FUNCTIONS

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

// BEGING COMMMAND FUNCTIONS

func cmdAdd(db *sql.DB) (err error) {
	var reg = time.Now()
	var note = args.Note

	if note == "" {
		note, err = editor()
		if err != nil {
			return
		}
	}

	res, err := db.Exec("insert into entries (init, fin, inserted, note, deleted) values (?, ?, ?, ?, 0)", args.DateInit.Unix(), args.DateEnd.Unix(), reg.Unix(), note)
	if err != nil {
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		logger.info.Println("Inserted, could not retrieve id")
		return
	} else {
		logger.info.Printf("Inserted, with id #%d", id)
	}

	if !args.NoAttach {
		askForAttachments(db, id)
	}

	return
}

func askForAttachments(db *sql.DB, id int64) {
	var k = bufio.NewScanner(os.Stdin)
	var buf = make([]byte, 0)

	for {
		var errF error
		var fp *os.File

		print("Attachment: ")
		k.Scan()

		if k.Text() == "" {
			break
		}

		var attachmentPath = k.Text()

		stat, errStat := os.Stat(attachmentPath)
		if errStat != nil {
			logger.warn.Printf("file does not exist: %s\n", attachmentPath)
			continue
		}

		fp, errF = os.OpenFile(attachmentPath, os.O_RDONLY, 0444)
		if errF != nil {
			logger.err.Printf("could not open file: %v\n", errF)
			continue
		}

		if stat.Size() > getMaxBlobSize() {
			logger.err.Printf("file too big: max %s\n", sizeNorm(getMaxBlobSize()))
			continue
		}

		clear(buf)
		buf = make([]byte, stat.Size())

		_, errF = fp.Read(buf)
		if errF != nil {
			logger.err.Printf("could not open file: %v\n", errF)
			continue
		}

		_, errF = db.Exec("insert into attachments (name, inserted, content, entry_id) values (?, ?, ?, ?)", stat.Name(), time.Now().Unix(), buf, id)
		if errF != nil {
			logger.err.Printf("could not store file: %v\n", errF)
			continue
		}
		logger.info.Printf("Attached %s (%s)\n", stat.Name(), sizeNorm(len(buf)))
	}
}

func cmdResume(db *sql.DB) (err error) {
	date, _ := time.ParseInLocation(time.DateOnly, args.DateInit.Format(time.DateOnly), time.Now().Location())

	rows, err := db.Query("select id, init, fin, inserted, note from entries where init >= ? and init < ? and deleted = 0 order by init", date.Unix(), date.Add(24*time.Hour).Unix())
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() && err == nil {
		var idIn int64
		var initIn int64
		var endIn int64
		var insertedIn int64
		var noteIn string
		var attachmentCount int

		err = rows.Scan(&idIn, &initIn, &endIn, &insertedIn, &noteIn)

		n, _ := fmt.Printf("[%d] %s --> %s\n", idIn, time.Unix(initIn, 0).Format(time.DateTime), time.Unix(endIn, 0).Format(time.DateTime))
		printLine(n, '-')

		fmt.Printf("%s\n", noteIn)

		rowsAttachments, err2 := db.Query("select id, name, length(content) from attachments where entry_id = ? order by inserted", idIn)
		if err2 != nil {
			err = err2
			return
		}

		for attachmentCount = 0; rowsAttachments.Next() && err == nil; attachmentCount++ {
			var attIdIn int64
			var nameIn string
			var lengthIn int64

			err = rowsAttachments.Scan(&attIdIn, &nameIn, &lengthIn)

			if attachmentCount == 0 {
				printLine(n, '-')
				fmt.Println("Attachments:")
			}

			fmt.Printf("[%d] %s (%s)\n", attIdIn, nameIn, sizeNorm(lengthIn))
		}

		rowsAttachments.Close()
		if attachmentCount > 0 {
			printLine(n, '-')
		}

		println()
	}

	return
}

func cmdDelete(db *sql.DB) (err error) {
	var aff int64

	if args.Id < 0 {
		err = errors.New("invalid id")
		return
	}

	res, err := db.Exec("UPDATE entries set deleted = 1 where id = ?", args.Id)
	if err == nil {
		aff, err = res.RowsAffected()

		if err == nil {
			logger.info.Printf("%d row(s) deleted\n", aff)
		}
	}

	return
}

func cmdFetch(db *sql.DB) (err error) {
	var lengthIn int64
	var buf []byte

	if args.Id < 0 {
		err = errors.New("invalid id")
		return
	}

	row := db.QueryRow("SELECT length(content) from attachments where id = ?", args.Id)
	err = row.Scan(&lengthIn)

	if err == nil {
		buf = make([]byte, lengthIn)

		row := db.QueryRow("SELECT content from attachments where id = ?", args.Id)
		err = row.Scan(&buf)

		if err == nil {
			args.OutputFile.Write(buf)
		}
	}

	return
}

// END COMMMAND FUNCTIONS

func parseArgs() (err error) {
	flag.StringVar(&args.Path, "path", "", "diary file path")
	flag.StringVar(&args.Command, "cmd", "", "command (add, resume, delete, fetch)")
	flag.StringVar(&args.Note, "note", "", "note to log into the diary")
	flag.Int64Var(&args.Id, "id", -1, "entry id")
	flag.BoolVar(&args.Help, "help", false, "show this menu")
	flag.BoolVar(&args.NoAttach, "na", false, "tells the program not to ask for attachments")
	flag.StringVar(&args.DateInitStr, "date-init", time.Now().Format(time.DateOnly), "init date for requested operation")
	flag.StringVar(&args.DateEndStr, "date-end", "", "end date for requested operation, if empty it's set equal tu date-init")
	flag.StringVar(&args.TimeInitStr, "time-init", time.Now().Format(time.TimeOnly), "init time for requested operation")
	flag.StringVar(&args.TimeEndStr, "time-end", "", "end time for requested operation, if empty it's set equal tu time-init")
	flag.StringVar(&args.OutputFileStr, "output", "", "output file path (default: stdout)")
	flag.StringVar(&args.OutputPermStr, "operm", "660", "output file path permission")

	flag.Parse()

	args.DateInit, err = time.ParseInLocation(time.DateTime, args.DateInitStr+" "+args.TimeInitStr, time.Now().Location())
	if err != nil {
		return fmt.Errorf("datetime init: %s", err.Error())
	}

	if args.DateEndStr == "" {
		args.DateEndStr = args.DateInitStr
	}
	if args.TimeEndStr == "" {
		args.TimeEndStr = args.TimeInitStr
	}
	args.DateEnd, err = time.ParseInLocation(time.DateTime, args.DateEndStr+" "+args.TimeEndStr, time.Now().Location())
	if err != nil {
		return fmt.Errorf("datetime end: %s", err.Error())
	}

	args.OutputPerm, err = permStrToInt(args.OutputPermStr)
	if err != nil {
		return
	}

	if args.OutputFileStr == "" {
		args.OutputFile = os.Stdout
	} else {
		args.OutputFile, err = os.OpenFile(args.OutputFileStr, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(args.OutputPerm))
	}

	return
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
	log.SetPrefix("[ERROR] ")

	logger.info = log.New(os.Stdout, "[INFO ] ", 0)
	logger.warn = log.New(os.Stdout, "[WARN ] ", 0)
	logger.err = log.New(os.Stderr, "[ERROR] ", 0)

	err := parseArgs()
	myerr(err, true)

	if args.Help {
		flag.Usage()
		os.Exit(0)
	}

	db, err := touch()
	myerr(err, true)
	defer db.Close()

	switch strings.ToLower(args.Command) {
	case "add":
		err = cmdAdd(db)
	case "resume":
		err = cmdResume(db)
	case "delete":
		err = cmdDelete(db)
	case "fetch":
		err = cmdFetch(db)
	default:
		logger.err.Printf("invalid command: %s", args.Command)
	}

	myerr(err, false)
}

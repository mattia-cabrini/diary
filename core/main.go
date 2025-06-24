// SPDX-License-Identifier: MIT

package diary

import (
	"log"
	"os"
	"strings"

	_ "embed"
)

//go:embed res/help.txt
var HELP_PAGE string

var stdnull *os.File

func Run() {
	logger.info = log.New(os.Stderr, "[\033[34mINFO \033[0m] ", 0)
	logger.warn = log.New(os.Stderr, "[\033[33mWARN \033[0m] ", 0)
	logger.err = log.New(os.Stderr, "[\033[31mERROR\033[0m] ", 0)

	stdnull, _ := os.OpenFile("/dev/null", os.O_WRONLY, 0400)
	defer stdnull.Close()

	err := parseArgs()
	defer args.Clear()
	myerr(err, true)

	if args.Help {
		print(HELP_PAGE)
		os.Exit(0)
	}

	if !args.Verbose {
		logger.info = log.New(stdnull, "[\033[34mINFO \033[0m] ", 0)
	}

	db, err := touch()
	myerr(err, true)
	defer db.Close()

	switch strings.ToLower(args.Command) {
	case "add":
		err = cmdAdd(db)
	case "resume":
		err = cmdResume(db)
	case "dump-day":
		err = cmdDumpDay(db)
	case "dump":
		err = cmdDump(db)
	case "delete":
		err = cmdDelete(db)
	case "fetch":
		err = cmdFetch(db)
	case "license":
		err = cmdLicense(db)
	case "info":
		err = cmdInfo(db)
	default:
		logger.err.Printf("invalid command: %s", args.Command)
	}

	myerr(err, false)
}

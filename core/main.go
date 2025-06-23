// SPDX-License-Identifier: MIT

package diary

import (
	"flag"
	"log"
	"os"
	"strings"

	_ "embed"
)

func Run() {
	logger.info = log.New(os.Stderr, "[\033[34mINFO \033[0m] ", 0)
	logger.warn = log.New(os.Stderr, "[\033[33mWARN \033[0m] ", 0)
	logger.err = log.New(os.Stderr, "[\033[31mERROR\033[0m] ", 0)

	err := parseArgs()
	defer args.Clear()
	myerr(err, true)

	if !args.Verbose {
		fp, _ := os.OpenFile("/dev/null", os.O_WRONLY, 0400)
		logger.info = log.New(fp, "[\033[34mINFO \033[0m] ", 0)
		defer fp.Close()
	}

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
	case "dump-day":
		err = cmdDumpDay(db)
	case "dump":
		err = cmdDump(db)
	case "delete":
		err = cmdDelete(db)
	case "fetch":
		err = cmdFetch(db)
	default:
		logger.err.Printf("invalid command: %s", args.Command)
	}

	myerr(err, false)
}

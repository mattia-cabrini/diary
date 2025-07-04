// SPDX-License-Identifier: MIT

package diary

import (
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
)

//go:embed res/schema.sql
var schema string

type arguments struct {
	Path    string
	Command string
	Help    bool
	Verbose bool
	Force   bool

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

var args arguments

var logger struct {
	info *log.Logger
	err  *log.Logger
	warn *log.Logger
}

var (
	NOT_FOUND = errors.New("not found")
)

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

func printLine(n int, ch rune, fp *os.File) {
	if fp == nil {
		fp = os.Stdout
	}

	for range n {
		fmt.Fprintf(fp, "%c", ch)
	}

	fmt.Fprintln(fp)
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
		logger.err.Println(err)

		if fatal {
			os.Exit(1)
		}
	}
}

func parseArgs() (err error) {
	var wd string

	f := flag.NewFlagSet("usage", flag.ContinueOnError)

	f.StringVar(&args.Path, "path", "", "diary file path")
	f.StringVar(&args.Command, "cmd", "", "command (add, resume, delete, fetch, dump-day, dump, license)")
	f.StringVar(&args.Note, "note", "", "note to log into the diary")
	f.Int64Var(&args.Id, "id", -1, "entry id")
	f.BoolVar(&args.Help, "help", false, "show this menu")
	f.BoolVar(&args.NoAttach, "na", false, "tells the program not to ask for attachments")
	f.StringVar(&args.DateInitStr, "di", time.Now().Format(time.DateOnly), "init date for requested operation")
	f.StringVar(&args.DateEndStr, "de", "", "end date for requested operation, if empty it's set equal tu date-init")
	f.StringVar(&args.TimeInitStr, "ti", time.Now().Format(time.TimeOnly), "init time for requested operation")
	f.StringVar(&args.TimeEndStr, "te", "", "end time for requested operation, if empty it's set equal tu time-init")
	f.StringVar(&args.OutputFileStr, "output", "", "output file path (default: stdout)")
	f.StringVar(&args.OutputPermStr, "operm", "660", "output file path permission")
	f.StringVar(&wd, "wd", "", "working directory")
	f.BoolVar(&args.Verbose, "v", false, "verbose info")
	f.BoolVar(&args.Force, "f", false, "force")

	out := f.Output()
	f.SetOutput(stdnull)
	defer func() {
		f.SetOutput(out)
	}()

	err = f.Parse(os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			err = nil
			args.Help = true
			return
		}
		return
	}

	if args.Command == "help" {
		args.Help = true
		return
	}

	if args.Force {
		logger.warn.Println("Using -f")
	}

	if wd != "" {
		err = os.Chdir(wd)
	}
	if err != nil {
		return
	}

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

	switch args.OutputFileStr {
	case "-":
		args.OutputFile = os.Stdout
	case "":
		break
	default:
		args.OutputFile, err = os.OpenFile(args.OutputFileStr, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(args.OutputPerm))
	}

	return
}

func (a arguments) Clear() {
	if a.OutputFile != nil && a.OutputFile != os.Stdout {
		a.OutputFile.Close()
	}
}

func rmR(path string, onlyChildren bool) (err error) {
	var ddee []os.DirEntry
	var stat os.FileInfo

	stat, err = os.Stat(path)

	if os.IsNotExist(err) {
		err = nil
		return
	}

	if err == nil && stat.IsDir() {

		ddee, err = os.ReadDir(path)
		if err == nil {
			for _, dex := range ddee {
				err = rmR(path+"/"+dex.Name(), false)

				if err != nil {
					break
				}
			}
		}
	}

	if err == nil && !onlyChildren {
		if stat.IsDir() {
			logger.info.Printf("Deleting directory \"%s\"", path)
		} else {
			logger.info.Printf("Deleting regular file \"%s\"", path)
		}

		err = os.Remove(path)
	}

	return
}

func createDirectoryIfNE(dir string) error {
	var ddee []os.DirEntry

	stat, err := os.Stat(dir)
	if err == nil {
		if !stat.IsDir() {
			err = errors.New(dir + " is not a directory")
		}

		if err == nil {
			ddee, err = os.ReadDir(dir)

			if len(ddee) > 0 {
				err = errors.New(dir + " not empty")
			}
		}
	} else if os.IsNotExist(err) {
		err = os.Mkdir(dir, os.FileMode(args.OutputPerm|0100))
		logger.info.Printf("created directory %s", dir)
	}

	return err
}

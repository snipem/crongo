package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli"
)

func runCommand(name string, args ...string) (stdout string, stderr string, exitCode int) {
	var outbuf, errbuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout = outbuf.String()
	stderr = errbuf.String()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			// Workaround for Mac
			exitCode = 1
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}
	return
}

func writeToDb(stdout string, stderr string, cmd string, errorCode int) {

	dbFile := os.Getenv("HOME") + "/crongo.db"
	log.Println("Accessing db in %s", dbFile)

	database, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS commands (id INTEGER PRIMARY KEY, sqltime TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL, cmd TEXT, stdout TEXT, stderr TEXT, error_code TEXT)")
	if err != nil {
		log.Fatal(err)
	}
	statement.Exec()
	statement, err = database.Prepare("INSERT INTO commands (cmd, stdout, stderr, error_code) VALUES (?, ?, ?, ?)")
	statement.Exec(cmd, stdout, stderr, errorCode)
}

func runCommandAndStoreIntoDatabase(cmd string) (exitCode int) {

	bash := "bash"
	args := []string{"-c", cmd}

	stdout, stderr, exitCode := runCommand(bash, args...)
	writeToDb(stdout, stderr, cmd, exitCode)

	fmt.Printf("stdout:\n%v\nstderr:\n%v\nexit_code: %v", stdout, stderr, exitCode)

	return exitCode

}

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:      "run",
			Aliases:   []string{"r"},
			Usage:     "run a command",
			ArgsUsage: "command to run",
			Action: func(c *cli.Context) error {
				if !c.IsSet("log") {
					// Don't log anything
					log.SetOutput(ioutil.Discard)
				}

				if len(c.Args()) != 1 {
					return cli.NewExitError("Command missing", 1)
				}
				exitCode := runCommandAndStoreIntoDatabase(c.Args().Get(0))

				// reflect exit code
				return cli.NewExitError("", exitCode)
			},
		},
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "log",
			Usage: "print log",
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

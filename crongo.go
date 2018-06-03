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
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/gosuri/uitable"
	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli"
)

var dbFile = os.Getenv("HOME") + "/crongo.db"

type command struct {
	id        int
	cmd       string
	date      *time.Time
	stdout    string
	stderr    string
	errorCode int
}

func runCommand(name string, args ...string) (c command) {
	var outbuf, errbuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	c.stdout = outbuf.String()
	c.stderr = errbuf.String()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			c.errorCode = ws.ExitStatus()
		} else {
			// Workaround for Mac
			c.errorCode = 1
			if c.stderr == "" {
				c.stderr = err.Error()
			}
		}
	} else {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		c.errorCode = ws.ExitStatus()
	}
	return c
}

func listAllRuns(limit int) {
	stmt := "select * from (select * from commands order by id DESC limit " + fmt.Sprint(limit) + ") order by id ASC"
	commands := runStatement(stmt)
	printCommands(commands)
}

func listAllFailedRuns(limit int) {
	stmt := "select * from (select * from commands order by id DESC limit " + fmt.Sprint(limit) + ") where error_code is not 0 order by id ASC"
	commands := runStatement(stmt)
	printCommands(commands)
}

func printCommands(commands []command) {

	table := uitable.New()
	table.MaxColWidth = 50
	statusDot := "◉"

	table.AddRow("CODE", "ID", "DATE", "CMD", "STDOUT", "STDERR")
	for _, command := range commands {
		statusLine := statusDot + " " + strconv.Itoa(command.errorCode)
		table.AddRow(statusLine, command.id, command.date.In(time.Local), command.cmd, command.stdout, command.stderr)
	}

	// Workaround: uitable counts non printable characters like colors, therefore garbeling the width of the table
	// paint all status codes red
	out := strings.Replace(table.String(), statusDot, color.RedString(statusDot), -1)
	// paint all red status codes with a follow up zero green
	out = strings.Replace(out, color.RedString(statusDot)+" 0", color.GreenString(statusDot)+" 0", -1)

	fmt.Println(out)
}

func runStatement(stmt string) []command {

	var commands []command

	database, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := database.Query(stmt)

	for rows.Next() {
		var c command
		err = rows.Scan(&c.id, &c.date, &c.cmd, &c.stdout, &c.stderr, &c.errorCode)
		if err != nil {
			log.Fatal(err)
		}
		commands = append(commands, c)
	}
	return commands
}

func writeToDb(c command) {

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
	statement.Exec(c.cmd, c.stdout, c.stderr, c.errorCode)
}

func runCommandAndStoreIntoDatabase(cmd string) (exitCode int) {

	bash := "bash"
	args := []string{"-c", cmd}

	c := runCommand(bash, args...)
	writeToDb(c)
	prettyPrintCommand(c)

	return exitCode

}

func prettyPrintCommand(c command) {
	fmt.Printf("stdout:\n%v\nstderr:\n%v\nexit_code: %v", c.stdout, c.stderr, c.errorCode)
}

func getCommandInfoFromDatabase(id int) {
	c := runStatement("select * from commands where id = " + fmt.Sprint(id))
	prettyPrintCommand(c[0])
}

func getLimit(c *cli.Context) (limit int) {

	if len(c.Args()) != 1 {
		return 500
	} else {
		limit, err := strconv.Atoi(c.Args().First())
		if err != nil {
			log.Fatal(err)
		}
		return limit
	}
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
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list runs",
			Subcommands: []cli.Command{
				{
					Name:  "all",
					Usage: "list all runs",
					Action: func(c *cli.Context) error {

						listAllRuns(getLimit(c))
						return nil
					},
				},
				{
					Name:  "failed",
					Usage: "list all failed runs",
					Action: func(c *cli.Context) error {
						listAllFailedRuns(getLimit(c))
						return nil
					},
				},
			},
		},
		{
			Name:      "id",
			Aliases:   []string{"i"},
			Usage:     "get info about command in database",
			ArgsUsage: "id of command",
			Action: func(c *cli.Context) error {
				if !c.IsSet("log") {
					// Don't log anything
					log.SetOutput(ioutil.Discard)
				}

				if len(c.Args()) != 1 {
					return cli.NewExitError("Command missing", 1)
				}
				id, _ := strconv.Atoi(c.Args().Get(0))
				getCommandInfoFromDatabase(id)
				return nil
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

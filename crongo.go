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

func listAllRuns(limit int, filter string) []command {
	filterAppendix := ""
	if filter != "" {
		filterAppendix = "where cmd like '%" + filter + "%'"
	}
	stmt := "select * from (select * from commands order by id DESC)  " + filterAppendix + " order by id ASC limit " + fmt.Sprint(limit) + ""
	return runStatement(stmt)
}

func listAllFailedRuns(limit int, filter string) []command {
	filterAppendix := ""
	if filter != "" {
		filterAppendix = "where cmd like '%" + filter + "%'"
	}
	stmt := "select * from (select * from commands order by id DESC)  " + filterAppendix + " order by id ASC limit " + fmt.Sprint(limit) + ""
	return runStatement(stmt)
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

	log.Printf("Accessing db in %s", dbFile)

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
	c.cmd = cmd
	writeToDb(c)
	prettyPrintCommand(c)

	return exitCode

}

func prettyPrintCommand(c command) {
	fmt.Printf("stdout:\n%v\nstderr:\n%v\nexit_code: %v", c.stdout, c.stderr, c.errorCode)
}

func getCommandInfoFromDatabase(id int) error {
	c := runStatement("select * from commands where id = " + fmt.Sprint(id))
	if len(c) > 0 {
		prettyPrintCommand(c[0])
		return nil
	}
	return fmt.Errorf("Command with id %s not found", strconv.Itoa(id))
}

func purgeDatabase(numberOfEntriesToKeep int) {
	runStatement(`
	DELETE FROM "commands"
	WHERE id NOT IN (
	  SELECT id
	  FROM (
		SELECT id
		FROM "commands"
		ORDER BY id DESC
		LIMIT ` + strconv.Itoa(numberOfEntriesToKeep) + `
	  ) purge
	);`)
	// Actually release occupied space
	runStatement("vacuum")
}

func main() {
	app := cli.NewApp()
	app.Version = "0.3.1"

	var limit int
	var filter string

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

						printCommands(listAllRuns(limit, filter))
						return nil
					},
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:        "limit",
							Value:       500,
							Usage:       "limit number of results",
							Destination: &limit,
						},
						cli.StringFlag{
							Name:        "filter",
							Usage:       "filter for command",
							Destination: &filter,
						},
					},
				},
				{
					Name:  "failed",
					Usage: "list all failed runs",
					Action: func(c *cli.Context) error {
						printCommands(listAllFailedRuns(limit, filter))
						return nil
					},
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:        "limit",
							Value:       500,
							Usage:       "limit number of results",
							Destination: &limit,
						},
						cli.StringFlag{
							Name:        "filter",
							Usage:       "filter for command",
							Destination: &filter,
						},
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
				err := getCommandInfoFromDatabase(id)
				if err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
				return nil
			},
		},
		{
			Name:      "purge",
			Aliases:   []string{"p"},
			Usage:     "purge all entries except the newest, default 100",
			ArgsUsage: "entries to purge",
			Action: func(c *cli.Context) error {

				if len(c.Args()) != 1 {
					purgeDatabase(100)
				} else if numberOfEntriesToKeep, err := strconv.Atoi(c.Args().Get(0)); err == nil {
					purgeDatabase(numberOfEntriesToKeep)
				} else {
					cli.NewExitError("Number of commands to keep", 1)
				}
				return nil
			},
		},
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "print debug log",
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

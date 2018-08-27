package main

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func Test_listAllRuns(t *testing.T) {
	dbFile = "test/crongo.db"

	randomDate, err := time.Parse("2006-01-02 15:04:05", "2018-08-23 07:00:00")
	if err != nil {
		log.Fatalf("Error converting random date %s", err)
	}

	type args struct {
		limit  int
		filter string
	}
	tests := []struct {
		name string
		args args
		want []command
	}{
		{
			name: "Limit 1",
			args: args{filter: "", limit: 2},
			want: []command{
				{
					id:        1,
					cmd:       "first",
					date:      &randomDate,
					stdout:    "stdout first",
					stderr:    "stderr first",
					errorCode: 1,
				},
				{
					id:        2,
					cmd:       "second",
					date:      &randomDate,
					stdout:    "stdout second",
					stderr:    "stderr second",
					errorCode: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := listAllRuns(tt.args.limit, tt.args.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listAllRuns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_listAllFailedRuns(t *testing.T) {
	dbFile = "test/crongo.db"

	randomDate, err := time.Parse("2006-01-02 15:04:05", "2018-08-23 07:00:00")

	if err != nil {
		log.Fatalf("Error converting random date %s", err)
	}
	type args struct {
		limit  int
		filter string
	}
	tests := []struct {
		name string
		args args
		want []command
	}{
		{
			name: "Limit 1",
			args: args{filter: "", limit: 1000},
			want: []command{
				{
					id:        1,
					cmd:       "first",
					date:      &randomDate,
					stdout:    "stdout first",
					stderr:    "stderr first",
					errorCode: 1,
				},
				{
					id:        2,
					cmd:       "second",
					date:      &randomDate,
					stdout:    "stdout second",
					stderr:    "stderr second",
					errorCode: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := listAllFailedRuns(tt.args.limit, tt.args.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listAllFailedRuns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_runCommand(t *testing.T) {

	command := runCommand("echo -n test")
	if !(command.cmd == "echo -n test" && command.stdout == "test" && command.errorCode >= 0 && command.id >= 0) {
		t.Errorf("runCommand() = is not valid: %v", command)
	}

}

func Test_runCommandAndStoreItIntoDatabase(t *testing.T) {
	tempFile, err := ioutil.TempFile("test/temp/", "runcommandandstoreitindatabase")
	if err != nil {
		log.Fatal(err)
	}

	dbFile = tempFile.Name()
	exitCode := runCommandAndStoreIntoDatabase("echo test")
	assert.Equal(t, 0, exitCode, "The error code was not zero")
	assert.FileExists(t, dbFile, "The newly created database file does not exist")
	err = getCommandInfoFromDatabase(1)
	assert.Nil(t, err, "Error is not nil")

	os.Remove(dbFile)
}

func Test_formatCommands(t *testing.T) {
	randomDate, _ := time.Parse("2006-01-02 15:04:05", "2018-08-23 07:00:00")

	commands := []command{
		{
			id:        1,
			cmd:       "first",
			date:      &randomDate,
			stdout:    "stdout",
			stderr:    "stderr",
			errorCode: 1,
		},
		{
			id:        2,
			cmd:       "second",
			date:      &randomDate,
			stdout:    "stdout",
			stderr:    "stderr",
			errorCode: 1,
		},
	}
	formattedString := formatCommands(commands)
	assert.Contains(t, formattedString, "first")
	assert.Contains(t, formattedString, "second")
}

// [{1 first 0xc4200b25a0 stdout first stderr first 1} {2 second 0xc4200b2660 stdout second stderr second 1} {3 second 0xc4200b2720 stdout second stderr second 0}], want
// [{1 first 0xc4200b2400 stdout first stderr first 1} {2 second 0xc4200b2400 stdout second stderr second 1}]

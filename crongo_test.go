package main

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/kami-zh/go-capturer"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

var testFolder = "test/temp/"

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
					id:        2,
					cmd:       "second",
					date:      &randomDate,
					stdout:    "stdout second",
					stderr:    "stderr second",
					errorCode: 1,
				},
				{
					id:        3,
					cmd:       "second",
					date:      &randomDate,
					stdout:    "stdout second",
					stderr:    "stderr second",
					errorCode: 0,
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
	tempFile, _ := ioutil.TempFile(testFolder, t.Name())
	dbFile = tempFile.Name()

	exitCode := runCommandAndStoreIntoDatabase("echo test")
	assert.Equal(t, 0, exitCode, "The error code was not zero")
	assert.FileExists(t, dbFile, "The newly created database file does not exist")
	err := getCommandInfoFromDatabase(1)
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

func Test_listAllFilesOfMany(t *testing.T) {
	tempFile, _ := ioutil.TempFile(testFolder, t.Name())
	dbFile = tempFile.Name()

	for index := 1; index <= 50; index++ {
		runCommandAndStoreIntoDatabase("echo " + strconv.Itoa(index))
	}

	// This should return the newst and not the oldest
	commands := listAllRuns(10, "")

	assert.Contains(t, commands[0].cmd, "echo 41")
	assert.Contains(t, commands[1].cmd, "echo 42")
	assert.Contains(t, commands[2].cmd, "echo 43")
	assert.Contains(t, commands[3].cmd, "echo 44")
	assert.Contains(t, commands[4].cmd, "echo 45")
	assert.Contains(t, commands[5].cmd, "echo 46")
	assert.Contains(t, commands[6].cmd, "echo 47")
	assert.Contains(t, commands[7].cmd, "echo 48")
	assert.Contains(t, commands[8].cmd, "echo 49")
	assert.Contains(t, commands[9].cmd, "echo 50")
}

func Test_listAllWithFilter(t *testing.T) {
	tempFile, _ := ioutil.TempFile(testFolder, t.Name())
	dbFile = tempFile.Name()

	runCommandAndStoreIntoDatabase("echo 1")
	runCommandAndStoreIntoDatabase("print 2")

	// This should return the newst and not the oldest
	commands := listAllRuns(10, "print")

	assert.Contains(t, commands[0].cmd, "print 2")
}

func Test_listAllFailedWithFilter(t *testing.T) {
	tempFile, _ := ioutil.TempFile(testFolder, t.Name())
	dbFile = tempFile.Name()

	runCommandAndStoreIntoDatabase("echo 1")
	runCommandAndStoreIntoDatabase("NOT_EXISTING_COMMAND_NVER_jfdhgjhdg 2")
	runCommandAndStoreIntoDatabase("NOT_EXISTING_SECOND_COMMAND_NVER_jfdhgjhdg 2")

	// This should return the newst and not the oldest
	commands := listAllFailedRuns(10, "NOT_EXISTING_COMMAND_NVER_jfdhgjhdg")

	assert.Len(t, commands, 1)
	assert.Contains(t, commands[0].cmd, "NOT_EXISTING_COMMAND_NVER_jfdhgjhdg 2")
}

func Test_getId(t *testing.T) {
	tempFile, _ := ioutil.TempFile(testFolder, t.Name())
	dbFile = tempFile.Name()

	runCommandAndStoreIntoDatabase("echo show me in result")

	// This should return the newst and not the oldest
	commands := listAllRuns(10, "")

	id := commands[0].id

	args := []string{"crongo", "id", strconv.Itoa(id)}
	run(args)

	out := capturer.CaptureStdout(func() {
		run(args)
	})

	assert.Equal(t, out, `stdout:
show me in result

stderr:

exit_code: 0
`)
}

func Test_purgeDatabase(t *testing.T) {
	tempFile, _ := ioutil.TempFile(testFolder, t.Name())
	dbFile = tempFile.Name()

	numFilesBefore := 20
	numFilesAfter := 10

	for index := 1; index <= numFilesBefore; index++ {
		runCommandAndStoreIntoDatabase("echo " + strconv.Itoa(index))
	}
	commands := listAllRuns(numFilesBefore, "")
	assert.Len(t, commands, numFilesBefore)

	args := []string{"crongo", "purge", strconv.Itoa(numFilesAfter)}
	run(args)

	lessCommands := listAllRuns(numFilesBefore, "")
	assert.Len(t, lessCommands, numFilesAfter)
}

func Test_purgeDefaultDatabase(t *testing.T) {
	tempFile, _ := ioutil.TempFile(testFolder, t.Name())
	dbFile = tempFile.Name()

	numFilesBefore := 120
	numFilesAfter := 100

	for index := 1; index <= numFilesBefore; index++ {
		runCommandAndStoreIntoDatabase("echo " + strconv.Itoa(index))
	}
	commands := listAllRuns(numFilesBefore, "")
	assert.Len(t, commands, numFilesBefore)

	args := []string{"crongo", "purge"}
	run(args)

	lessCommands := listAllRuns(numFilesBefore, "")
	assert.Len(t, lessCommands, numFilesAfter)
}

func Test_runAndList(t *testing.T) {
	tempFile, _ := ioutil.TempFile(testFolder, t.Name())
	dbFile = tempFile.Name()

	fakeExit := func(exitCode int) {
		log.Printf("os.Exit called with %s", strconv.Itoa(exitCode))
	}

	if runtime.GOOS == "darwin" {
		t.Skip("Mocking os.Exit does not work on Mac")
	}

	patch := monkey.Patch(os.Exit, fakeExit)
	argsRun := []string{"crongo", "run", "echo hello world"}
	run(argsRun)

	argsList := []string{"crongo", "list", "all"}

	out := capturer.CaptureStdout(func() {
		defer patch.Unpatch()
		run(argsList)
	})

	assert.Contains(t, out, "echo hello world")
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func TestMain(m *testing.M) {
	if _, err := os.Stat(testFolder); os.IsNotExist(err) {
		err := os.Mkdir(testFolder, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	os.Exit(m.Run())
}

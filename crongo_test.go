package main

import (
	"log"
	"reflect"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
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
					errorCode: 0,
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

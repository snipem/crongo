package main

import (
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func Test_listAllRuns(t *testing.T) {
	dbFile = "test/crongo.db"

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
			args: args{filter: "", limit: 1},
			want: []command{
				{
					id:        1,
					cmd:       "first",
					date:      nil,
					stdout:    "stdout first",
					stderr:    "stderr first",
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

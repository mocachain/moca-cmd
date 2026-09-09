package main

import (
	"testing"

	"github.com/urfave/cli/v2"
)

// runObjectSubcommand parses args through a throwaway app with the subcommand's
// action swapped out, so flag handling is exercised without a chain client.
func runObjectSubcommand(t *testing.T, cmd *cli.Command, args []string, action cli.ActionFunc) error {
	t.Helper()
	cmd.Action = action
	app := &cli.App{
		Name:     "moca-cmd",
		Commands: []*cli.Command{{Name: "object", Subcommands: []*cli.Command{cmd}}},
	}
	return app.Run(append([]string{"moca-cmd", "object"}, args...))
}

func Test_delegateFlag(t *testing.T) {
	tests := []struct {
		name string
		cmd  func() *cli.Command
		args []string
		want bool
	}{
		{
			name: "object put defaults to a local createObject txn",
			cmd:  cmdPutObj,
			args: []string{"put", "file.txt", "moca://moca-bucket/moca-object"},
			want: false,
		},
		{
			name: "object put --delegate",
			cmd:  cmdPutObj,
			args: []string{"put", "--delegate", "--contentType", "text/plain", "file.txt", "moca://moca-bucket/moca-object"},
			want: true,
		},
		{
			name: "object update defaults to a visibility update",
			cmd:  cmdUpdateObject,
			args: []string{"update", "--visibility=public-read", "moca://moca-bucket/moca-object"},
			want: false,
		},
		{
			name: "object update --delegate",
			cmd:  cmdUpdateObject,
			args: []string{"update", "--delegate", "--bypassSeal", "file.txt", "moca://moca-bucket/moca-object"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			err := runObjectSubcommand(t, tt.cmd(), tt.args, func(ctx *cli.Context) error {
				got = ctx.Bool(delegateFlag)
				return nil
			})
			if err != nil {
				t.Fatalf("run %v: unexpected error: %v", tt.args, err)
			}
			if got != tt.want {
				t.Errorf("run %v: delegate = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

// The delegated update reuses the upload option flags of object put.
func Test_updateObjectDelegateFlags(t *testing.T) {
	for _, name := range []string{delegateFlag, contentTypeFlag, partSizeFlag, resumableFlag, bypassSealFlag, visibilityFlag} {
		found := false
		for _, f := range cmdUpdateObject().Flags {
			for _, n := range f.Names() {
				if n == name {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("object update is missing the --%s flag", name)
		}
	}
}

package main

import (
	"flag"
	"fmt"
	"os"

	fluffyssh "github.com/odvcencio/fluffyui/ssh"
)

func runSSH(args []string) error {
	fs := flag.NewFlagSet("ssh", flag.ContinueOnError)
	addr := fs.String("addr", ":2222", "listen address")
	hostKey := fs.String("host-key", "", "path to host key file")
	password := fs.String("password", "", "password for authentication")
	noAuth := fs.Bool("no-auth", false, "disable authentication (dev only)")
	fs.SetOutput(os.Stderr)

	if err := fs.Parse(args); err != nil {
		return err
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: fluffy ssh [flags] -- <command> [args...]")
		fmt.Fprintln(os.Stderr, "\nServes a FluffyUI application over SSH.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
		return fmt.Errorf("no command specified")
	}

	srv, err := fluffyssh.New(fluffyssh.Config{
		Addr:        *addr,
		HostKeyPath: *hostKey,
		Password:    *password,
		NoAuth:      *noAuth,
		Command:     remaining,
	})
	if err != nil {
		return err
	}

	return srv.ListenAndServe()
}

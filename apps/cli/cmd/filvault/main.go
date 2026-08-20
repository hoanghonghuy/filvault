package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"filvault/cli/internal/client"
	"filvault/cli/internal/command"
	"filvault/cli/internal/config"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usage()
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	apiBase := cfg.APIBase
	if apiBase == "" {
		apiBase = os.Getenv("FILVAULT_API_BASE")
	}
	if apiBase == "" {
		apiBase = config.DefaultAPIBase
	}

	c := client.New(apiBase)
	c.Access = cfg.AccessToken
	c.Refresh = cfg.RefreshToken
	c.OnRefresh = func(access, refresh string) {
		cfg.AccessToken = access
		cfg.RefreshToken = refresh
		_ = config.Save(cfg)
	}
	c.OnAuthFail = func() {
		cfg.AccessToken = ""
		cfg.RefreshToken = ""
		_ = config.Save(cfg)
	}

	cmds := &command.Commands{Client: c, Out: os.Stdout, Err: os.Stderr}

	switch args[0] {
	case "login":
		return login(cmds, cfg)
	case "whoami":
		return cmds.Whoami()
	case "ls":
		folderID := ""
		if len(args) > 1 {
			folderID = args[1]
		}
		return cmds.Ls(folderID)
	case "upload":
		return upload(cmds, args[1:])
	case "download":
		if len(args) < 2 {
			return fmt.Errorf("usage: filvault download <name>")
		}
		return cmds.Download(args[1])
	default:
		return usage()
	}
}

func login(cmds *command.Commands, cfg config.Config) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprint(os.Stderr, "email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Fprint(os.Stderr, "password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	return cmds.Login(email, password, func(access, refresh string) error {
		cfg.AccessToken = access
		cfg.RefreshToken = refresh
		cfg.APIBase = cmds.Client.Base
		return config.Save(cfg)
	})
}

func upload(cmds *command.Commands, args []string) error {
	fs := flag.NewFlagSet("upload", flag.ContinueOnError)
	folder := fs.String("folder", "", "target folder id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return fmt.Errorf("usage: filvault upload <file> [--folder <id>]")
	}
	return cmds.Upload(rest[0], *folder)
}

func usage() error {
	return fmt.Errorf(`usage:
  filvault login
  filvault whoami
  filvault ls [folderId]
  filvault upload <file> [--folder <id>]
  filvault download <name>`)
}

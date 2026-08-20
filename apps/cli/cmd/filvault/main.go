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

	"golang.org/x/term"
)

const version = "0.1.0"

// exitCode distinguishes usage errors (2) from runtime errors (1).
type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string { return e.err.Error() }

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if ee, ok := err.(*exitError); ok {
			os.Exit(ee.code)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	switch args[0] {
	case "-h", "--help", "help":
		printUsage(os.Stdout)
		return nil
	case "-v", "--version", "version":
		fmt.Fprintf(os.Stdout, "filvault %s\n", version)
		return nil
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
	case "logout":
		return logout(cmds, cfg)
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
			return usageError()
		}
		return cmds.Download(args[1])
	case "mkdir":
		return mkdir(cmds, args[1:])
	case "rm":
		if len(args) < 2 {
			return usageError()
		}
		return cmds.Rm(args[1])
	case "mv":
		return mv(cmds, args[1:])
	case "trash":
		return cmds.Trash()
	case "restore":
		if len(args) < 2 {
			return usageError()
		}
		return cmds.Restore(args[1])
	case "search":
		if len(args) < 2 {
			return usageError()
		}
		return cmds.Search(args[1])
	case "storage":
		return cmds.Storage()
	default:
		return usageError()
	}
}

func login(cmds *command.Commands, cfg config.Config) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprint(os.Stderr, "email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Fprint(os.Stderr, "password: ")
	password, err := readPassword()
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr)

	return cmds.Login(email, password, func(access, refresh string) error {
		cfg.AccessToken = access
		cfg.RefreshToken = refresh
		cfg.APIBase = cmds.Client.Base
		return config.Save(cfg)
	})
}

func logout(cmds *command.Commands, cfg config.Config) error {
	return cmds.Logout(cfg.RefreshToken, func() error {
		cfg.AccessToken = ""
		cfg.RefreshToken = ""
		return config.Save(cfg)
	})
}

// readPassword reads a password without echoing it to the terminal.
// Falls back to plain stdin when not attached to a TTY (e.g. piped input).
func readPassword() (string, error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		data, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func upload(cmds *command.Commands, args []string) error {
	fs := flag.NewFlagSet("upload", flag.ContinueOnError)
	folder := fs.String("folder", "", "target folder id")
	if err := fs.Parse(args); err != nil {
		return usageError()
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return usageError()
	}
	return cmds.Upload(rest[0], *folder)
}

func mkdir(cmds *command.Commands, args []string) error {
	fs := flag.NewFlagSet("mkdir", flag.ContinueOnError)
	parent := fs.String("parent", "", "parent folder id")
	if err := fs.Parse(args); err != nil {
		return usageError()
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return usageError()
	}
	return cmds.Mkdir(rest[0], *parent)
}

func mv(cmds *command.Commands, args []string) error {
	fs := flag.NewFlagSet("mv", flag.ContinueOnError)
	to := fs.String("to", "", "target folder id")
	if err := fs.Parse(args); err != nil {
		return usageError()
	}
	rest := fs.Args()
	if *to != "" {
		if len(rest) != 1 {
			return usageError()
		}
		return cmds.Mv(rest[0], "", *to)
	}
	if len(rest) != 2 {
		return usageError()
	}
	return cmds.Mv(rest[0], rest[1], "")
}

func usageError() error {
	return &exitError{code: 2, err: fmt.Errorf("usage:\n%s", usageText())}
}

func printUsage(w *os.File) {
	fmt.Fprintln(w, usageText())
}

func usageText() string {
	return `  filvault login
  filvault logout
  filvault whoami
  filvault ls [folderId]
  filvault upload <file> [--folder <id>]
  filvault download <name>
  filvault mkdir <name> [--parent <id>]
  filvault rm <name>
  filvault mv <name> <newName>
  filvault mv <name> --to <folderId>
  filvault trash
  filvault restore <name>
  filvault search <q>
  filvault storage
  filvault --help
  filvault --version`
}

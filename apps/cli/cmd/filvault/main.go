package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
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

	cmds := &command.Commands{
		Client:          c,
		Out:             os.Stdout,
		Err:             os.Stderr,
		CurrentFolderID: cfg.CurrentFolderID,
		SaveFolder: func(id string) error {
			cfg.CurrentFolderID = id
			return config.Save(cfg)
		},
	}

	switch args[0] {
	case "login":
		return login(cmds, cfg)
	case "logout":
		return logout(cmds, cfg)
	case "whoami":
		return cmds.Whoami()
	case "ls":
		if len(args) > 1 && hasGlob(args[1]) {
			return cmds.LsGlob(args[1])
		}
		folderID := ""
		if len(args) > 1 {
			folderID = args[1]
		}
		return cmds.Ls(folderID)
	case "cd":
		name := ""
		if len(args) > 1 {
			name = args[1]
		}
		return cmds.Cd(name)
	case "pwd":
		return cmds.Pwd()
	case "upload":
		return upload(cmds, args[1:])
	case "versions":
		if len(args) < 2 {
			return usageError()
		}
		return cmds.Versions(args[1])
	case "version-download":
		if len(args) < 3 {
			return usageError()
		}
		return cmds.VersionDownload(args[1], args[2])
	case "download":
		if len(args) < 2 {
			return usageError()
		}
		return cmds.Download(args[1])
	case "mkdir":
		return mkdir(cmds, args[1:])
	case "rm":
		if len(args) >= 2 && args[1] == "-f" {
			if len(args) < 3 {
				return usageError()
			}
			return cmds.Purge(args[2])
		}
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
	case "share":
		if len(args) < 3 {
			return usageError()
		}
		return cmds.Share(args[1], args[2])
	case "shares":
		return cmds.Shares()
	case "shared-with-me":
		return cmds.SharedWithMe()
	case "unshare":
		if len(args) < 2 {
			return usageError()
		}
		return cmds.Unshare(args[1])
	case "shared-ls":
		if len(args) < 2 {
			return usageError()
		}
		return cmds.SharedLs(args[1])
	case "shared-download":
		if len(args) < 3 {
			return usageError()
		}
		return cmds.SharedDownload(args[1], args[2])
	default:
		return usageError()
	}
}

func login(cmds *command.Commands, cfg config.Config) error {
	var email, password string
	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprint(os.Stderr, "email: ")
		e, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && e == "" {
			return err
		}
		email = e

		fmt.Fprint(os.Stderr, "password: ")
		data, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return err
		}
		password = string(data)
	} else {
		// Piped input: read both lines from one shared reader.
		var err error
		email, password, err = readCredentials(os.Stdin)
		if err != nil {
			return err
		}
	}
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	return cmds.Login(email, password, func(access, refresh string) error {
		cfg.AccessToken = access
		cfg.RefreshToken = refresh
		cfg.APIBase = cmds.Client.Base
		return config.Save(cfg)
	})
}

// readCredentials reads email and password from r using a single reader so
// piped input is not drained before the password line.
func readCredentials(r io.Reader) (string, string, error) {
	reader := bufio.NewReader(r)
	readLine := func() (string, error) {
		line, err := reader.ReadString('\n')
		if err != nil && line == "" {
			return "", err
		}
		return strings.TrimSpace(line), nil
	}
	email, err := readLine()
	if err != nil {
		return "", "", err
	}
	password, err := readLine()
	if err != nil && password == "" {
		return "", "", err
	}
	return email, password, nil
}

func logout(cmds *command.Commands, cfg config.Config) error {
	return cmds.Logout(cfg.RefreshToken, func() error {
		cfg.AccessToken = ""
		cfg.RefreshToken = ""
		return config.Save(cfg)
	})
}

func upload(cmds *command.Commands, args []string) error {
	fs := flag.NewFlagSet("upload", flag.ContinueOnError)
	folder := fs.String("folder", "", "target folder id")
	replace := fs.String("replace", "", "replace existing file by name")
	if err := fs.Parse(args); err != nil {
		return usageError()
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return usageError()
	}
	if *replace != "" {
		return cmds.UploadReplace(rest[0], *replace)
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

// hasGlob reports whether s contains glob metacharacters.
func hasGlob(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

func printUsage(w *os.File) {
	fmt.Fprintln(w, usageText())
}

func usageText() string {
	return `  filvault login
  filvault logout
  filvault whoami
  filvault ls [folderId|glob]
  filvault cd [name|..]
  filvault pwd
  filvault upload <file> [--folder <id>]
  filvault upload --replace <name> <file>
  filvault download <name>
  filvault versions <name>
  filvault version-download <name> <versionId>
  filvault mkdir <name> [--parent <id>]
  filvault rm <name>
  filvault rm -f <name>
  filvault mv <name> <newName>
  filvault mv <name> --to <folderId>
  filvault trash
  filvault restore <name>
  filvault search <q>
  filvault storage
  filvault share <name> <email>
  filvault shares
  filvault shared-with-me
  filvault unshare <id>
  filvault shared-ls <folderId>
  filvault shared-download <fileId> <name>
  filvault --help
  filvault --version`
}

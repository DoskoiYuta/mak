package mkfile

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

// RunOptions configures how `make` is invoked.
type RunOptions struct {
	// Makefile is the path to the Makefile (passed via -f). May be empty
	// to let make use its default discovery.
	Makefile string
	// Directory is the working directory (passed via -C). May be empty.
	Directory string
	// Target is the target to build.
	Target string
	// Variables is the ordered list of VAR=value assignments to pass on
	// the command line.
	Variables []VarAssignment
}

// VarAssignment is a single NAME=VALUE pair.
type VarAssignment struct {
	Name  string
	Value string
}

// BuildArgs returns the argv for `make` given the supplied options. The
// first element is the subcommand name ("make") to make it easy to pass to
// exec.Command.
func BuildArgs(opts RunOptions) []string {
	args := []string{"make"}
	if opts.Makefile != "" {
		args = append(args, "-f", opts.Makefile)
	}
	if opts.Directory != "" && opts.Directory != "." {
		args = append(args, "-C", opts.Directory)
	}
	if opts.Target != "" {
		args = append(args, opts.Target)
	}
	for _, v := range opts.Variables {
		args = append(args, v.Name+"="+v.Value)
	}
	return args
}

// PreviewCommand returns a human-readable version of the command that
// would be executed, shell-quoted where needed.
func PreviewCommand(opts RunOptions) string {
	args := BuildArgs(opts)
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = shellQuote(a)
	}
	return "$ " + strings.Join(parts, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	needsQuote := false
	for _, r := range s {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') &&
			r != '_' && r != '-' && r != '.' && r != '/' && r != '=' && r != ':' && r != ',' && r != '@' && r != '+' {
			needsQuote = true
			break
		}
	}
	if !needsQuote {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// Run executes `make` with the supplied options, streaming stdio to the
// current process. It returns the exit code from make.
func Run(opts RunOptions) (int, error) {
	args := BuildArgs(opts)
	if len(args) < 1 {
		return 1, errors.New("no command")
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}
	return 0, nil
}

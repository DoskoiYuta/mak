// Package cmd implements the mak CLI.
package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/doskoiyuta/mak/mkfile"
	"github.com/doskoiyuta/mak/tui"
)

// Version is overridden at build time via -ldflags "-X ...".
var Version = "dev"

type options struct {
	makefile  string
	directory string
	dryRun    bool
}

// Execute is the entrypoint used by main.
func Execute() error {
	var opts options

	rootCmd := &cobra.Command{
		Use:           "mak [Makefileパス]",
		Short:         "Fuzzy-find Makefile targets and run them interactively",
		Version:       Version,
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Positional arg overrides -f if both given and -f is still the
			// default; otherwise -f wins (explicit flag).
			if len(args) == 1 && !cmd.Flags().Changed("file") {
				opts.makefile = args[0]
			}
			return run(opts)
		},
	}

	rootCmd.Flags().StringVarP(&opts.makefile, "file", "f", "./Makefile", "Makefileパスを指定")
	rootCmd.Flags().StringVarP(&opts.directory, "directory", "C", ".", "実行ディレクトリを指定")
	rootCmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "コマンドを表示するだけで実行しない")

	return rootCmd.Execute()
}

func run(opts options) error {
	path, err := resolveMakefile(opts.directory, opts.makefile)
	if err != nil {
		return err
	}

	parsed, err := mkfile.Parse(path)
	if err != nil {
		return err
	}
	if len(parsed.Targets) == 0 {
		return errors.New("ターゲットが見つかりません")
	}

	result, err := tui.Run(parsed)
	if err != nil {
		return err
	}

	switch result.Action {
	case tui.ExitCancel:
		return nil
	case tui.ExitCopy:
		cmdStr := mkfile.PreviewCommand(mkfile.RunOptions{
			Makefile:  opts.makefile,
			Directory: opts.directory,
			Target:    result.Target,
			Variables: result.Variables,
		})
		// Strip the leading "$ " used by the preview.
		if len(cmdStr) > 2 && cmdStr[:2] == "$ " {
			cmdStr = cmdStr[2:]
		}
		if err := clipboard.WriteAll(cmdStr); err != nil {
			return fmt.Errorf("クリップボードへのコピーに失敗: %w", err)
		}
		fmt.Fprintln(os.Stderr, "コピー済み: "+cmdStr)
		return nil
	case tui.ExitRun:
		runOpts := mkfile.RunOptions{
			Makefile:  opts.makefile,
			Directory: opts.directory,
			Target:    result.Target,
			Variables: result.Variables,
		}
		if opts.dryRun {
			fmt.Println(mkfile.PreviewCommand(runOpts))
			return nil
		}
		code, err := mkfile.Run(runOpts)
		if err != nil {
			return err
		}
		if code != 0 {
			os.Exit(code)
		}
		return nil
	}
	return nil
}

// resolveMakefile resolves the Makefile path relative to the -C directory
// when needed, and validates that it exists.
func resolveMakefile(dir, mf string) (string, error) {
	candidate := mf
	if !filepath.IsAbs(mf) && dir != "" && dir != "." {
		candidate = filepath.Join(dir, mf)
	}
	info, err := os.Stat(candidate)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("Makefileが見つかりません: %s", candidate)
		}
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("Makefileパスがディレクトリです: %s", candidate)
	}
	return candidate, nil
}

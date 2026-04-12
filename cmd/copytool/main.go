package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"copytool/internal/cli"
	"copytool/internal/core"
	"copytool/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	root := flag.String("root", ".", "project root")
	output := flag.String("output", "", "output target: empty = <root>/tools/output.txt, '.' = <cwd>/tools/output.txt, file path = write there, directory path = <dir>/tools/output.txt")
	tuiMode := flag.Bool("tui", false, "force terminal UI mode (default unless -cli is supplied)")
	cliMode := flag.Bool("cli", false, "run direct export without opening the TUI")
	clipboard := flag.Bool("clipboard", false, "copy exported output to clipboard")
	includeExt := flag.String("include-ext", "", "comma-separated file extensions to include")
	excludeExt := flag.String("exclude-ext", "", "comma-separated file extensions to exclude")
	excludeDirs := flag.String("exclude-dirs", "", "comma-separated directory names to exclude")
	versionFlag := flag.Bool("version", false, "print version/build information and exit")

	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = func() {
		prog := filepath.Base(os.Args[0])

		fmt.Fprintf(flag.CommandLine.Output(), "%s\n\n", prog)
		fmt.Fprintf(flag.CommandLine.Output(), "A gitignore-aware code/context export tool with both TUI and direct CLI modes.\n\n")

		fmt.Fprintf(flag.CommandLine.Output(), "USAGE\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s [flags] [paths...]\n\n", prog)

		fmt.Fprintf(flag.CommandLine.Output(), "MODES\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  Default / TUI mode\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    Opens the interactive terminal UI.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    Example:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "      %s -root \"D:\\codes\\myproject\"\n\n", prog)

		fmt.Fprintf(flag.CommandLine.Output(), "  CLI export mode (-cli)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    Runs direct export without opening the TUI.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    If no paths are passed, it reads from tools/list.txt.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "    Example:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "      %s -cli -root \"D:\\codes\\myproject\" src README.md\n\n", prog)

		fmt.Fprintf(flag.CommandLine.Output(), "FLAGS\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\n")

		fmt.Fprintf(flag.CommandLine.Output(), "PATH ARGUMENTS\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  Paths are interpreted relative to -root unless they are absolute.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  In TUI mode, passed paths become startup selections when possible.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  In CLI mode, passed paths are exported directly.\n\n")

		fmt.Fprintf(flag.CommandLine.Output(), "OUTPUT TARGET\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  no -output       => <root>/tools/output.txt\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -output .        => <cwd>/tools/output.txt\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -output <file>   => write directly to that file\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -output <dir>    => <dir>/tools/output.txt\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  tools/list.txt always remains under the root project.\n\n")

		fmt.Fprintf(flag.CommandLine.Output(), "FILTERS\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -include-ext    Whitelist extensions, e.g. \".go,.md,.py\"\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -exclude-ext    Hide selected file types, e.g. \".png,.jpg,.exe\"\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -exclude-dirs   Hide extra directory names, e.g. \"docs,dist,tests\"\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  Root and nested .gitignore files are respected automatically.\n\n")

		fmt.Fprintf(flag.CommandLine.Output(), "OUTPUT BEHAVIOR\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  Exports create/use the tools/ directory under the root project for list.txt.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  tools/list.txt   exact selected file list\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  output target    merged markdown-fenced output\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  If a root .gitignore exists, tools/ is added when needed.\n\n")

		fmt.Fprintf(flag.CommandLine.Output(), "EXAMPLES\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s\n", prog)
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -root \"D:\\codes\\myproject\"\n", prog)
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -cli -root \"D:\\codes\\myproject\" src README.md\n", prog)
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -cli -output . .\n", prog)
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -cli -output \"C:\\temp\\merged.md\" src\n", prog)
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -cli -include-ext \".py,.md\" -exclude-dirs \"docs,tests\" .\n", prog)
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -version\n", prog)
	}

	flag.Parse()

	if *versionFlag {
		fmt.Printf("copytool %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built:  %s\n", buildDate)
		return
	}

	filters := core.DefaultFilterConfig()

	if strings.TrimSpace(*includeExt) != "" {
		filters.IncludeExts = core.ParseExtensionCSV(*includeExt)
	}

	if strings.TrimSpace(*excludeExt) != "" {
		filters.ExcludeExts = core.ParseExtensionCSV(*excludeExt)
	}

	if strings.TrimSpace(*excludeDirs) != "" {
		filters.ExcludeDirs = core.ParseNameCSV(*excludeDirs)
	}

	useTUI := true
	if *cliMode {
		useTUI = false
	}
	if *tuiMode {
		useTUI = true
	}

	if !useTUI {
		if err := cli.RunExport(cli.Options{
			RootPath:    *root,
			Paths:       flag.Args(),
			Filters:     filters,
			OutputSpec:  *output,
			ToClipboard: *clipboard,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "copytool CLI export failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	model, err := tui.NewModel(*root, flag.Args(), *clipboard, filters, *output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize TUI: %v\n", err)
		os.Exit(1)
	}

	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
	)

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "copytool crashed: %v\n", err)
		os.Exit(1)
	}
}
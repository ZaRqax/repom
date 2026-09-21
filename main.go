package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ZaRqax/repom/internal/repository/finder"
	"github.com/ZaRqax/repom/internal/repository/git"
	"github.com/ZaRqax/repom/internal/service/repomanager"
	"github.com/ZaRqax/repom/internal/transport/tui"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			fmt.Println("repom", version)
			return
		}
	}

	workDir := "."
	if len(os.Args) > 1 {
		workDir = os.Args[1]
	}

	workDir, err := filepath.Abs(workDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	gitClient := git.NewClient()
	repoFinder := finder.New(gitClient)
	service := repomanager.New(gitClient, repoFinder)

	if err := tui.Run(service, workDir); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Delta456/box-cli-maker/v2"
	"github.com/briandowns/spinner"
	"github.com/joho/godotenv"
	"github.com/rm-hull/git-commit-summary/internal"
	"github.com/rm-hull/git-commit-summary/internal/git"
	"github.com/ttacon/chalk"
	"google.golang.org/genai"
)

//go:embed prompt.md
var prompt string

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	ctx := context.Background()
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = chalk.Magenta.Color(" Running git diff")
	s.Start()
	defer s.Stop()

	out, err := git.Diff()
	if err != nil {
		log.Fatal(err)
	}

	if len(out) == 0 {
		s.FinalMSG = chalk.Red.Color("No changes are staged\n")
		os.Exit(1)
	}

	s.Suffix = chalk.Blue.Color(" Generating commit summary")
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	text := fmt.Sprintf(prompt, out)

	result, err := client.Models.GenerateContent(
		ctx,
		// "gemini-2.5-pro",
		// "gemini-2.5-flash",
		"gemini-2.5-flash-preview-09-2025",
		// "gemini-2.5-flash-lite-preview-09-2025",
		genai.Text(text),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	s.Stop()

	message := result.Text()

	Box := box.New(box.Config{Px: 1, Py: 0, Type: "Round", Color: "Cyan", TitlePos: "Top"})
	Box.Print("Commit message", message)
	confirm, err := internal.Ask(chalk.Yellow.Color("Confirm commit?"))
	if err != nil {
		log.Fatal(err)
	}
	if confirm {
		if err := git.Commit(message); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println(chalk.Red.Color("ABORTED!"))
	}
}

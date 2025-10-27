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
	"github.com/gookit/color"
	"github.com/joho/godotenv"
	"github.com/rm-hull/git-commit-summary/internal"
	"github.com/rm-hull/git-commit-summary/internal/git"
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
	s.Suffix = color.Magenta.Render(" Running git diff")
	s.Start()
	defer s.Stop()

	out, err := git.Diff()
	if err != nil {
		log.Fatal(err)
	}

	if len(out) == 0 {
		s.FinalMSG = color.Red.Render("No changes are staged")
		s.Stop()
		os.Exit(1)
	}

	s.Suffix = color.Blue.Render(" Generating commit summary")
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
	confirm, err := internal.Ask(color.Yellow.Render("Confirm commit?"))
	if err != nil {
		log.Fatal(err)
	}
	if confirm {
		if err := git.Commit(message); err != nil {
			log.Fatal(err)
		}
	} else {
		color.Red.Println("ABORTED!")
	}
}

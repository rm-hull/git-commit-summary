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
	s.Suffix = color.Render(" <magenta>Running git diff</>")
	s.Start()

	out, err := git.Diff()
	if err != nil {
		log.Fatal(err)
	}

	if len(out) == 0 {
		s.FinalMSG = color.Render("<fg=red;op=bold>No changes are staged</>")
		s.Stop()
		os.Exit(1)
	}

	model := getModel()
	s.Suffix = color.Render(fmt.Sprintf(" <blue>Generating commit summary (using: </><fg=blue;op=bold>%s</><blue>)</>", model))
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	text := fmt.Sprintf(prompt, out)

	result, err := client.Models.GenerateContent(
		ctx,
		model,
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
	confirm, err := internal.Ask(color.Render("<yellow>Confirm commit?</>"))
	if err != nil {
		log.Fatal(err)
	}
	if confirm {
		if err := git.Commit(message); err != nil {
			log.Fatal(err)
		}
	} else {
		color.Println("<fg=red;op=bold>ABORTED!</>")
	}
}

func getModel() string {
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-flash-preview-09-2025"
	}
	return model
}

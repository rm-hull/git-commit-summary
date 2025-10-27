package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/joho/godotenv"
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

	args := strings.Split("--no-pager diff --no-ext-diff --no-textconv --staged --diff-filter=ACMRTUXB", " ")
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	s.Suffix = " Generating git summary"
	s.Start()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	text := fmt.Sprintf(prompt, out)

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash-lite-preview-09-2025",
		genai.Text(text),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	s.Stop()
	fmt.Println(result.Text())
}

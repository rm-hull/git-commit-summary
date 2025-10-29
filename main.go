package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Delta456/box-cli-maker/v2"
	"github.com/adrg/xdg"
	"github.com/briandowns/spinner"
	"github.com/earthboundkid/versioninfo/v2"
	"github.com/gookit/color"
	"github.com/joho/godotenv"
	"github.com/rm-hull/git-commit-summary/internal"
	"github.com/rm-hull/git-commit-summary/internal/git"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
	"github.com/spf13/cobra"
)

//go:embed prompt.md
var prompt string

var userMessage string

var rootCmd = &cobra.Command{
	Use:   "git-commit-summary",
	Short: "Generate a commit summary using Gemini",
	Run:   run,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&userMessage, "message", "m", "", "Append a message to the commit summary")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Display version information")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(cmd *cobra.Command, args []string) {
	version, _ := cmd.Flags().GetBool("version")
	if version {
		fmt.Println(versioninfo.Short())
		os.Exit(0)
	}

	configFile, err := xdg.ConfigFile("git-commit-summary/config.env")
	if err != nil {
		log.Fatal(err)
	}

	_ = godotenv.Load(configFile)
	_ = godotenv.Overload(".env")

	ctx := context.Background()
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = color.Render(" <magenta>Running git diff</>")
	s.Start()

	out, err := git.Diff()
	if err != nil {
		s.Stop()
		log.Fatal(err)
	}

	if len(out) == 0 {
		s.FinalMSG = color.Render("<fg=red;op=bold>No changes are staged</>")
		s.Stop()
		os.Exit(1)
	}

	provider, err := llmprovider.NewProvider(ctx)
	if err != nil {
		s.Stop()
		log.Fatal(err)
	}

	s.Suffix = color.Render(fmt.Sprintf(" <blue>Generating commit summary (using: </><fg=blue;op=bold>%s</><blue>)</>", provider.Model()))
	text := fmt.Sprintf(prompt, out)

	message, err := provider.Call(ctx, "", text)
	if err != nil {
		s.Stop()
		log.Fatal(err)
	}

	s.Stop()

	if userMessage != "" {
		message = fmt.Sprintf("%s\n\n%s", userMessage, message)
	}

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

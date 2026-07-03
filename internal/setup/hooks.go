package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rm-hull/git-commit-summary/internal/ui"
)

func InstallHook() error {
	gitDir, err := exec.Command("git", "rev-parse", "--git-dir").Output()
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	gitDirStr := strings.TrimSpace(string(gitDir))
	hookPath := filepath.Join(gitDirStr, "hooks", "prepare-commit-msg")

	if _, err := os.Stat(hookPath); err == nil {
		existingContent, err := os.ReadFile(hookPath)
		if err == nil && !strings.Contains(string(existingContent), "git-commit-summary") {
			return fmt.Errorf("a prepare-commit-msg hook already exists at %s; please back up and remove it before installing", hookPath)
		}
	}
	absExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	content := fmt.Sprintf("#!/bin/sh\n\"%s\" --yolo \"$1\"\n", absExe)

	err = os.MkdirAll(filepath.Dir(hookPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	err = os.WriteFile(hookPath, []byte(content), 0755)
	if err != nil {
		return fmt.Errorf("failed to write hook file: %w", err)
	}

	fmt.Println(ui.Green.Bold(true).Render("Git hook installed successfully!"))
	return nil
}

func UninstallHook() error {
	gitDir, err := exec.Command("git", "rev-parse", "--git-dir").Output()
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	gitDirStr := strings.TrimSpace(string(gitDir))
	hookPath := filepath.Join(gitDirStr, "hooks", "prepare-commit-msg")
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		fmt.Println(ui.BoldYellow.Render("Hook not found, nothing to uninstall."))
		return nil
	}

	existingContent, err := os.ReadFile(hookPath)
	if err != nil {
		return fmt.Errorf("failed to read hook file: %w", err)
	}
	if !strings.Contains(string(existingContent), "git-commit-summary") {
		fmt.Println(ui.BoldYellow.Render("Hook was not installed by git-commit-summary, skipping uninstall."))
		return nil
	}

	err = os.Remove(hookPath)
	if err != nil {
		return fmt.Errorf("failed to remove hook file: %w", err)
	}

	fmt.Println(ui.Green.Bold(true).Render("Git hook uninstalled successfully!"))
	return nil
}

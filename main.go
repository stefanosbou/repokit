package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type RepoResult struct {
	Name   string
	Status string
	Err    error
}

func main() {
	root := "."

	entries, err := os.ReadDir(root)
	if err != nil {
		fmt.Printf("Failed to read directory: %v\n", err)
		os.Exit(1)
	}

	var results []RepoResult

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repoPath := filepath.Join(root, entry.Name())

		if !isGitRepo(repoPath) {
			continue
		}

		result := updateRepo(repoPath)
		results = append(results, result)
	}

	fmt.Println("\n========== SUMMARY ==========")

	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("❌ %s -> %s (%v)\n", r.Name, r.Status, r.Err)
		} else {
			fmt.Printf("✅ %s -> %s\n", r.Name, r.Status)
		}
	}
}

func updateRepo(repoPath string) RepoResult {
	fmt.Printf("\n=== Processing %s ===\n", repoPath)

	currentBranch, err := getCurrentBranch(repoPath)
	if err != nil {
		return RepoResult{
			Name:   repoPath,
			Status: "failed to detect branch",
			Err:    err,
		}
	}

	stashed := false

	if isDirty(repoPath) {
		fmt.Println("Working tree dirty -> stashing changes")

		if err := stashChanges(repoPath); err != nil {
			return RepoResult{
				Name:   repoPath,
				Status: "failed to stash changes",
				Err:    err,
			}
		}

		stashed = true
	}

	if err := run(repoPath, "git", "fetch", "--all", "--prune"); err != nil {
		restoreAfterFailure(repoPath, currentBranch, stashed)

		return RepoResult{
			Name:   repoPath,
			Status: "fetch failed",
			Err:    err,
		}
	}

	mainBranch := detectMainBranch(repoPath)

	if mainBranch != "" {
		fmt.Printf("Updating %s...\n", mainBranch)

		if err := run(repoPath, "git", "checkout", mainBranch); err != nil {
			restoreAfterFailure(repoPath, currentBranch, stashed)

			return RepoResult{
				Name:   repoPath,
				Status: "failed checkout to main branch",
				Err:    err,
			}
		}

		if err := run(repoPath, "git", "pull", "--ff-only"); err != nil {
			restoreAfterFailure(repoPath, currentBranch, stashed)

			return RepoResult{
				Name:   repoPath,
				Status: "pull failed",
				Err:    err,
			}
		}
	}

	if err := restoreBranch(repoPath, currentBranch); err != nil {
		return RepoResult{
			Name:   repoPath,
			Status: "updated but failed restoring branch",
			Err:    err,
		}
	}

	if stashed {
		fmt.Println("Restoring stashed changes")

		if err := popStash(repoPath); err != nil {
			return RepoResult{
				Name:   repoPath,
				Status: "updated but failed restoring stash",
				Err:    err,
			}
		}
	}

	return RepoResult{
		Name:   repoPath,
		Status: "updated successfully",
	}
}

func isGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

func isDirty(repoPath string) bool {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")

	output, err := cmd.Output()
	if err != nil {
		return true
	}

	return len(bytes.TrimSpace(output)) > 0
}

func stashChanges(repoPath string) error {
	stashMessage := fmt.Sprintf(
		"auto-stash-%d",
		time.Now().Unix(),
	)

	return run(
		repoPath,
		"git",
		"stash",
		"push",
		"-u",
		"-m",
		stashMessage,
	)
}

func popStash(repoPath string) error {
	// Apply latest stash without deleting it first
	if err := run(repoPath, "git", "stash", "apply"); err != nil {
		return err
	}

	// Only drop stash if apply succeeded
	return run(repoPath, "git", "stash", "drop")
}

func getCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command(
		"git",
		"-C",
		repoPath,
		"branch",
		"--show-current",
	)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(output)), nil
}

func detectMainBranch(repoPath string) string {
	branches := []string{"main", "master"}

	for _, branch := range branches {
		cmd := exec.Command(
			"git",
			"-C",
			repoPath,
			"show-ref",
			"--verify",
			"--quiet",
			"refs/heads/"+branch,
		)

		if cmd.Run() == nil {
			return branch
		}
	}

	return ""
}

func restoreBranch(repoPath, branch string) error {
	if branch == "" {
		return nil
	}

	fmt.Printf("Restoring branch %s...\n", branch)

	return run(repoPath, "git", "checkout", branch)
}

func restoreAfterFailure(repoPath, branch string, stashed bool) {
	_ = restoreBranch(repoPath, branch)

	if stashed {
		_ = popStash(repoPath)
	}
}

func run(repoPath string, name string, args ...string) error {
	cmd := exec.Command(name, args...)

	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

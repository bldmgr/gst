package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitSearchTool struct {
	repoPath string
}

func NewGitSearchTool(path string) *GitSearchTool {
	return &GitSearchTool{
		repoPath: path,
	}
}

// isGitRepo checks if the current directory is a git repository
func (g *GitSearchTool) isGitRepo() bool {
	gitDir := filepath.Join(g.repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return false
	}
	return true
}

// getLastCommitMessage retrieves the last commit message
func (g *GitSearchTool) getLastCommitMessage() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=format:%s")
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit message: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getLastCommitDetails retrieves detailed information about the last commit
func (g *GitSearchTool) getLastCommitDetails() (map[string]string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=format:%H|%an|%ae|%ad|%s|%b", "--date=short")
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit details: %v", err)
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) < 5 {
		return nil, fmt.Errorf("unexpected git log output format")
	}

	details := map[string]string{
		"hash":    parts[0],
		"author":  parts[1],
		"email":   parts[2],
		"date":    parts[3],
		"subject": parts[4],
		"body":    "",
	}

	if len(parts) > 5 {
		details["body"] = parts[5]
	}

	return details, nil
}

// searchInCommitHistory searches for a query in commit messages
func (g *GitSearchTool) searchInCommitHistory(query string, maxResults int) ([]map[string]string, error) {
	cmd := exec.Command("git", "log", "--grep="+query, "-i",
		fmt.Sprintf("-%d", maxResults),
		"--pretty=format:%H|%an|%ad|%s", "--date=short")
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to search commit history: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var results []map[string]string

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			result := map[string]string{
				"hash":    parts[0],
				"author":  parts[1],
				"date":    parts[2],
				"subject": parts[3],
			}
			results = append(results, result)
		}
	}

	return results, nil
}

// searchInFiles searches for a query in tracked files
func (g *GitSearchTool) searchInFiles(query string, maxResults int) ([]string, error) {
	cmd := exec.Command("git", "grep", "-n", "-i", "--", query)
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		// git grep returns non-zero exit code when no matches found
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to search in files: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	// Limit results
	if len(lines) > maxResults {
		lines = lines[:maxResults]
	}

	return lines, nil
}

func (g *GitSearchTool) displayLastCommit() {
	fmt.Println("=== Last Commit Information ===")

	details, err := g.getLastCommitDetails()
	if err != nil {
		log.Printf("Error getting commit details: %v", err)
		return
	}

	fmt.Printf("Hash:    %s\n", details["hash"][:8])
	fmt.Printf("Author:  %s <%s>\n", details["author"], details["email"])
	fmt.Printf("Date:    %s\n", details["date"])
	fmt.Printf("Subject: %s\n", details["subject"])

	if details["body"] != "" {
		fmt.Printf("Body:    %s\n", details["body"])
	}

	fmt.Println()
}

func (g *GitSearchTool) interactiveSearch() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Enter search query (or 'quit' to exit): ")
		if !scanner.Scan() {
			break
		}

		query := strings.TrimSpace(scanner.Text())
		if query == "quit" || query == "exit" || query == "q" {
			break
		}

		if query == "" {
			continue
		}

		g.performSearch(query)
	}
}

func (g *GitSearchTool) performSearch(query string) {
	fmt.Printf("\n=== Search Results for: \"%s\" ===\n", query)

	hash := "64fc5dd7"
	// Search in last commit messages
	fmt.Println("\n--- Commit Messages ---")
	commits, err := g.searchInCommitHistory(query, 1)
	if err != nil {
		log.Printf("Error searching commits: %v", err)
	} else if len(commits) == 0 {
		fmt.Println("No matches found in commit messages.")
	} else {
		for i, commit := range commits {
			if commit["hash"][:8] == hash {
				fmt.Printf("%d. [%s] %s - %s (%s)\n",
					i+1, commit["hash"][:8], commit["subject"],
					commit["author"], commit["date"])
			}
		}
	}

	// Search in commit messages
	fmt.Println("\n--- Commit Messages ---")
	commits, err = g.searchInCommitHistory(query, 10)
	if err != nil {
		log.Printf("Error searching commits: %v", err)
	} else if len(commits) == 0 {
		fmt.Println("No matches found in commit messages.")
	} else {
		for i, commit := range commits {
			fmt.Printf("%d. [%s] %s - %s (%s)\n",
				i+1, commit["hash"][:8], commit["subject"],
				commit["author"], commit["date"])
		}
	}

	// Search in files
	fmt.Println("\n--- File Contents ---")
	fileMatches, err := g.searchInFiles(query, 20)
	if err != nil {
		log.Printf("Error searching files: %v", err)
	} else if len(fileMatches) == 0 {
		fmt.Println("No matches found in tracked files.")
	} else {
		for i, match := range fileMatches {
			fmt.Printf("%d. %s\n", i+1, match)
		}
		if len(fileMatches) == 20 {
			fmt.Println("... (showing first 20 matches)")
		}
	}

	fmt.Println()
}

func main() {
	var (
		repoPath = flag.String("path", ".", "Path to git repository")
		query    = flag.String("query", "", "Search query (if empty, enters interactive mode)")
		showHelp = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	if *showHelp {
		fmt.Println("Git Commit Search Tool")
		fmt.Println("Usage:")
		fmt.Println("  -path string    Path to git repository (default: current directory)")
		fmt.Println("  -query string   Search query (if empty, enters interactive mode)")
		fmt.Println("  -help           Show this help message")
		fmt.Println("\nExamples:")
		fmt.Println("  ./git-search                          # Interactive mode in current directory")
		fmt.Println("  ./git-search -query \"bug fix\"         # Search for 'bug fix'")
		fmt.Println("  ./git-search -path /path/to/repo      # Use different repository")
		return
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(*repoPath)
	if err != nil {
		log.Fatalf("Error resolving path: %v", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", absPath)
	}

	tool := NewGitSearchTool(absPath)

	// Check if it's a git repository
	if !tool.isGitRepo() {
		log.Fatalf("Not a git repository: %s", absPath)
	}

	fmt.Printf("Git repository: %s\n", absPath)

	// Display last commit information
	tool.displayLastCommit()

	// Handle search
	if *query != "" {
		// Single query mode
		tool.performSearch(*query)
	} else {
		// Interactive mode
		fmt.Println("=== Interactive Search Mode ===")
		fmt.Println("You can search for text in commit messages and file contents.")
		tool.interactiveSearch()
	}

	fmt.Println("Goodbye!")
}

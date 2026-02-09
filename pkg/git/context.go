// Package git provides git repository detection and context
package git

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
)

// Context holds git repository information
type Context struct {
	// RepoPath is the root path of the repository
	RepoPath string
	
	// RootPath is alias for RepoPath (for compatibility)
	RootPath string
	
	// Branch is the current branch name
	Branch string
	
	// RemoteURL is the origin remote URL
	RemoteURL string
	
	// TaskCode extracted from branch (e.g., GS-123)
	TaskCode string
	
	// Provider is github/gitlab/bitbucket
	Provider string
	
	// RepoFullName is owner/repo format
	RepoFullName string
}

// TaskCodePattern matches task codes like ABC-123
// Each GitScrum project has its own prefix (e.g., GS, WEB, API)
var TaskCodePattern = regexp.MustCompile(`([A-Z]{2,5})-(\d+)`)

// ResolveContext detects git context from current directory
func ResolveContext(path string) (*Context, error) {
	// Find .git directory
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, err
	}

	ctx := &Context{}

	// Get worktree path
	worktree, err := repo.Worktree()
	if err == nil {
		ctx.RepoPath = worktree.Filesystem.Root()
		ctx.RootPath = ctx.RepoPath // Alias for compatibility
	}

	// Get current branch
	head, err := repo.Head()
	if err == nil {
		ctx.Branch = head.Name().Short()
		ctx.TaskCode = extractTaskCode(ctx.Branch)
	}

	// Get remote URL
	remote, err := repo.Remote("origin")
	if err == nil && len(remote.Config().URLs) > 0 {
		ctx.RemoteURL = remote.Config().URLs[0]
		ctx.Provider, ctx.RepoFullName = parseRemoteURL(ctx.RemoteURL)
	}

	return ctx, nil
}

// NewContext is an alias for ResolveContext (for compatibility)
func NewContext(path string) (*Context, error) {
	return ResolveContext(path)
}

// extractTaskCode extracts task code from branch name
// The prefix is project-specific and configured in GitScrum settings
// Examples: feature/GS-123-fix-bug -> GS-123
//           bugfix/WEB-456-auth    -> WEB-456
func extractTaskCode(branch string) string {
	matches := TaskCodePattern.FindStringSubmatch(branch)
	if len(matches) >= 3 {
		return matches[1] + "-" + matches[2]
	}
	return ""
}

// parseRemoteURL extracts provider and repo name from remote URL
func parseRemoteURL(url string) (provider, repoFullName string) {
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH URLs: git@github.com:owner/repo
	if strings.HasPrefix(url, "git@") {
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 2 {
			host := strings.TrimPrefix(parts[0], "git@")
			provider = detectProvider(host)
			repoFullName = parts[1]
			return
		}
	}

	// Handle HTTPS URLs: https://github.com/owner/repo
	if strings.Contains(url, "://") {
		parts := strings.SplitN(url, "://", 2)
		if len(parts) == 2 {
			pathParts := strings.SplitN(parts[1], "/", 2)
			if len(pathParts) == 2 {
				provider = detectProvider(pathParts[0])
				repoFullName = pathParts[1]
				return
			}
		}
	}

	return "", ""
}

// detectProvider determines git provider from hostname
func detectProvider(host string) string {
	host = strings.ToLower(host)
	switch {
	case strings.Contains(host, "github"):
		return "github"
	case strings.Contains(host, "gitlab"):
		return "gitlab"
	case strings.Contains(host, "bitbucket"):
		return "bitbucket"
	default:
		return "unknown"
	}
}

// IsGitRepo checks if the path is inside a git repository
func IsGitRepo(path string) bool {
	_, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	return err == nil
}

// FindRepoRoot finds the root of the git repository
func FindRepoRoot(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	for {
		gitPath := filepath.Join(absPath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return absPath, nil
		}

		parent := filepath.Dir(absPath)
		if parent == absPath {
			return "", git.ErrRepositoryNotExists
		}
		absPath = parent
	}
}

package git

import (
	"testing"
)

func TestExtractTaskCode(t *testing.T) {
	tests := []struct {
		name   string
		branch string
		want   string
	}{
		// Standard patterns
		{"feature branch with GS prefix", "feature/GS-123-fix-auth-bug", "GS-123"},
		{"feature branch with TK prefix", "feature/TK-456-add-feature", "TK-456"},
		{"bugfix branch", "bugfix/BUG-789-handle-error", "BUG-789"},
		{"hotfix branch", "hotfix/FIX-101-urgent-fix", "FIX-101"},

		// Edge cases
		{"code at start", "GS-123-fix-bug", "GS-123"},
		{"code at end", "fix-bug-GS-123", "GS-123"},
		{"code with long prefix", "PROJ-12345-long-description", "PROJ-12345"},
		{"multiple codes takes first", "GS-123-merge-TK-456", "GS-123"},

		// No match cases
		{"main branch", "main", ""},
		{"develop branch", "develop", ""},
		{"master branch", "master", ""},
		{"release branch without code", "release/v1.0.0", ""},
		{"lowercase prefix", "gs-123-fix", ""},
		{"single letter prefix", "G-123-test", ""},
		{"no hyphen number", "GS123-test", ""},
		{"empty string", "", ""},

		// Two-letter prefix (minimum)
		{"two letter prefix", "AB-1-test", "AB-1"},

		// Five-letter prefix (maximum in pattern)
		{"five letter prefix", "ABCDE-999-test", "ABCDE-999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTaskCode(tt.branch)
			if got != tt.want {
				t.Errorf("extractTaskCode(%q) = %q, want %q", tt.branch, got, tt.want)
			}
		})
	}
}

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantProvider string
		wantRepo     string
	}{
		// SSH URLs
		{"github SSH", "git@github.com:owner/repo.git", "github", "owner/repo"},
		{"gitlab SSH", "git@gitlab.com:owner/repo.git", "gitlab", "owner/repo"},
		{"bitbucket SSH", "git@bitbucket.org:owner/repo.git", "bitbucket", "owner/repo"},
		{"SSH without .git", "git@github.com:owner/repo", "github", "owner/repo"},

		// HTTPS URLs
		{"github HTTPS", "https://github.com/owner/repo.git", "github", "owner/repo"},
		{"gitlab HTTPS", "https://gitlab.com/owner/repo.git", "gitlab", "owner/repo"},
		{"bitbucket HTTPS", "https://bitbucket.org/owner/repo.git", "bitbucket", "owner/repo"},
		{"HTTPS without .git", "https://github.com/owner/repo", "github", "owner/repo"},

		// Nested repos
		{"github nested", "git@github.com:org/team/repo.git", "github", "org/team/repo"},
		{"gitlab nested", "https://gitlab.com/org/team/repo.git", "gitlab", "org/team/repo"},

		// Self-hosted
		{"self-hosted github", "git@git.company.com:team/repo.git", "unknown", "team/repo"},
		{"self-hosted gitlab", "https://gitlab.company.com/team/repo", "gitlab", "team/repo"},

		// Edge cases
		{"empty string", "", "", ""},
		{"invalid URL", "not-a-url", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProvider, gotRepo := parseRemoteURL(tt.url)
			if gotProvider != tt.wantProvider {
				t.Errorf("parseRemoteURL(%q) provider = %q, want %q", tt.url, gotProvider, tt.wantProvider)
			}
			if gotRepo != tt.wantRepo {
				t.Errorf("parseRemoteURL(%q) repo = %q, want %q", tt.url, gotRepo, tt.wantRepo)
			}
		})
	}
}

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		// Standard providers
		{"github.com", "github"},
		{"www.github.com", "github"},
		{"gitlab.com", "gitlab"},
		{"www.gitlab.com", "gitlab"},
		{"bitbucket.org", "bitbucket"},
		{"www.bitbucket.org", "bitbucket"},

		// Enterprise/self-hosted
		{"github.company.com", "github"},
		{"gitlab.company.com", "gitlab"},
		{"bitbucket.company.com", "bitbucket"},

		// Case insensitive
		{"GITHUB.COM", "github"},
		{"GitLab.Com", "gitlab"},

		// Unknown
		{"git.company.com", "unknown"},
		{"custom-git.io", "unknown"},
		{"example.com", "unknown"},
		{"", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := detectProvider(tt.host)
			if got != tt.want {
				t.Errorf("detectProvider(%q) = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}

func TestTaskCodePattern(t *testing.T) {
	tests := []struct {
		input string
		match bool
	}{
		{"GS-123", true},
		{"TK-1", true},
		{"ABCDE-99999", true},
		{"AB-0", true},
		{"gs-123", false},  // lowercase
		{"G-123", false},   // single char prefix
		{"ABCDEF-1", true}, // matches BCDEF-1 substring
		{"GS123", false},   // no hyphen
		{"GS-", false},     // no number
		{"-123", false},    // no prefix
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := TaskCodePattern.MatchString(tt.input)
			if got != tt.match {
				t.Errorf("TaskCodePattern.MatchString(%q) = %v, want %v", tt.input, got, tt.match)
			}
		})
	}
}

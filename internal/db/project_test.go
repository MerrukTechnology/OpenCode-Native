package db

import (
	"testing"
)

// testCase represents a common test case structure for string transformation tests.
type testCase struct {
	name     string
	input    string
	expected string
}

// runStringTests is a helper function that runs a set of test cases against a test function.
func runStringTests(t *testing.T, tests []testCase, testFn func(string) string) {
	t.Run("table-driven", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := testFn(tt.input)
				if result != tt.expected {
					t.Errorf("%s(%q) = %q, want %q", t.Name(), tt.input, result, tt.expected)
				}
			})
		}
	})
}

func TestNormalizeGitURL(t *testing.T) {
	tests := []testCase{
		{name: "HTTPS URL with .git suffix", input: "https://github.com/MerrukTechnology/OpenCode-Native.git", expected: "github.com/MerrukTechnology/OpenCode-Native"},
		{name: "HTTPS URL without .git suffix", input: "https://gitlab.com/myteam/myproject", expected: "gitlab.com/myteam/myproject"},
		{name: "SSH URL with .git suffix", input: "git@github.com:MerrukTechnology/OpenCode-Native.git", expected: "github.com/MerrukTechnology/OpenCode-Native"},
		{name: "SSH URL without .git suffix", input: "git@gitlab.com:myteam/myproject", expected: "gitlab.com/myteam/myproject"},
		{name: "HTTP URL with .git suffix", input: "http://github.com/MerrukTechnology/OpenCode-Native.git", expected: "github.com/MerrukTechnology/OpenCode-Native"},
		{name: "URL with trailing slash", input: "https://github.com/MerrukTechnology/OpenCode-Native/", expected: "github.com/MerrukTechnology/OpenCode-Native"},
		{name: "URL with trailing slash and .git", input: "https://github.com/MerrukTechnology/OpenCode-Native.git/", expected: "github.com/MerrukTechnology/OpenCode-Native"},
		{name: "Plain URL without protocol", input: "github.com/MerrukTechnology/OpenCode-Native", expected: "github.com/MerrukTechnology/OpenCode-Native"},
	}
	runStringTests(t, tests, normalizeGitURL)
}

func TestGetProjectIDFromDirectory(t *testing.T) {
	tests := []testCase{
		{name: "Unix path", input: "/Users/john/projects/my-app", expected: "my-app"},
		{name: "Unix path with trailing slash", input: "/Users/john/projects/my-app/", expected: "my-app"},
		{name: "Relative path", input: "./my-app", expected: "my-app"},
		{name: "Single directory", input: "my-app", expected: "my-app"},
	}
	runStringTests(t, tests, getProjectIDFromDirectory)
}

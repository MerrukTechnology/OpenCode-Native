// Package editor provides UI components for code viewing and editing in the TUI.
// This file contains syntax highlighting functionality using chroma.
package editor

import (
	"strings"

	"github.com/MerrukTechnology/OpenCode-Native/internal/tui/theme"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
)

// fileExtensionToLanguage maps file extensions to chroma language identifiers.
var fileExtensionToLanguage = map[string]string{
	".go":         "go",
	".rs":         "rust",
	".py":         "python",
	".js":         "javascript",
	".ts":         "typescript",
	".jsx":        "javascript",
	".tsx":        "typescript",
	".java":       "java",
	".c":          "c",
	".cpp":        "cpp",
	".h":          "c",
	".hpp":        "cpp",
	".cs":         "csharp",
	".rb":         "ruby",
	".php":        "php",
	".swift":      "swift",
	".kt":         "kotlin",
	".scala":      "scala",
	".html":       "html",
	".htm":        "html",
	".css":        "css",
	".scss":       "scss",
	".less":       "less",
	".json":       "json",
	".yaml":       "yaml",
	".yml":        "yaml",
	".xml":        "xml",
	".toml":       "toml",
	".md":         "markdown",
	".sql":        "sql",
	".sh":         "bash",
	".bash":       "bash",
	".zsh":        "bash",
	".fish":       "bash",
	".ps1":        "powershell",
	".psm1":       "powershell",
	".dockerfile": "dockerfile",
	".makefile":   "makefile",
	"Makefile":    "makefile",
	".r":          "r",
	".lua":        "lua",
	".pl":         "perl",
	".hs":         "haskell",
	".ex":         "elixir",
	".exs":        "elixir",
	".erl":        "erlang",
	".clj":        "clojure",
	".vim":        "vim",
	".s":          "asm",
	".asm":        "asm",
}

// languageAliases provides friendly names for language detection.
var languageAliases = map[string]string{
	"golang":      "go",
	"py":          "python",
	"js":          "javascript",
	"ts":          "typescript",
	"c#":          "csharp",
	"c++":         "cpp",
	"objectivec":  "objectivec",
	"objective-c": "objectivec",
	"shell":       "bash",
	"zsh":         "bash",
	"sh":          "bash",
	"yml":         "yaml",
}

// defaultStyle is the fallback chroma style when none is specified.
const defaultStyle = "monokai"

// Highlight highlights the given code with syntax highlighting using chroma.
// It takes the code content and an optional language identifier.
// If language is empty, it will attempt to detect the language from the filename.
func Highlight(code string, language string) string {
	return HighlightWithFilename(code, language, "")
}

// HighlightWithFilename highlights the given code with syntax highlighting.
// It uses the filename to detect the language if language is not specified.
func HighlightWithFilename(code string, language string, filename string) string {
	// Get the theme's syntax style
	styleName := defaultStyle
	if t := theme.CurrentTheme(); t != nil {
		if themeName := t.GetSyntaxTheme(); themeName != "" {
			styleName = themeName
		}
	}

	// If no language specified, try to detect from filename
	if language == "" && filename != "" {
		language = detectLanguageFromFilename(filename)
	}

	// Normalize language name
	language = normalizeLanguage(language)

	// Get lexer
	var lexer chroma.Lexer
	if language != "" {
		lexer = lexers.Get(language)
	}
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// Get style
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}

	// Create formatter
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Get theme colors for lipgloss
	t := theme.CurrentTheme()
	textColor := "#d4d4d4" // Default text color
	if t != nil {
		textColor = getColorHex(t.Text())
	}

	// Highlight the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		// Return plain text on error
		return lipgloss.NewStyle().Foreground(lipgloss.Color(textColor)).Render(code)
	}

	// Render with chroma formatter
	var result strings.Builder
	err = formatter.Format(&result, style, iterator)
	if err != nil {
		// Return plain text on error
		return lipgloss.NewStyle().Foreground(lipgloss.Color(textColor)).Render(code)
	}

	return result.String()
}

// detectLanguageFromFilename detects the language from a filename.
func detectLanguageFromFilename(filename string) string {
	// Check for exact filename matches first (like Makefile)
	if strings.HasPrefix(filename, "Makefile") {
		return "makefile"
	}

	// Find the last dot for extension
	lastDot := strings.LastIndex(filename, ".")
	if lastDot == -1 {
		return ""
	}

	ext := strings.ToLower(filename[lastDot:])
	return fileExtensionToLanguage[ext]
}

// normalizeLanguage normalizes a language identifier to chroma format.
func normalizeLanguage(language string) string {
	if language == "" {
		return ""
	}

	language = strings.ToLower(language)

	// Check aliases
	if alias, ok := languageAliases[language]; ok {
		return alias
	}

	return language
}

// GetSupportedLanguages returns a list of supported language identifiers.
func GetSupportedLanguages() []string {
	return []string{
		"go", "rust", "python", "javascript", "typescript", "java", "c", "cpp",
		"csharp", "ruby", "php", "swift", "kotlin", "scala", "html", "css",
		"scss", "less", "json", "yaml", "xml", "toml", "markdown", "sql",
		"bash", "powershell", "dockerfile", "makefile", "r", "lua", "perl",
		"haskell", "elixir", "erlang", "clojure", "vim", "asm",
	}
}

// getColorHex converts an AdaptiveColor to a hex string for chroma.
func getColorHex(color lipgloss.AdaptiveColor) string {
	// Return dark by default as we typically use dark terminals
	// The AdaptiveColor stores both light and dark variants
	// For chroma, we need to provide a hex value - we'll use dark mode
	return color.Dark
}

// Package shared provides common UI components for the OpenCode TUI,
// including modals, spinners, and icon constants.
package shared

// Icon constants using Nerd Fonts glyphs for consistent UI icons.
// These icons are used throughout the TUI for visual indicators.
const (
	// OpenCode specific icons
	IconOpenCode = "⌬" // OpenCode logo icon

	// General UI Icons
	IconCheck        = "✓"  // Checkmark / success
	IconCheckAlt     = "✓"  // Checkmark alternative
	IconError        = "✖"  // Error / close
	IconErrorAlt     = "✗"  // Error alternative / failure
	IconWarning      = "⚠"  // Warning indicator
	IconInfo         = "ℹ"  // Information
	IconHint         = "💡"  // Hint / tip
	IconHintAlt      = "i"  // Hint alternative (simple)
	IconQuestion     = "❓"  // Question mark
	IconStar         = "★"  // Star / favorite
	IconStarOutline  = "☆"  // Star outline
	IconHeart        = "♥"  // Heart
	IconHeartOutline = "♡"  // Heart outline
	IconLightning    = "⚡"  // Lightning bolt
	IconPlug         = "🔌"  // Plug / power
	IconTool         = "🛠️" // Tool / wrench

	// Navigation Icons
	IconArrowRight   = "→" // Arrow right
	IconArrowLeft    = "←" // Arrow left
	IconArrowUp      = "↑" // Arrow up
	IconArrowDown    = "↓" // Arrow down
	IconChevronRight = "›" // Chevron right
	IconChevronLeft  = "‹" // Chevron left
	IconChevronUp    = "⌃" // Chevron up
	IconChevronDown  = "⌄" // Chevron down
	IconCaretDown    = "▼"
	IconCaretRight   = "▶"
	IconStatusDot    = "●"

	// File and Folder Icons
	IconFolder     = "📁" // Folder
	IconFolderOpen = "📂" // Open folder
	IconFile       = "📄" // Generic file
	IconDocument   = "📝" // Document
	IconImage      = "🖼" // Image file
	IconCode       = "💻" // Code file
	IconDatabase   = "🗄" // Database
	IconArchive    = "📦" // Archive / zip

	// Action Icons
	IconSearch   = "🔍" // Search
	IconZoomIn   = "🔎" // Zoom in
	IconZoomOut  = "🔍" // Zoom out
	IconEdit     = "✏" // Edit
	IconPencil   = "✎" // Pencil
	IconCopy     = "📋" // Copy
	IconCut      = "✂" // Cut
	IconPaste    = "📌" // Paste
	IconTrash    = "🗑" // Delete / trash
	IconSave     = "💾" // Save
	IconDownload = "⬇" // Download
	IconUpload   = "⬆" // Upload
	IconRefresh  = "⟳" // Refresh / reload
	IconSync     = "🔄" // Sync
	IconSettings = "⚙" // Settings / gear
	IconCog      = "⚙" // Settings (alternative)
	IconPlus     = "+" // Add / plus
	IconMinus    = "−" // Remove / minus
	IconClose    = "✕" // Close
	IconMenu     = "☰" // Menu / hamburger
	IconList     = "☰" // List view

	// Status Icons
	IconPlay        = "▶" // Play
	IconPause       = "⏸" // Pause
	IconStop        = "⏹" // Stop
	IconRecord      = "⏺" // Record
	IconRewind      = "⏪" // Rewind
	IconFastForward = "⏩" // Fast forward
	IconSkipBack    = "⏮" // Skip back
	IconSkipForward = "⏭" // Skip forward

	// Chat and Message Icons
	IconMessage        = "💬" // Speech bubble
	IconMessageOutline = "💭" // Thought bubble
	IconBot            = "🤖" // Bot / robot
	IconUser           = "👤" // User
	IconUsers          = "👥" // Users group
	IconPerson         = "🧑" // Person
	IconSend           = "➤" // Send
	IconMail           = "✉" // Email / mail

	// Technology Icons
	IconTerminal   = "⌨" // Terminal
	IconCommand    = "⌘" // Command / cmd
	IconKey        = "⌥" // Option / alt key
	IconKeyboard   = "⌨" // Keyboard
	IconLink       = "🔗" // Link / hyperlink
	IconLinkBroken = "🔗" // Broken link
	IconLock       = "🔒" // Lock / locked
	IconLockOpen   = "🔓" // Unlock / unlocked
	IconWifi       = "📶" // WiFi signal
	IconSignal     = "📡" // Signal
	IconBluetooth  = "📘" // Bluetooth

	// Development Icons
	IconBug         = "🐛" // Bug
	IconCommit      = "⬤" // Commit / dot
	IconBranch      = "⎇" // Branch
	IconGit         = "⎇" // Git branch
	IconMerge       = "⇄" // Merge
	IconPullRequest = "⇄" // Pull request
	IconIssue       = "⚠" // Issue / warning
	IconPullDown    = "⇩" // Pull down

	// Time and Date Icons
	IconClock    = "🕐" // Clock
	IconTime     = "🕒" // Time
	IconCalendar = "📅" // Calendar
	IconAlarm    = "⏰" // Alarm

	// Misc Icons
	IconLightbulb = "💡" // Lightbulb / idea
	IconMagic     = "✨" // Sparkles / magic
	IconFire      = "🔥" // Fire / hot
	IconRocket    = "🚀" // Rocket
	IconDiamond   = "💎" // Diamond / gem
	IconTrophy    = "🏆" // Trophy
	IconMedal     = "🏅" // Medal
	IconCrown     = "👑" // Crown
	IconGem       = "💠" // Gem / diamond

	// Loading/Progress Icons
	IconSpinner   = "⟳" // Spinner / loading
	IconLoading   = "⏳" // Hourglass / loading
	IconHourglass = "⏳" // Hourglass
	IconCircle    = "⬤" // Circle / dot

	// Nerd Fonts Dev Icons (alternative with NF specific glyphs)
	// These are the proper Nerd Fonts versions if available
	NFArrowRight   = "→"
	NFArrowLeft    = "←"
	NFArrowUp      = "↑"
	NFArrowDown    = "↓"
	NFChevronRight = "›"
	NFChevronLeft  = "‹"
	NFClose        = "✕"
	NFCheck        = "✓"
	NFError        = "✗"
	NFStar         = "★"
	NFStarOutline  = "☆"
	NFTerminal     = "⌨"
	NFCode         = "⌘"
	NFKey          = "⌥"
	NFGitBranch    = "⎇"
	NFGitCommit    = "⬤"
	NFConfig       = "⚙"
	NFSettings     = "⚙"
	NFSearch       = "🔍"
	NFRefresh      = "⟳"
	NFSpinner      = "⠄"
)

// IconSet provides a categorized set of icons for specific UI contexts.
var (
	// StatusIcons contains icons for different status states
	StatusIcons = map[string]string{
		"success":  IconCheck,
		"error":    IconError,
		"warning":  IconWarning,
		"info":     IconInfo,
		"loading":  IconSpinner,
		"pending":  IconClock,
		"complete": IconCheck,
	}

	// FileTypeIcons maps file extensions to their corresponding icons
	FileTypeIcons = map[string]string{
		"go":   IconCode,
		"js":   IconCode,
		"ts":   IconCode,
		"py":   IconCode,
		"rs":   IconCode,
		"java": IconCode,
		"c":    IconCode,
		"cpp":  IconCode,
		"h":    IconCode,
		"hpp":  IconCode,
		"md":   IconDocument,
		"txt":  IconDocument,
		"json": IconCode,
		"yaml": IconCode,
		"yml":  IconCode,
		"toml": IconCode,
		"png":  IconImage,
		"jpg":  IconImage,
		"jpeg": IconImage,
		"gif":  IconImage,
		"svg":  IconImage,
		"zip":  IconArchive,
		"tar":  IconArchive,
		"gz":   IconArchive,
		"db":   IconDatabase,
		"sql":  IconDatabase,
	}

	// ActionIcons contains icons for user actions
	ActionIcons = map[string]string{
		"add":      IconPlus,
		"remove":   IconMinus,
		"edit":     IconEdit,
		"delete":   IconTrash,
		"save":     IconSave,
		"search":   IconSearch,
		"copy":     IconCopy,
		"paste":    IconPaste,
		"cut":      IconCut,
		"refresh":  IconRefresh,
		"settings": IconSettings,
		"close":    IconClose,
		"menu":     IconMenu,
		"download": IconDownload,
		"upload":   IconUpload,
	}
)

// GetIcon returns the icon for the given name, or a default if not found.
func GetIcon(name string) string {
	if icon, ok := StatusIcons[name]; ok {
		return icon
	}
	if icon, ok := ActionIcons[name]; ok {
		return icon
	}
	return IconQuestion
}

// GetFileIcon returns the appropriate icon for a file based on its extension.
func GetFileIcon(filename string) string {
	// Try to get extension
	ext := ""
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			ext = filename[i+1:]
			break
		}
		if filename[i] == '/' || filename[i] == '\\' {
			break
		}
	}

	if ext != "" {
		if icon, ok := FileTypeIcons[ext]; ok && icon != "" {
			return icon
		}
	}

	// Check if it's a directory-like name
	if filename == "" || filename[len(filename)-1] == '/' {
		return IconFolder
	}

	return IconFile
}

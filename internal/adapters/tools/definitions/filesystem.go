package definitions

import "github.com/phamdaiminhquan/vibe-devops/internal/ports"

// ReadFile defines the read_file tool metadata
var ReadFile = ports.ToolDefinition{
	Name:         "read_file",
	DisplayTitle: "Read File",
	Description:  "Read a text file, optionally by line range. Returns file content with line numbers.",
	WouldLikeTo:  "read the following file",
	IsCurrently:  "reading file",
	HasAlready:   "read the file",
	ReadOnly:     true,
	InputSchema: `{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "Path to the file to read"
			},
			"startLine": {
				"type": "integer",
				"description": "Starting line number (1-based, optional)"
			},
			"endLine": {
				"type": "integer",
				"description": "Ending line number (inclusive, optional)"
			},
			"maxBytes": {
				"type": "integer",
				"description": "Maximum bytes to read (default: 65536)"
			}
		},
		"required": ["path"]
	}`,
	DefaultPolicy: ports.PolicyAllowed,
	Group:         "filesystem",
}

// ListDir defines the list_dir tool metadata
var ListDir = ports.ToolDefinition{
	Name:         "list_dir",
	DisplayTitle: "List Directory",
	Description:  "List entries in a directory. Returns file and folder names.",
	WouldLikeTo:  "list the contents of directory",
	IsCurrently:  "listing directory",
	HasAlready:   "listed the directory",
	ReadOnly:     true,
	InputSchema: `{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "Path to the directory to list"
			},
			"maxEntries": {
				"type": "integer",
				"description": "Maximum entries to return (default: 200)"
			}
		},
		"required": ["path"]
	}`,
	DefaultPolicy: ports.PolicyAllowed,
	Group:         "filesystem",
}

// Grep defines the grep tool metadata
var Grep = ports.ToolDefinition{
	Name:         "grep",
	DisplayTitle: "Grep Search",
	Description:  "Search for a pattern in files. Returns matching lines with context.",
	WouldLikeTo:  "search for pattern in files",
	IsCurrently:  "searching files",
	HasAlready:   "searched the files",
	ReadOnly:     true,
	InputSchema: `{
		"type": "object",
		"properties": {
			"pattern": {
				"type": "string",
				"description": "Pattern to search for"
			},
			"path": {
				"type": "string",
				"description": "File or directory to search in"
			},
			"maxResults": {
				"type": "integer",
				"description": "Maximum results to return (default: 100)"
			}
		},
		"required": ["pattern", "path"]
	}`,
	DefaultPolicy: ports.PolicyAllowed,
	Group:         "filesystem",
}

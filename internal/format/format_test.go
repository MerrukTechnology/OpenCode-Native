package format

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected OutputFormat
		wantErr  bool
	}{
		{"text", Text, false},
		{"json", JSON, false},
		{"json_schema", JSONSchema, false},
		{"TEXT", Text, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.expected {
				t.Errorf("Parse(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseWithSchema(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantFormat OutputFormat
		wantSchema bool
		wantErr    bool
	}{
		{
			name:       "plain text",
			input:      "text",
			wantFormat: Text,
			wantSchema: false,
		},
		{
			name:       "plain json",
			input:      "json",
			wantFormat: JSON,
			wantSchema: false,
		},
		{
			name:       "json_schema with schema",
			input:      `json_schema={"type":"object","properties":{"name":{"type":"string"}}}`,
			wantFormat: JSONSchema,
			wantSchema: true,
		},
		{
			name:       "json_schema with quoted schema",
			input:      `json_schema='{"type":"object","properties":{"name":{"type":"string"}}}'`,
			wantFormat: JSONSchema,
			wantSchema: true,
		},
		{
			name:    "json_schema with invalid JSON and non-existent file",
			input:   `json_schema=not_a_file_or_json`,
			wantErr: true,
		},
		{
			name:    "json_schema missing type",
			input:   `json_schema={"properties":{}}`,
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "xml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, schema, err := ParseWithSchema(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWithSchema(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if format != tt.wantFormat {
				t.Errorf("format = %q, want %q", format, tt.wantFormat)
			}
			if tt.wantSchema && schema == nil {
				t.Error("expected schema to be non-nil")
			}
			if !tt.wantSchema && schema != nil {
				t.Error("expected schema to be nil")
			}
		})
	}
}

func TestParseWithSchema_FileHandling(t *testing.T) {
	dir := t.TempDir()
	schemaFile := filepath.Join(dir, "schema.json")
	os.WriteFile(schemaFile, []byte(`{"type":"object","properties":{"name":{"type":"string"}}}`), 0o644)

	realSchemaFile := filepath.Join(dir, "real-schema.json")
	os.WriteFile(realSchemaFile, []byte(`{"type":"object","properties":{"score":{"type":"number"}},"required":["score"]}`), 0o644)

	badSchemaFile := filepath.Join(dir, "bad.json")
	os.WriteFile(badSchemaFile, []byte(`not json`), 0o644)

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, format OutputFormat, schema map[string]any)
	}{
		{
			name:  "file path loads schema",
			input: "json_schema=" + schemaFile,
			check: func(t *testing.T, format OutputFormat, schema map[string]any) {
				if format != JSONSchema {
					t.Errorf("format = %q, want %q", format, JSONSchema)
				}
				if schema == nil {
					t.Fatal("expected schema to be non-nil")
				}
				if schema["type"] != "object" {
					t.Errorf("schema type = %v, want 'object'", schema["type"])
				}
			},
		},
		{
			name:  "$ref to file loads schema",
			input: `json_schema={"$ref":"` + realSchemaFile + `"}`,
			check: func(t *testing.T, format OutputFormat, schema map[string]any) {
				if format != JSONSchema {
					t.Errorf("format = %q, want %q", format, JSONSchema)
				}
				if schema == nil {
					t.Fatal("expected schema to be non-nil")
				}
				if schema["type"] != "object" {
					t.Errorf("schema type = %v, want 'object'", schema["type"])
				}
				props, ok := schema["properties"].(map[string]any)
				if !ok {
					t.Fatal("expected properties map")
				}
				if _, ok := props["score"]; !ok {
					t.Error("expected 'score' property from referenced file")
				}
			},
		},
		{
			name:    "$ref to non-existent path returns error",
			input:   `json_schema={"$ref":"/nonexistent/path.json"}`,
			wantErr: true,
		},
		{
			name:    "$ref non-string returns error",
			input:   `json_schema={"$ref":42}`,
			wantErr: true,
		},
		{
			name:    "file path with invalid JSON returns error",
			input:   "json_schema=" + badSchemaFile,
			wantErr: true,
		},
		{
			name:    "non-existent file path returns error",
			input:   "json_schema=/tmp/nonexistent-schema-12345.json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, schema, err := ParseWithSchema(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWithSchema(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.check != nil {
				tt.check(t, format, schema)
			}
		})
	}
}

func TestValidateJSONSchema(t *testing.T) {
	tests := []struct {
		name    string
		schema  map[string]any
		wantErr bool
	}{
		{
			name:    "nil schema",
			schema:  nil,
			wantErr: true,
		},
		{
			name:    "missing type",
			schema:  map[string]any{"properties": map[string]any{}},
			wantErr: true,
		},
		{
			name:    "non-string type",
			schema:  map[string]any{"type": 42},
			wantErr: true,
		},
		{
			name:   "valid schema",
			schema: map[string]any{"type": "object"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSONSchema(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJSONSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatOutput(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		outputFormat OutputFormat
		want         string
	}{
		{
			name:         "text format returns content as-is",
			content:      "hello",
			outputFormat: Text,
			want:         "hello",
		},
		{
			name:         "json_schema returns content as-is",
			content:      `{"summary":"test"}`,
			outputFormat: JSONSchema,
			want:         `{"summary":"test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatOutput(tt.content, tt.outputFormat)
			if got != tt.want {
				t.Errorf("FormatOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"text", true},
		{"json", true},
		{"json_schema", true},
		{`json_schema={"type":"object"}`, true},
		{"xml", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsValid(tt.input); got != tt.valid {
				t.Errorf("IsValid(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestResolveSchemaString(t *testing.T) {
	dir := t.TempDir()
	schemaFile := filepath.Join(dir, "schema.json")
	os.WriteFile(schemaFile, []byte(`{"type":"object","properties":{"x":{"type":"number"}}}`), 0o644)

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, schema map[string]any)
	}{
		{
			name:  "inline JSON",
			input: `{"type":"string"}`,
			check: func(t *testing.T, schema map[string]any) {
				if schema["type"] != "string" {
					t.Errorf("type = %v, want 'string'", schema["type"])
				}
			},
		},
		{
			name:  "file path",
			input: schemaFile,
			check: func(t *testing.T, schema map[string]any) {
				if schema["type"] != "object" {
					t.Errorf("type = %v, want 'object'", schema["type"])
				}
			},
		},
		{
			name:  "$ref to file",
			input: `{"$ref":"` + schemaFile + `"}`,
			check: func(t *testing.T, schema map[string]any) {
				if schema["type"] != "object" {
					t.Errorf("type = %v, want 'object'", schema["type"])
				}
			},
		},
		{
			name:  "$ref ignores other fields",
			input: `{"$ref":"` + schemaFile + `","description":"ignored"}`,
			check: func(t *testing.T, schema map[string]any) {
				if _, ok := schema["description"]; ok {
					t.Error("expected 'description' from inline to be ignored when $ref is present")
				}
			},
		},
		{
			name:    "$ref empty string returns error",
			input:   `{"$ref":""}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := resolveSchemaString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveSchemaString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.check != nil {
				tt.check(t, schema)
			}
		})
	}
}

func TestResolveSchemaRef(t *testing.T) {
	dir := t.TempDir()
	schemaFile := filepath.Join(dir, "real-schema.json")
	os.WriteFile(schemaFile, []byte(`{"type":"object","properties":{"score":{"type":"number"}}}`), 0o755)

	subDir := filepath.Join(dir, "sub")
	os.MkdirAll(subDir, 0o755)

	tests := []struct {
		name    string
		schema  map[string]any
		baseDir string
		wantErr bool
		check   func(t *testing.T, result map[string]any)
	}{
		{
			name:   "no $ref returns schema unchanged",
			schema: map[string]any{"type": "object", "properties": map[string]any{}},
			check: func(t *testing.T, result map[string]any) {
				if result["type"] != "object" {
					t.Errorf("type = %v, want 'object'", result["type"])
				}
			},
		},
		{
			name:   "nil schema returns nil",
			schema: nil,
			check: func(t *testing.T, result map[string]any) {
				if result != nil {
					t.Error("expected nil result for nil schema")
				}
			},
		},
		{
			name:   "absolute $ref",
			schema: map[string]any{"$ref": schemaFile},
			check: func(t *testing.T, result map[string]any) {
				if result["type"] != "object" {
					t.Errorf("type = %v, want 'object'", result["type"])
				}
			},
		},
		{
			name:    "relative $ref resolved against baseDir",
			schema:  map[string]any{"$ref": "real-schema.json"},
			baseDir: dir,
			check: func(t *testing.T, result map[string]any) {
				if result["type"] != "object" {
					t.Errorf("type = %v, want 'object'", result["type"])
				}
				props, ok := result["properties"].(map[string]any)
				if !ok {
					t.Fatal("expected properties map")
				}
				if _, ok := props["score"]; !ok {
					t.Error("expected 'score' property from file")
				}
			},
		},
		{
			name:    "relative $ref with ../ resolved against baseDir",
			schema:  map[string]any{"$ref": "../real-schema.json"},
			baseDir: subDir,
			check: func(t *testing.T, result map[string]any) {
				if result["type"] != "object" {
					t.Errorf("type = %v, want 'object'", result["type"])
				}
			},
		},
		{
			name:    "relative $ref without baseDir uses cwd returns error",
			schema:  map[string]any{"$ref": "/nonexistent/path.json"},
			baseDir: "",
			wantErr: true,
		},
		{
			name:    "$ref non-string value returns error",
			schema:  map[string]any{"$ref": 42},
			baseDir: "",
			wantErr: true,
		},
		{
			name:    "$ref empty string returns error",
			schema:  map[string]any{"$ref": ""},
			baseDir: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveSchemaRef(tt.schema, tt.baseDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveSchemaRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

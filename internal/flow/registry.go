// Package flow provides discovery, caching, and validation for YAML-defined flows.
// Filenames map to flow IDs (kebab-case) and output schemas may use $ref which are resolved at load.
package flow

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/MerrukTechnology/OpenCode-Native/internal/config"
	"github.com/MerrukTechnology/OpenCode-Native/internal/format"
	"github.com/MerrukTechnology/OpenCode-Native/internal/logging"
	"gopkg.in/yaml.v3"
)

const (
	// Upper bound on accepted flow YAML size to keep discovery fast and safe.
	maxFlowFileSize = 100 * 1024 // 100KB
	// Maximum length for flow IDs (derived from filenames).
	maxNameLength = 64
	// Maximum length for step IDs inside a flow.
	maxStepIDLength = 64
)

var (
	// Validates flow and step identifiers: lowercase alphanumeric with hyphens.
	kebabCaseRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

	// Cache of discovered flows guarded by flowCacheLock.
	flowCache     map[string]Flow
	flowCacheLock sync.Mutex
	flowCacheInit bool

	// flowConflicts holds duplicate flow IDs discovered during flow loading.
	flowConflicts *Conflicts
	conflictsLock sync.Mutex
)

// flowFile represents the raw YAML structure of a flow file.
type flowFile struct {
	Name        string   `yaml:"name"`
	Disabled    bool     `yaml:"disabled,omitempty"`
	Description string   `yaml:"description"`
	Flow        FlowSpec `yaml:"flow"`
}

// Get returns a flow by ID, or ErrFlowNotFound.
func Get(id string) (*Flow, error) {
	flows := state()
	f, ok := flows[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrFlowNotFound, id)
	}
	return &f, nil
}

// All returns all discovered flows.
func All() []Flow {
	flows := state()
	result := make([]Flow, 0, len(flows))
	for _, f := range flows {
		result = append(result, f)
	}
	return result
}

// GetConflicts returns any duplicate flow ID conflicts discovered during loading.
func GetConflicts() *Conflicts {
	state() // Ensure flows are discovered
	conflictsLock.Lock()
	defer conflictsLock.Unlock()
	return flowConflicts
}

// Invalidate clears the cached flows, forcing re-discovery on next access.
func Invalidate() {
	flowCacheLock.Lock()
	defer flowCacheLock.Unlock()
	flowCache = nil
	flowCacheInit = false

	conflictsLock.Lock()
	defer conflictsLock.Unlock()
	flowConflicts = nil
}

// state returns the memoized set of flows, discovering on first access.
func state() map[string]Flow {
	flowCacheLock.Lock()
	defer flowCacheLock.Unlock()
	if !flowCacheInit {
		flowCache, flowConflicts = discoverFlows()
		flowCacheInit = true
	}
	return flowCache
}

// discoverFlows merges project and global definitions, tracking duplicate IDs.
// Returns flows map and any conflicts found (project flows take precedence on conflict).
func discoverFlows() (map[string]Flow, *Conflicts) {
	flows := make(map[string]Flow)
	conflictMap := make(map[string][]string)

	// Project flows have higher priority — discover first, so they win on conflict
	projectFlows := discoverProjectFlows()
	for _, f := range projectFlows {
		if existing, exists := flows[f.ID]; exists {
			// Track the conflict: keep the first occurrence, track duplicates
			conflictMap[f.ID] = append(conflictMap[f.ID], existing.Location, f.Location)
			continue
		}
		flows[f.ID] = f
	}

	// Global flows — only add if not already found
	globalFlows := discoverGlobalFlows()
	for _, f := range globalFlows {
		if existing, exists := flows[f.ID]; exists {
			// Track the conflict: keep the project flow, track global duplicate
			conflictMap[f.ID] = append(conflictMap[f.ID], existing.Location, f.Location)
			continue
		}
		flows[f.ID] = f
	}

	// Build Conflicts struct if there are any duplicates
	var conflicts *Conflicts
	if len(conflictMap) > 0 {
		conflicts = &Conflicts{
			Conflicts: make([]FlowConflict, 0, len(conflictMap)),
		}
		for id, locations := range conflictMap {
			// Remove duplicates from locations
			seen := make(map[string]bool)
			uniqueLocations := make([]string, 0)
			for _, loc := range locations {
				if !seen[loc] {
					seen[loc] = true
					uniqueLocations = append(uniqueLocations, loc)
				}
			}
			conflicts.Conflicts = append(conflicts.Conflicts, FlowConflict{
				ID:        id,
				Locations: uniqueLocations,
			})
		}
		// Log warning about conflicts
		logging.Warn("Duplicate flow IDs found, using first occurrence", "count", len(conflictMap))
	}

	return flows, conflicts
}

func discoverProjectFlows() []Flow {
	cfg := config.Get()
	var result []Flow

	projectDirs := []string{
		filepath.Join(cfg.WorkingDir, ".opencode", "flows"),
		filepath.Join(cfg.WorkingDir, ".agents", "flows"),
	}

	for _, dir := range projectDirs {
		result = append(result, scanFlowDirectory(dir)...)
	}
	return result
}

func discoverGlobalFlows() []Flow {
	home, err := os.UserHomeDir()
	if err != nil {
		logging.Warn("Could not determine home directory for global flow discovery", "error", err)
		return nil
	}

	var result []Flow
	globalDirs := []string{
		filepath.Join(home, ".config", "opencode", "flows"),
		filepath.Join(home, ".agents", "flows"),
	}

	for _, dir := range globalDirs {
		result = append(result, scanFlowDirectory(dir)...)
	}
	return result
}

// scanFlowDirectory parses all YAML files in dir as flows, skipping invalid ones.
func scanFlowDirectory(dir string) []Flow {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		logging.Warn("Failed to read flow directory", "dir", dir, "error", err)
		return nil
	}

	var flows []Flow
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		path := filepath.Join(dir, name)
		f, err := parseFlowFile(path)
		if err != nil {
			logging.Warn("Failed to parse flow file", "path", path, "error", err)
			continue
		}
		flows = append(flows, *f)
	}
	return flows
}

// parseFlowFile loads YAML into a Flow, deriving ID from filename, resolving
// $ref in output schemas, and validating the final spec.
func parseFlowFile(path string) (*Flow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading flow file: %w", err)
	}

	if len(data) > maxFlowFileSize {
		return nil, fmt.Errorf("%w: file exceeds %d bytes", ErrInvalidYAML, maxFlowFileSize)
	}

	var ff flowFile
	if err := yaml.Unmarshal(data, &ff); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidYAML, err)
	}

	// Derive ID from filename (basename without extension)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	id := strings.TrimSuffix(base, ext)

	if err := validateFlowID(id); err != nil {
		return nil, err
	}

	// Resolve $ref in step output schemas
	baseDir := filepath.Dir(path)
	for i, step := range ff.Flow.Steps {
		if step.Output != nil && step.Output.Schema != nil {
			resolved, err := format.ResolveSchemaRef(step.Output.Schema, baseDir)
			if err != nil {
				return nil, fmt.Errorf("resolving output schema $ref for step %q: %w", step.ID, err)
			}
			ff.Flow.Steps[i].Output.Schema = resolved
		}
	}

	flow := Flow{
		ID:          id,
		Name:        ff.Name,
		Disabled:    ff.Disabled,
		Description: ff.Description,
		Spec:        ff.Flow,
		Location:    path,
	}

	if err := validateFlow(&flow); err != nil {
		return nil, fmt.Errorf("validating flow %q: %w", id, err)
	}

	return &flow, nil
}

// validateFlowID checks the flow ID (derived from filename) is valid kebab-case.
func validateFlowID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: empty ID", ErrInvalidFlowName)
	}
	if len(id) > maxNameLength {
		return fmt.Errorf("%w: %q exceeds %d characters", ErrInvalidFlowName, id, maxNameLength)
	}
	if !kebabCaseRegex.MatchString(id) {
		return fmt.Errorf("%w: %q must be kebab-case (lowercase alphanumeric with hyphens)", ErrInvalidFlowName, id)
	}
	return nil
}

// validateFlow validates the entire flow definition.
func validateFlow(f *Flow) error {
	if len(f.Spec.Steps) == 0 {
		return ErrNoSteps
	}

	// Build a set of step IDs for reference validation
	stepIDs := make(map[string]bool, len(f.Spec.Steps))
	for _, step := range f.Spec.Steps {
		if err := validateStepID(step.ID); err != nil {
			return err
		}
		if strings.TrimSpace(step.Prompt) == "" {
			return fmt.Errorf("step %q has an empty prompt", step.ID)
		}
		if stepIDs[step.ID] {
			return fmt.Errorf("%w: %q", ErrDuplicateStepID, step.ID)
		}
		stepIDs[step.ID] = true
	}

	// Build adjacency list for cycle detection
	adj := make(map[string][]string, len(f.Spec.Steps))
	for _, step := range f.Spec.Steps {
		adj[step.ID] = nil // Initialize to ensure all nodes exist
	}
	for _, step := range f.Spec.Steps {
		for _, rule := range step.Rules {
			adj[step.ID] = append(adj[step.ID], rule.Then)
		}
		if step.Fallback != nil && step.Fallback.To != "" {
			adj[step.ID] = append(adj[step.ID], step.Fallback.To)
		}
	}

	// Check for cycles using DFS
	if cycle := detectCycle(adj, stepIDs); cycle != nil {
		return fmt.Errorf("%w: %s", ErrCycleDetected, strings.Join(cycle, " -> "))
	}

	// Validate rule and fallback references
	thenTargets := make(map[string]int) // track how many rules target each step
	for _, step := range f.Spec.Steps {
		for _, rule := range step.Rules {
			if !stepIDs[rule.Then] {
				return fmt.Errorf("%w: step %q rule references %q", ErrInvalidRule, step.ID, rule.Then)
			}
			thenTargets[rule.Then]++
		}
		if step.Fallback != nil && step.Fallback.To != "" {
			if !stepIDs[step.Fallback.To] {
				return fmt.Errorf("%w: step %q fallback references %q", ErrInvalidFallback, step.ID, step.Fallback.To)
			}
		}
	}

	// Warn about potential convergence (multiple rules targeting same step)
	for targetID, count := range thenTargets {
		if count > 1 {
			logging.Warn("Multiple rules target the same step (potential diamond convergence, first-to-arrive wins)",
				"flow", f.ID, "target_step", targetID, "source_count", count)
		}
	}

	return nil
}

// detectCycle uses DFS to detect cycles in the step graph.
// Returns the cycle path (as slice of step IDs) if found, nil otherwise.
func detectCycle(adj map[string][]string, stepIDs map[string]bool) []string {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var path []string

	var dfs func(node string, currentPath []string) bool
	dfs = func(node string, currentPath []string) bool {
		visited[node] = true
		recStack[node] = true
		currentPath = append(currentPath, node)

		for _, neighbor := range adj[node] {
			if !stepIDs[neighbor] {
				continue // Skip invalid references (will be caught by other validation)
			}
			if !visited[neighbor] {
				if dfs(neighbor, currentPath) {
					return true
				}
			} else if recStack[neighbor] {
				// Cycle detected - build the cycle path
				cycleStart := -1
				for i, n := range currentPath {
					if n == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					path = append(currentPath[cycleStart:], neighbor)
				}
				return true
			}
		}

		recStack[node] = false
		return false
	}

	// Run DFS from each node to handle disconnected components
	for node := range stepIDs {
		if !visited[node] {
			if dfs(node, nil) {
				return path
			}
		}
	}

	return nil
}

// validateStepID checks a step ID is valid kebab-case and within length limit.
func validateStepID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: empty step ID", ErrInvalidStepID)
	}
	if len(id) > maxStepIDLength {
		return fmt.Errorf("%w: %q exceeds %d characters", ErrInvalidStepID, id, maxStepIDLength)
	}
	if !kebabCaseRegex.MatchString(id) {
		return fmt.Errorf("%w: %q must be kebab-case", ErrInvalidStepID, id)
	}
	return nil
}

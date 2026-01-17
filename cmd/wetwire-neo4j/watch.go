// Command watch monitors Go source files and auto-rebuilds on changes.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/lex00/wetwire-neo4j-go/domain"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch [path]",
	Short: "Watch Go source files and auto-rebuild",
	Long: `Monitor Go source files for changes and automatically rebuild
when changes are detected.

The watch command uses file system notifications to detect changes to .go files
in the specified directory (and subdirectories). When changes are detected, it
automatically triggers a rebuild with debouncing to avoid excessive rebuilds.

Examples:
  wetwire-neo4j watch ./schema
  wetwire-neo4j watch ./schema --output schema.json
  wetwire-neo4j watch ./schema --debounce 500ms
  wetwire-neo4j watch ./schema --lint-only

The command will:
  - Monitor all .go files in the specified directory
  - Debounce rapid changes (default 300ms)
  - Show timestamps for each rebuild
  - Display clear success/failure messages
  - Continue watching until interrupted (Ctrl+C)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWatch,
}

func init() {
	watchCmd.Flags().StringP("output", "o", "", "Output file path")
	watchCmd.Flags().StringP("format", "f", "json", "Output format: json or cypher")
	watchCmd.Flags().StringP("debounce", "d", "300ms", "Debounce duration for file changes")
	watchCmd.Flags().Bool("lint-only", false, "Only run lint, skip build")
}

func newWatchCmd() *cobra.Command {
	return watchCmd
}

func runWatch(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	output, _ := cmd.Flags().GetString("output")
	format, _ := cmd.Flags().GetString("format")
	debounceStr, _ := cmd.Flags().GetString("debounce")
	lintOnly, _ := cmd.Flags().GetBool("lint-only")

	// Parse debounce duration
	debounceDuration, err := time.ParseDuration(debounceStr)
	if err != nil {
		return fmt.Errorf("invalid debounce duration: %w", err)
	}

	// Convert path to absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// Verify path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path not found: %s", path)
	}

	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer watcher.Close()

	// Add path to watcher
	if err := addWatchPaths(watcher, absPath, info); err != nil {
		return fmt.Errorf("add watch paths: %w", err)
	}

	fmt.Printf("[%s] Watching %s for changes (debounce: %s)\n", formatTimestamp(time.Now()), absPath, debounceStr)
	fmt.Println("Press Ctrl+C to stop watching")

	// Do initial build
	fmt.Printf("[%s] Initial build...\n", formatTimestamp(time.Now()))
	runWatchBuild(absPath, output, format, lintOnly)

	// Start watching for changes
	debounceTimer := time.NewTimer(0)
	if !debounceTimer.Stop() {
		<-debounceTimer.C
	}

	pendingRebuild := false

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Filter events
			if !shouldProcessWatchEvent(event.Op.String(), event.Name) {
				continue
			}

			// Reset debounce timer
			pendingRebuild = true
			debounceTimer.Reset(debounceDuration)

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "[%s] Watch error: %v\n", formatTimestamp(time.Now()), err)

		case <-debounceTimer.C:
			if pendingRebuild {
				pendingRebuild = false
				fmt.Printf("[%s] Rebuilding...\n", formatTimestamp(time.Now()))
				runWatchBuild(absPath, output, format, lintOnly)
			}
		}
	}
}

// addWatchPaths recursively adds directories to the watcher
func addWatchPaths(watcher *fsnotify.Watcher, path string, info os.FileInfo) error {
	if info.IsDir() {
		// Add the directory
		if err := watcher.Add(path); err != nil {
			return err
		}

		// Walk subdirectories
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				subPath := filepath.Join(path, entry.Name())
				// Skip hidden directories and vendor
				if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "vendor" {
					continue
				}

				subInfo, err := entry.Info()
				if err != nil {
					continue
				}

				if err := addWatchPaths(watcher, subPath, subInfo); err != nil {
					return err
				}
			}
		}
	} else {
		// Watch the directory containing the file
		dir := filepath.Dir(path)
		if err := watcher.Add(dir); err != nil {
			return err
		}
	}

	return nil
}

// shouldProcessWatchEvent determines if an event should trigger a rebuild
func shouldProcessWatchEvent(op, path string) bool {
	// Only process .go files
	if !strings.HasSuffix(path, ".go") {
		return false
	}

	// Skip test files
	if strings.HasSuffix(path, "_test.go") {
		return false
	}

	// Process CREATE, WRITE, REMOVE, RENAME events
	// Ignore CHMOD events
	switch op {
	case "CREATE", "WRITE", "REMOVE", "RENAME":
		return true
	default:
		return false
	}
}

// formatTimestamp formats a time as HH:MM:SS
func formatTimestamp(t time.Time) string {
	return t.Format("15:04:05")
}

// runWatchBuild runs lint and optionally build
func runWatchBuild(path, output, format string, lintOnly bool) {
	d := &domain.Neo4jDomain{}
	ctx := &domain.Context{}

	// Run lint first
	lintResult, err := d.Linter().Lint(ctx, path, domain.LintOpts{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] Lint error: %v\n", formatTimestamp(time.Now()), err)
		return
	}

	if !lintResult.Success {
		fmt.Printf("[%s] Lint issues found:\n", formatTimestamp(time.Now()))
		for _, e := range lintResult.Errors {
			fmt.Printf("  %s: %s\n", e.Path, e.Message)
		}
		return
	}

	fmt.Printf("[%s] Lint passed\n", formatTimestamp(time.Now()))

	if lintOnly {
		return
	}

	// Run build
	buildOpts := domain.BuildOpts{
		Format: format,
		Output: output,
	}

	buildResult, err := d.Builder().Build(ctx, path, buildOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] Build error: %v\n", formatTimestamp(time.Now()), err)
		return
	}

	if !buildResult.Success {
		fmt.Printf("[%s] Build failed:\n", formatTimestamp(time.Now()))
		for _, e := range buildResult.Errors {
			fmt.Printf("  %s: %s\n", e.Path, e.Message)
		}
		return
	}

	if output != "" {
		fmt.Printf("[%s] Build successful: %s\n", formatTimestamp(time.Now()), output)
	} else {
		fmt.Printf("[%s] Build successful\n", formatTimestamp(time.Now()))
	}
}

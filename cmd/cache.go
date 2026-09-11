package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/FacileStudio/sonar/internal/config"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Show the result cache: entry count and location",
	Args:  cobra.NoArgs,
	RunE:  runCacheStatus,
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Empty the result cache",
	Args:  cobra.NoArgs,
	RunE:  runCacheClear,
}

func init() {
	cacheCmd.AddCommand(cacheClearCmd)
}

func runCacheStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath())
	if err != nil {
		return err
	}
	count := cacheEntryCount(cfg.CacheDir)
	fmt.Printf("%d cached result set(s) in %s\n", count, cfg.CacheDir)
	return nil
}

func runCacheClear(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath())
	if err != nil {
		return err
	}
	removed := removeCached(cfg.CacheDir)
	fmt.Printf("cleared %d cached result set(s) from %s\n", removed, cfg.CacheDir)
	return nil
}

func dirEntryPath(dir, name string) string {
	if strings.HasSuffix(dir, "/") {
		return dir + name
	}
	return dir + "/" + name
}

// removeCached deletes the cached result files in dir and returns how many it
// removed; a missing directory is an empty cache, not an error.
func removeCached(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	removed := 0
	for _, e := range entries {
		if !e.IsDir() && os.Remove(dirEntryPath(dir, e.Name())) == nil {
			removed++
		}
	}
	return removed
}

// cacheEntryCount lists the cache directory and counts regular entries; a
// missing directory is an empty cache, not an error.
func cacheEntryCount(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			count++
		}
	}
	return count
}

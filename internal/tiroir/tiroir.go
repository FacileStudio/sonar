// Package tiroir reads secret values from tiroir, the encrypted local env
// store. sonar reaches tiroir directly when a key or zone is absent from the
// config file and the process environment, because an MCP-spawned process
// inherits no shell exports.
package tiroir

import (
	"os/exec"
	"strings"
)

// Get asks tiroir for one value by name, returning "" when tiroir is not
// installed or the value is absent. The trailing newline tiroir prints is
// trimmed so the value lands clean.
func Get(name string) string {
	cmd := exec.Command("tiroir", "get", name)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// CanonicalKey is the primary SONAR_<NAME>_KEY name an engine's secret is
// stored under, both in the environment and in tiroir. The SONAR_ prefix keeps
// sonar's secrets namespaced so they never collide with generic provider keys.
func CanonicalKey(name string) string {
	return "SONAR_" + strings.ToUpper(name) + "_KEY"
}

// EnvCandidates lists the env vars an engine's key may live in: the canonical
// SONAR_<NAME>_KEY plus the common provider-generic names, so keys stored in
// the environment under their bare names (EXA_KEY, FIRECRAWL_KEY, ...) are
// picked up without renaming. tiroir itself only ever stores the canonical,
// SONAR_-namespaced form.
func EnvCandidates(name string) []string {
	switch name {
	case "exa":
		return []string{CanonicalKey(name), "EXA_KEY"}
	case "firecrawl":
		return []string{CanonicalKey(name), "FIRECRAWL_KEY"}
	case "tavily":
		return []string{CanonicalKey(name), "TAVILY_KEY"}
	case "serpapi":
		return []string{CanonicalKey(name), "SERPAPI_API_KEY", "SERPAPI_KEY"}
	case "brave":
		return []string{CanonicalKey(name), "BRAVE_KEY"}
	case "browserbase":
		return []string{CanonicalKey(name), "BROWSERBASE_KEY", "BROWSERBASE_API_KEY"}
	case "brightdata":
		return []string{CanonicalKey(name), "BRIGHTDATA_KEY", "BRIGHTDATA_API_KEY"}
	case "linkup":
		return []string{CanonicalKey(name), "LINKUP_KEY", "LINKUP_API_KEY"}
	case "parallel":
		return []string{CanonicalKey(name), "PARALLEL_API_KEY", "PARALLEL_KEY", "parallel_api_key"}
	default:
		return []string{CanonicalKey(name)}
	}
}

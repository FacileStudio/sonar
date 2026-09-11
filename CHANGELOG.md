# Changelog

All notable changes are documented here. Format derives from Keep a Changelog.

## [Unreleased]

## [v0.6.0]

### Added

- `sonar install` registers the search MCP server in the agent harnesses on
  this machine: claude (~/.claude.json), gemini (~/.gemini/settings.json),
  opencode (~/.config/opencode/opencode.json) and nacelle-tui
  (~/.nacelle.yml sources.mcp). The merge is surgical for the YAML nacelle
  config so comments and other servers survive byte for byte, and refuses
  loudly (exit 1) on a config it cannot safely touch.
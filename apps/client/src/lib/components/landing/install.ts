import type { CliCommand } from '@facile/muse';

/** The sonar CLI's GitHub releases: linux and macOS archives plus checksums.txt. */
export const RELEASES = 'https://github.com/FacileStudio/sonar/releases';

/** The install routes the release pipeline actually publishes: the Homebrew cask and the Go module. */
export const INSTALL: CliCommand[] = [
	{ label: 'Homebrew', command: 'brew install facilestudio/tap/sonar' },
	{ label: 'Go', command: 'go install github.com/FacileStudio/sonar@latest' }
];

import { browser } from '$app/environment';

export type ThemeMode = 'system' | 'light' | 'dark';

const KEY = 'sonar-theme';

function stored(): ThemeMode {
	if (!browser) return 'system';
	const raw = localStorage.getItem(KEY);
	return raw === 'light' || raw === 'dark' || raw === 'system' ? raw : 'system';
}

export const theme = $state({ mode: stored() as ThemeMode });

/*
 * Both classes are written, and `system` writes neither. muse's tokens flip on
 * `prefers-color-scheme` scoped to `:root:not(.light)`, so the `.light` class is the only
 * thing that lets someone force light on a dark OS.
 */
export function setTheme(mode: ThemeMode) {
	theme.mode = mode;
	if (!browser) return;
	const root = document.documentElement;
	root.classList.toggle('dark', mode === 'dark');
	root.classList.toggle('light', mode === 'light');
	localStorage.setItem(KEY, mode);
}

/* What the page is actually painting right now, whatever route got it there. */
export function isDark(): boolean {
	if (!browser) return false;
	if (theme.mode !== 'system') return theme.mode === 'dark';
	return window.matchMedia('(prefers-color-scheme: dark)').matches;
}

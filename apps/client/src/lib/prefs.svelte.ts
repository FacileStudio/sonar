import { browser } from '$app/environment';

const KEY = 'sonar-suggest';

function stored(): boolean {
	if (!browser) return true;
	return localStorage.getItem(KEY) !== 'off';
}

/** Per-browser preferences that never reach the server. */
export const prefs = $state({ suggest: stored() });

/** Turns typing suggestions on or off in this browser. */
export function setSuggest(on: boolean) {
	prefs.suggest = on;
	if (browser) localStorage.setItem(KEY, on ? 'on' : 'off');
}

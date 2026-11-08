/**
 * The result a key press moves to: ArrowDown and j go forward, ArrowUp and k go back, Home and End
 * jump to the ends. -1 is the search box. Any other key returns null.
 */
export function nextIndex(key: string, current: number, count: number): number | null {
	if (count === 0) return null;
	switch (key) {
		case 'ArrowDown':
		case 'j':
			return Math.min(current + 1, count - 1);
		case 'ArrowUp':
		case 'k':
			return current <= 0 ? -1 : current - 1;
		case 'Home':
			return 0;
		case 'End':
			return count - 1;
	}
	return null;
}

/** Whether a key press comes from a place where typing must win over shortcuts. */
export function isTyping(target: EventTarget | null): boolean {
	if (!(target instanceof HTMLElement)) return false;
	return target.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName);
}

/** Splits an external !bang from the query the way the server does, for the redirect hint only. */
export function hasBang(query: string): boolean {
	return query.split(/\s+/).some((word) => word.length > 1 && word.startsWith('!'));
}

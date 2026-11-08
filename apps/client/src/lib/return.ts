/** Takes the page a visitor was on before signing in, once. Only same-origin paths come back. */
export function takeReturn(): string | null {
	try {
		const path = sessionStorage.getItem('sonar-return');
		sessionStorage.removeItem('sonar-return');
		return path && path.startsWith('/') && !path.startsWith('//') ? path : null;
	} catch {
		return null;
	}
}

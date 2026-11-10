/** A failed API call: the HTTP status and the decoded error body, if any. */
export class ApiError extends Error {
	constructor(
		readonly status: number,
		readonly body: unknown,
		message: string
	) {
		super(message);
	}
}

export type Kind = 'web' | 'images' | 'news';
export type WebResult = { title: string; url: string; snippet: string; engine: string };
export type ImageResult = { title: string; url: string; thumbnail: string; source: string; width?: number; height?: number };
export type NewsResult = { title: string; url: string; snippet: string; source: string; age?: string; thumbnail?: string };
export type SearchResponse = {
	query: string;
	kind: Kind;
	redirect?: string;
	own_keys: boolean;
	web?: WebResult[];
	images?: ImageResult[];
	news?: NewsResult[];
};
export type Usage = { used: number; limit: number; resets_at: string; own_keys: string[]; ai: boolean };
export type Features = { sign_in: boolean; summary: boolean; keys: boolean; suggest: boolean };
export type SessionUser = { email: string; name?: string; avatar_url?: string };
export type SavedKey = { engine: string; hint: string; updated_at: string };
export type Bang = { trigger: string; name: string; kind?: string; engine?: string; url?: string };

async function call<T>(path: string, init: RequestInit = {}): Promise<T> {
	const headers = new Headers(init.headers);
	if (init.method && init.method !== 'GET') headers.set('X-Facile-CSRF', '1');
	if (init.body) headers.set('Content-Type', 'application/json');
	const response = await fetch(`/api${path}`, { ...init, headers, credentials: 'same-origin' });
	const text = await response.text();
	const body: unknown = text ? JSON.parse(text) : undefined;
	if (!response.ok) {
		const message = (body as { error?: { message?: string } } | undefined)?.error?.message;
		throw new ApiError(response.status, body, message || `Request failed with status ${response.status}`);
	}
	return body as T;
}

export const api = {
	session: () => call<{ user: SessionUser | null; features: Features }>('/auth/session'),
	logout: () => call<void>('/auth/logout', { method: 'POST' }),
	search: (q: string, kind: Kind, signal?: AbortSignal) =>
		call<SearchResponse>(`/search?${new URLSearchParams({ q, type: kind })}`, { signal }),
	usage: () => call<Usage>('/usage'),
	suggest: (q: string, signal?: AbortSignal) => call<[string, string[]]>(`/suggest?${new URLSearchParams({ q })}`, { signal }),
	bangs: () => call<{ bangs: Bang[] }>('/bangs'),
	keys: () => call<{ enabled: boolean; engines: string[]; keys: SavedKey[] }>('/keys'),
	saveKey: (engine: string, key: string) =>
		call<void>(`/keys/${encodeURIComponent(engine)}`, { method: 'PUT', body: JSON.stringify({ key }) }),
	removeKey: (engine: string) => call<void>(`/keys/${encodeURIComponent(engine)}`, { method: 'DELETE' }),
	summary: (query: string) =>
		call<{ summary: string; model?: string; sources?: { title: string; url: string }[] }>('/summary', {
			method: 'POST',
			body: JSON.stringify({ query })
		})
};

/** The sign-in entry point. */
export const SIGN_IN_HREF = '/api/auth/oidc';

import { api, type Features, type SessionUser } from './api';

/** The signed-in person and the features this deployment runs, loaded once per page load. */
export const session = $state<{ loaded: boolean; user: SessionUser | null; features: Features }>({
	loaded: false,
	user: null,
	features: { sign_in: false, summary: false, keys: false, suggest: false }
});

/** Loads the session; a failure leaves the visitor signed out. */
export async function loadSession() {
	try {
		const answer = await api.session();
		session.user = answer.user;
		session.features = answer.features;
	} catch {
		session.user = null;
	}
	session.loaded = true;
}

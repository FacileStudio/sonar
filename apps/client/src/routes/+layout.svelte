<script lang="ts">
	import '../app.css';
	import { browser } from '$app/environment';
	import { provideLocale, Toaster } from '@facile/muse';
	import { registerAppIcons } from '$lib/icons';
	import { loadSession } from '$lib/session.svelte';
	import { theme } from '$lib/theme.svelte';
	import { bindLocale, localeSource } from '$lib/i18n/index.svelte';

	let { children } = $props();

	const i18n = provideLocale(localeSource);
	bindLocale(() => i18n.locale);
	registerAppIcons();

	if (browser) {
		theme.restore();
		loadSession();
	}
</script>

{@render children()}
<Toaster />

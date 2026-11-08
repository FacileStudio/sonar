<script lang="ts">
	import {
		PreferencesSection,
		ProfileSection,
		SettingsLayout,
		SettingsRow,
		SignInWithFacile,
		SubscriptionSection,
		Switch,
		settingsRoute
	} from '@facile/muse';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import KeysPanel from '$lib/components/KeysPanel.svelte';
	import Shell from '$lib/components/Shell.svelte';
	import { api, SIGN_IN_HREF, type Usage } from '$lib/api';
	import { prefs, setSuggest } from '$lib/prefs.svelte';
	import { session } from '$lib/session.svelte';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';
	import { setTheme, theme, type ThemeMode } from '$lib/theme.svelte';

	const sections = $derived([
		{ id: 'profile', label: t('settings.profile'), href: '/settings/profile' },
		{ id: 'preferences', label: t('settings.preferences'), href: '/settings/preferences' },
		{ id: 'keys', label: t('settings.keys'), href: '/settings/keys' },
		{ id: 'subscription', label: t('settings.subscription'), href: '/settings/subscription' }
	]);
	const route = $derived(settingsRoute(page.params.section, sections));

	let usage = $state<Usage | null>(null);

	$effect(() => {
		if (!route.known) goto('/settings', { replaceState: true });
	});

	$effect(() => {
		if (session.user) api.usage().then((u) => (usage = u)).catch(() => undefined);
	});
</script>

<svelte:head><title>{t('settings.title')} — Sonar</title></svelte:head>

<Shell>
	{#if session.loaded && !session.user}
		<div class="flex justify-center py-16"><SignInWithFacile href={SIGN_IN_HREF} label={t('shell.signIn')} /></div>
	{:else if session.user}
		<SettingsLayout {sections} active={page.params.section} backHref="/settings" title={t('settings.title')}>
			{#snippet children(current)}
				{#if current === 'profile'}
					<ProfileSection
						name={session.user?.name || session.user?.email || ''}
						email={session.user?.email}
						avatar={session.user?.avatar_url}
						manageHref="https://sso.facile.studio/account"
					/>
				{:else if current === 'preferences'}
					<PreferencesSection
						theme={theme.mode}
						onThemeChange={(next) => setTheme(next as ThemeMode)}
						onLocaleChange={rememberLocale}
					>
						{#if session.features.suggest}
							<SettingsRow label={t('settings.suggestions')} description={t('settings.suggestionsHelp')}>
								<Switch checked={prefs.suggest} onchange={(e) => setSuggest(e.currentTarget.checked)} aria-label={t('settings.suggestions')} />
							</SettingsRow>
						{/if}
					</PreferencesSection>
				{:else if current === 'keys'}
					<KeysPanel />
				{:else if current === 'subscription'}
					<SubscriptionSection
						plan={usage?.ai ? t('settings.planPaid') : t('settings.planFree')}
						manageHref="https://sso.facile.studio/spaces"
						upgradeHref={usage?.ai ? undefined : 'https://sso.facile.studio/spaces'}
						usage={usage && usage.limit >= 0 ? [{ label: t('settings.searchesToday'), used: usage.used, limit: usage.limit }] : []}
					/>
				{/if}
			{/snippet}
		</SettingsLayout>
	{/if}
</Shell>

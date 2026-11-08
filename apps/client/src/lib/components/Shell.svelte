<script lang="ts">
	import type { Snippet } from 'svelte';
	import { LandingFooter, LocaleSwitch, SignInWithFacile, UserMenu } from '@facile/muse';
	import { Icon } from '@facile/muse';
	import { api, SIGN_IN_HREF } from '$lib/api';
	import { SONAR_ICON } from '$lib/icons';
	import { session } from '$lib/session.svelte';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';
	import { setTheme, theme, type ThemeMode } from '$lib/theme.svelte';

	let { search, children, wide = false }: { search?: Snippet; children: Snippet; wide?: boolean } = $props();

	async function signOut() {
		await api.logout().catch(() => undefined);
		session.user = null;
		location.assign('/');
	}

	const menu = $derived({
		subscription: { href: 'https://sso.facile.studio/spaces' },
		settingsHref: '/settings',
		helpHref: 'https://github.com/FacileStudio/sonar/blob/main/docs/hosted.md',
		onSignOut: signOut,
		onLocaleChange: rememberLocale,
		theme: theme.mode,
		onThemeChange: (next: string) => setTheme(next as ThemeMode)
	});

	const columns = $derived([
		{
			title: 'Sonar',
			links: [
				{ href: '/bangs', label: t('shell.bangs') },
				{ href: '/terms', label: t('shell.terms') },
				{ href: '/docs', label: t('shell.api'), external: true }
			]
		},
		{ title: 'Facile', links: [{ href: 'https://facile.studio', label: t('shell.facile'), external: true }] }
	]);
</script>

<a href="#main" class="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50">{t('search.results')}</a>
<div class="flex min-h-[100dvh] flex-col bg-fc-page text-fc-fg">
	<header class="sticky top-0 z-30 bg-fc-page/90 backdrop-blur">
		<div class="mx-auto flex w-full max-w-6xl items-center gap-3 px-4 py-3 sm:gap-4 sm:px-6">
			<a href="/" class="flex shrink-0 items-center gap-2 rounded-fc-pill font-fc-title text-fc-md" aria-label={t('shell.home')}>
				<Icon icon={SONAR_ICON} size={26} />
				<span class="hidden sm:inline">Sonar</span>
			</a>
			<div class="min-w-0 flex-1">{@render search?.()}</div>
			<div class="flex shrink-0 items-center gap-2">
				{#if session.user}
					<UserMenu
						trigger="avatar"
						user={{ name: session.user.name?.trim() || session.user.email, email: session.user.email, avatar: session.user.avatar_url }}
						{...menu}
					/>
				{:else if session.loaded}
					<LocaleSwitch onChange={rememberLocale} />
					{#if session.features.sign_in}
						<SignInWithFacile href={SIGN_IN_HREF} size="sm" label={t('shell.signIn')} class="hidden sm:inline-flex" />
					{/if}
				{/if}
			</div>
		</div>
	</header>

	<main id="main" class="mx-auto flex w-full flex-1 flex-col px-4 py-6 sm:px-6 {wide ? 'max-w-6xl' : 'max-w-3xl'}">
		{@render children()}
	</main>

	<LandingFooter
		appName="Sonar"
		icon={SONAR_ICON}
		tagline={t('shell.tagline')}
		{columns}
		legal={{ hideLocaleSwitch: true }}
		year={new Date().getFullYear()}
	/>
</div>

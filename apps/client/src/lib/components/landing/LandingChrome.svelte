<script lang="ts">
	import type { Snippet } from 'svelte';
	import { LandingFooter, LandingHeader, LocaleSwitch, SignInWithFacile } from '@facile/muse';
	import { SIGN_IN_HREF } from '$lib/api';
	import { SONAR_ICON } from '$lib/icons';
	import { session } from '$lib/session.svelte';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';

	let { links, children }: { links: { href: string; label: string }[]; children: Snippet } = $props();

	const columns = $derived([
		{
			title: 'Sonar',
			links: [
				{ href: '/downloads', label: t('landing.downloadsTitle') },
				{ href: '/bangs', label: t('shell.bangs') },
				{ href: '/terms', label: t('shell.terms') },
				{ href: '/docs', label: t('shell.api'), external: true }
			]
		},
		{ title: 'Facile', links: [{ href: 'https://facile.studio', label: t('shell.facile'), external: true }] }
	]);
</script>

<div class="min-h-dvh bg-fc-page text-fc-fg">
	<LandingHeader appName="Sonar" icon={SONAR_ICON} {links} skipHref="#main">
		{#snippet actions()}
			<LocaleSwitch onChange={rememberLocale} />
			{#if session.features.sign_in}
				<SignInWithFacile href={SIGN_IN_HREF} size="sm" label={t('shell.signIn')} />
			{/if}
		{/snippet}
	</LandingHeader>

	<main id="main" class="flex flex-col gap-24 pb-24">
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

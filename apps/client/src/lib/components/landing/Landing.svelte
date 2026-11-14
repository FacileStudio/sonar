<script lang="ts">
	import { Button, DownloadSection, Icon, LandingFooter, LandingHeader, LocaleSwitch, SignInWithFacile, icons } from '@facile/muse';
	import { SIGN_IN_HREF } from '$lib/api';
	import { SONAR_ICON } from '$lib/icons';
	import { session } from '$lib/session.svelte';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';
	import FetchDemo from './FetchDemo.svelte';
	import SearchDemo from './SearchDemo.svelte';
	import { INSTALL, RELEASES } from './install';

	const links = $derived([
		{ href: '#features', label: t('landing.navFeatures') },
		{ href: '#cli', label: t('landing.navCli') }
	]);

	const features = $derived([
		{ icon: icons.layers, title: t('landing.enginesTitle'), body: t('landing.enginesBody') },
		{ icon: icons.shield, title: t('landing.privacyTitle'), body: t('shell.privacyLine') },
		{ icon: icons.bolt, title: t('landing.bangsTitle'), body: t('landing.bangsBody') },
		{ icon: icons.key, title: t('landing.keysTitle'), body: t('landing.keysBody') }
	]);

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
		<section class="mx-auto flex w-full max-w-fc-lg flex-col items-center gap-10 px-4 pt-12 sm:px-6 md:pt-20">
			<div class="flex max-w-fc-md flex-col items-center gap-6 text-center">
				<h1 class="font-fc-title text-fc-3xl font-bold text-balance">{t('landing.title')}</h1>
				<p class="text-fc-md text-balance text-fc-fg-muted">{t('landing.lead')}</p>
				<div class="flex flex-wrap items-center justify-center gap-3">
					{#if session.features.sign_in}
						<SignInWithFacile href={SIGN_IN_HREF} size="lg" label={t('shell.signIn')} />
					{/if}
					<Button href="#cli" variant="secondary" size="lg" icon={icons.code}>{t('landing.installCli')}</Button>
				</div>
			</div>
			<div class="w-full"><SearchDemo /></div>
		</section>

		<section id="features" aria-labelledby="features-title" class="mx-auto flex w-full max-w-fc-lg flex-col gap-8 px-4 sm:px-6">
			<h2 id="features-title" class="font-fc-title text-fc-2xl font-bold">{t('landing.featuresTitle')}</h2>
			<ul class="grid gap-4 sm:grid-cols-2">
				{#each features as feature (feature.icon)}
					<li class="flex flex-col gap-3 rounded-fc-lg bg-fc-component p-6">
						<span class="flex size-10 items-center justify-center rounded-fc-md bg-fc-surface"><Icon icon={feature.icon} size={20} /></span>
						<h3 class="text-fc-md font-bold">{feature.title}</h3>
						<p class="text-fc-sm text-fc-fg-muted">{feature.body}</p>
					</li>
				{/each}
			</ul>
		</section>

		<section id="cli" aria-labelledby="cli-title" class="mx-auto flex w-full max-w-fc-lg flex-col gap-8 px-4 sm:px-6">
			<div class="flex flex-col gap-2">
				<h2 id="cli-title" class="font-fc-title text-fc-2xl font-bold">{t('landing.cliTitle')}</h2>
				<p class="max-w-fc-md text-fc-md text-fc-fg-muted">{t('landing.cliLead')}</p>
			</div>
			<div class="grid items-center gap-6 lg:grid-cols-[minmax(0,5fr)_minmax(0,6fr)]">
				<div class="w-full max-w-[440px]"><FetchDemo /></div>
				<DownloadSection commands={INSTALL} releasesHref={RELEASES} headingLevel={3} />
			</div>
		</section>

		{#if session.features.sign_in}
			<section class="mx-auto flex w-full max-w-fc-md flex-col items-center gap-6 px-4 text-center sm:px-6">
				<h2 class="font-fc-title text-fc-2xl font-bold">{t('landing.closingTitle')}</h2>
				<p class="text-fc-md text-balance text-fc-fg-muted">{t('shell.signInLead', { count: 20 })}</p>
				<SignInWithFacile href={SIGN_IN_HREF} size="lg" label={t('shell.signIn')} />
			</section>
		{/if}
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

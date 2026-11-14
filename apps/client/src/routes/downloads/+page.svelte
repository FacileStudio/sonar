<script lang="ts">
	import { DownloadSection } from '@facile/muse';
	import LandingChrome from '$lib/components/landing/LandingChrome.svelte';
	import { INSTALL, RELEASES } from '$lib/components/landing/install';
	import { t } from '$lib/i18n/index.svelte';

	const links = $derived([
		{ href: '/#features', label: t('landing.navFeatures') },
		{ href: '/#cli', label: t('landing.navCli') }
	]);
</script>

<svelte:head><title>{t('landing.downloadsTitle')} · Sonar</title></svelte:head>

<LandingChrome {links}>
	<section aria-labelledby="downloads-title" class="mx-auto flex w-full max-w-fc-lg flex-col gap-8 px-4 pt-12 sm:px-6 md:pt-20">
		<div class="flex flex-col gap-2">
			<h1 id="downloads-title" class="font-fc-title text-fc-3xl font-bold text-balance">{t('landing.downloadsTitle')}</h1>
			<p class="max-w-fc-md text-fc-md text-fc-fg-muted">{t('landing.downloadsLead')}</p>
		</div>
		<DownloadSection commands={INSTALL} releasesHref={RELEASES} />
	</section>
</LandingChrome>

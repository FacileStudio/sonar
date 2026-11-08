<script lang="ts">
	import { Icon } from '@facile/muse';
	import { goto } from '$app/navigation';
	import Shell from '$lib/components/Shell.svelte';
	import SearchBox from '$lib/components/SearchBox.svelte';
	import SignInCard from '$lib/components/SignInCard.svelte';
	import { SONAR_ICON } from '$lib/icons';
	import { takeReturn } from '$lib/return';
	import { session } from '$lib/session.svelte';
	import { t } from '$lib/i18n/index.svelte';

	$effect(() => {
		if (!session.loaded || !session.user) return;
		const back = takeReturn();
		if (back && back !== '/') goto(back, { replaceState: true });
	});
</script>

<svelte:head><title>Sonar — {t('shell.tagline')}</title></svelte:head>

<Shell>
	<section class="flex flex-1 flex-col items-center justify-center gap-8 py-10 sm:py-20">
		<div class="flex flex-col items-center gap-3 text-center">
			<Icon icon={SONAR_ICON} size={56} />
			<h1 class="font-fc-title text-fc-3xl">Sonar</h1>
			<p class="max-w-md text-fc-sm text-fc-fg-muted">{t('shell.tagline')}</p>
		</div>
		<div class="w-full max-w-2xl">
			<SearchBox large autofocus={!!session.user} />
		</div>
		{#if session.loaded && !session.user && session.features.sign_in}
			<div class="w-full max-w-2xl"><SignInCard /></div>
		{/if}
		<p class="max-w-lg text-center text-fc-xs text-fc-fg-muted">{t('shell.privacyLine')}</p>
	</section>
</Shell>

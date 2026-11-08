<script lang="ts">
	import { goto } from '$app/navigation';
	import Shell from '$lib/components/Shell.svelte';
	import SearchBox from '$lib/components/SearchBox.svelte';
	import SignInCard from '$lib/components/SignInCard.svelte';
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
		<h1 class="sr-only">Sonar</h1>
		<div class="flex w-full max-w-2xl flex-col items-center gap-4">
			<div class="w-full">
				<SearchBox large autofocus={!!session.user} />
			</div>
			<p class="max-w-md text-center text-fc-sm text-fc-fg-muted">{t('shell.tagline')}</p>
		</div>
		{#if session.loaded && !session.user && session.features.sign_in}
			<div class="w-full max-w-2xl"><SignInCard /></div>
		{/if}
	</section>
</Shell>

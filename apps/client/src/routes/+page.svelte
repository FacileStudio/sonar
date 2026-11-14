<script lang="ts">
	import { goto } from '$app/navigation';
	import Landing from '$lib/components/landing/Landing.svelte';
	import Shell from '$lib/components/Shell.svelte';
	import SearchBox from '$lib/components/SearchBox.svelte';
	import { takeReturn } from '$lib/return';
	import { session } from '$lib/session.svelte';
	import { t } from '$lib/i18n/index.svelte';

	$effect(() => {
		if (!session.loaded || !session.user) return;
		const back = takeReturn();
		if (back && back !== '/') goto(back, { replaceState: true });
	});
</script>

<svelte:head><title>sonar · Multi-engine web search</title></svelte:head>

{#if session.user}
	<Shell>
		<section class="flex flex-1 flex-col items-center justify-center gap-8 py-10 sm:py-20">
			<h1 class="sr-only">Sonar</h1>
			<div class="flex w-full max-w-2xl flex-col items-center gap-4">
				<div class="w-full">
					<SearchBox large autofocus />
				</div>
				<p class="max-w-md text-center text-fc-sm text-fc-fg-muted">{t('shell.tagline')}</p>
			</div>
		</section>
	</Shell>
{:else if session.loaded}
	<Landing />
{:else}
	<div class="min-h-dvh bg-fc-page"></div>
{/if}

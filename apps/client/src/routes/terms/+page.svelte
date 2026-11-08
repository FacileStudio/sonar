<script lang="ts">
	import { Alert } from '@facile/muse';
	import Shell from '$lib/components/Shell.svelte';
	import { termsFor } from '$lib/legal/terms';
	import { locale, t } from '$lib/i18n/index.svelte';

	const terms = $derived(termsFor(locale()));
	const note = $derived(t('terms.languageNote'));
</script>

<svelte:head><title>{terms.title}</title></svelte:head>

<Shell>
	<article class="flex flex-col gap-6">
		<Alert tone="warning">{t('terms.draft')}</Alert>
		{#if note}<p class="text-fc-sm text-fc-fg-muted">{note}</p>{/if}
		<header class="flex flex-col gap-1">
			<h1 class="font-fc-title text-fc-2xl">{terms.title}</h1>
			<p class="text-fc-xs text-fc-fg-muted">{terms.updated}</p>
		</header>
		{#each terms.sections as section (section.heading)}
			<section class="flex flex-col gap-2">
				<h2 class="text-fc-md font-medium">{section.heading}</h2>
				{#each section.body as paragraph (paragraph)}<p class="text-fc-sm leading-relaxed">{paragraph}</p>{/each}
			</section>
		{/each}
	</article>
</Shell>

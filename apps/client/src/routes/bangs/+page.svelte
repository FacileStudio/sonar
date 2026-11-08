<script lang="ts">
	import { Alert, Card } from '@facile/muse';
	import Shell from '$lib/components/Shell.svelte';
	import { api, type Bang } from '$lib/api';
	import { t } from '$lib/i18n/index.svelte';

	let bangs = $state<Bang[]>([]);
	$effect(() => {
		api.bangs().then((answer) => (bangs = answer.bangs)).catch(() => undefined);
	});
	const internal = $derived(bangs.filter((b) => !b.url));
	const external = $derived(bangs.filter((b) => b.url));
</script>

<svelte:head><title>{t('bangs.title')} — Sonar</title></svelte:head>

{#snippet list(items: Bang[])}
	<ul class="grid gap-2 sm:grid-cols-2">
		{#each items as bang (bang.trigger)}
			<li class="flex items-center gap-3 rounded-fc-md bg-fc-component px-3 py-2">
				<code class="shrink-0 font-fc-mono text-fc-sm">!{bang.trigger}</code>
				<span class="truncate text-fc-sm text-fc-fg-muted">{bang.name}</span>
			</li>
		{/each}
	</ul>
{/snippet}

<Shell>
	<div class="flex flex-col gap-6">
		<h1 class="font-fc-title text-fc-2xl">{t('bangs.title')}</h1>
		<p class="text-fc-sm text-fc-fg-muted">{t('bangs.lead')}</p>
		<Card class="flex flex-col gap-3 p-4">
			<h2 class="font-medium">{t('bangs.internal')}</h2>
			{@render list(internal)}
		</Card>
		<Card class="flex flex-col gap-3 p-4">
			<h2 class="font-medium">{t('bangs.external')}</h2>
			<Alert tone="warning">{t('bangs.privacy')}</Alert>
			{@render list(external)}
		</Card>
	</div>
</Shell>

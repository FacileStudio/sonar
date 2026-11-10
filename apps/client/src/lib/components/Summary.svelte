<script lang="ts">
	import { Button, Card, PlanLimit, planLimitFrom, type PlanLimitError } from '@facile/muse';
	import { api, ApiError } from '$lib/api';
	import { SPARK_ICON } from '$lib/icons';
	import { t } from '$lib/i18n/index.svelte';

	let { query }: { query: string } = $props();

	let text = $state('');
	let sources = $state<{ title: string; url: string }[]>([]);
	let busy = $state(false);
	let failed = $state(false);
	let limit = $state<PlanLimitError | null>(null);

	$effect(() => {
		query;
		text = '';
		sources = [];
		failed = false;
	});

	async function summarize() {
		busy = true;
		failed = false;
		try {
			const answer = await api.summary(query);
			sources = answer.sources ?? [];
			text = answer.summary;
		} catch (error) {
			if (error instanceof ApiError && error.status === 402) limit = planLimitFrom(error.body);
			else failed = true;
		} finally {
			busy = false;
		}
	}
</script>

{#if text}
	<Card class="flex flex-col gap-2 p-4">
		<h2 class="text-fc-sm font-medium">{t('search.summary')}</h2>
		<p class="whitespace-pre-line text-fc-sm">{text}</p>
		{#if sources.length}
			<ol class="flex list-decimal flex-col gap-1 pl-5 text-fc-xs">
				{#each sources as source (source.url)}
					<li><a class="text-fc-accent hover:underline" href={source.url} rel="noopener noreferrer" target="_blank">{source.title || source.url}</a></li>
				{/each}
			</ol>
		{/if}
		<p class="text-fc-xs text-fc-fg-muted">{t('search.summaryNote')}</p>
	</Card>
{:else}
	<div class="flex items-center gap-3">
		<Button variant="secondary" size="sm" icon={SPARK_ICON} onclick={summarize} disabled={busy}>{t('search.summarize')}</Button>
		{#if failed}<span class="text-fc-xs text-fc-danger">{t('search.summaryFailed')}</span>{/if}
	</div>
{/if}
<PlanLimit error={limit} variant="dialog" onClose={() => (limit = null)} />

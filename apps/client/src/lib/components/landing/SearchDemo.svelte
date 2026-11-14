<script lang="ts">
	import { fade } from 'svelte/transition';
	import { Badge, Icon, icons } from '@facile/muse';
	import { Bar, Roll, Stage, Window, grow, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';

	const QUERY = 'circuit breaker pattern';
	const RESULTS = [
		{ host: 'martinfowler.com', title: 'Circuit Breaker', engine: 'brave' },
		{ host: 'learn.microsoft.com', title: 'Circuit Breaker pattern', engine: 'tavily' },
		{ host: 'en.wikipedia.org', title: 'Circuit breaker design pattern', engine: 'exa' }
	];

	let draft = $state('');
	let shown = $state(0);
	let status = $state('');

	const tabs = $derived([t('search.tabWeb'), t('search.tabImages'), t('search.tabNews')]);
	const done = $derived(t('landing.engines', { count: RESULTS.length }));

	function reset() {
		draft = '';
		shown = 0;
		status = '';
	}

	function poster() {
		draft = QUERY;
		shown = RESULTS.length + 1;
		status = done;
	}

	async function script(act: Act) {
		await act.wait(700);
		await act.type(QUERY, (v) => (draft = v), 55);
		await act.wait(350);
		for (const [i, result] of RESULTS.entries()) {
			status = result.engine;
			shown = i + 1;
			await act.wait(520);
		}
		shown = RESULTS.length + 1;
		status = done;
		await act.wait(2400);
		reset();
		await act.wait(700);
	}
</script>

<Stage
	width={960}
	height={580}
	compact={{ below: 640, width: 358, height: 520 }}
	label={t('landing.searchDemo')}
	{script}
	{reset}
	{poster}
>
	{#snippet children({ compact })}
		<Window style={compact ? 'top:16px;left:16px;width:326px;height:488px' : 'top:24px;left:48px;width:864px;height:532px'}>
			{#snippet trailing()}
				{#if status}
					<span class="sd-status" transition:fade={{ duration: 200 }}><Roll value={status} /></span>
				{/if}
			{/snippet}
			<div class="sd-body" class:sd-compact={compact}>
				<div class="sd-search">
					<Icon icon={icons.search} size={compact ? 18 : 20} class="shrink-0 text-fc-fg-muted" />
					<span class:sd-placeholder={!draft}>{draft || t('search.placeholder')}</span>
				</div>
				<div class="sd-tabs">
					{#each tabs as tab, i (i)}
						<span class:sd-active={i === 0}>{tab}</span>
					{/each}
				</div>
				<ol class="sd-results">
					{#each RESULTS as result, i (result.host)}
						{#if shown > i}
							<li in:grow out:fade={{ duration: 200 }}>
								<div class="sd-row">
									<span class="sd-meta"><span>{result.host}</span><Badge class="shrink-0 whitespace-nowrap">{result.engine}</Badge></span>
									<span class="sd-title">{result.title}</span>
									<Bar w={['82%', '74%', '88%'][i]} />
									{#if !compact}<Bar w={['56%', '48%', '62%'][i]} />{/if}
								</div>
							</li>
						{/if}
					{/each}
					{#if shown > RESULTS.length && !compact}
						<li in:grow out:fade={{ duration: 200 }}>
							<div class="sd-row">
								<Bar w="22%" />
								<Bar w="38%" h={12} />
							</div>
						</li>
					{/if}
				</ol>
			</div>
		</Window>
	{/snippet}
</Stage>

<style>
	.sd-body {
		display: flex;
		flex: 1;
		flex-direction: column;
		min-height: 0;
		padding: 24px 40px 0;
		overflow: hidden;
		font-family: var(--font-fc-body);
		color: var(--color-fc-fg);
	}
	.sd-compact {
		padding: 16px 16px 0;
	}
	.sd-search {
		display: flex;
		align-items: center;
		gap: 12px;
		height: 48px;
		padding: 0 20px;
		border-radius: 999px;
		background: var(--color-fc-field);
		font-size: 16px;
	}
	.sd-compact .sd-search {
		height: 44px;
		padding: 0 14px;
		font-size: 14px;
	}
	.sd-search > span {
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}
	.sd-placeholder {
		color: var(--color-fc-fg-muted);
	}
	.sd-tabs {
		display: flex;
		gap: 4px;
		margin: 16px 0 8px;
		font-size: 13px;
		color: var(--color-fc-fg-muted);
	}
	.sd-tabs span {
		padding: 6px 12px;
		border-radius: 999px;
		white-space: nowrap;
	}
	.sd-tabs .sd-active {
		background: var(--color-fc-surface);
		color: var(--color-fc-fg);
		font-weight: 700;
	}
	.sd-results {
		margin: 0;
		padding: 0;
		list-style: none;
	}
	.sd-row {
		display: grid;
		gap: 8px;
		padding: 12px;
	}
	.sd-compact .sd-row {
		padding: 10px 4px;
	}
	.sd-meta {
		display: flex;
		align-items: center;
		gap: 8px;
		min-width: 0;
		font-size: 12px;
		color: var(--color-fc-fg-muted);
	}
	.sd-meta > span:first-child {
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}
	.sd-title {
		font-size: 17px;
		font-weight: 700;
		white-space: nowrap;
	}
	.sd-compact .sd-title {
		font-size: 15px;
	}
	.sd-status {
		display: inline-flex;
		min-width: 96px;
		justify-content: flex-end;
		font-size: 12px;
		color: var(--color-fc-fg-muted);
	}
</style>

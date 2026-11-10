<script lang="ts">
	import { Alert, Badge, Button, EmptyState, PlanLimit, planLimitFrom, Skeleton, Tabs, type PlanLimitError } from '@facile/muse';
	import { page } from '$app/state';
	import Results from '$lib/components/Results.svelte';
	import SearchBox from '$lib/components/SearchBox.svelte';
	import Shell from '$lib/components/Shell.svelte';
	import SignInCard from '$lib/components/SignInCard.svelte';
	import Summary from '$lib/components/Summary.svelte';
	import { api, ApiError, type Kind, type SearchResponse, type Usage } from '$lib/api';
	import { IMAGES_ICON, NEWS_ICON } from '$lib/icons';
	import { isTyping, nextIndex } from '$lib/keyboard';
	import { session } from '$lib/session.svelte';
	import { t } from '$lib/i18n/index.svelte';

	const q = $derived(page.url.searchParams.get('q') ?? '');
	const kind = $derived(toKind(page.url.searchParams.get('type')));

	let response = $state<SearchResponse | null>(null);
	let usage = $state<Usage | null>(null);
	let loading = $state(false);
	let failed = $state('');
	let limit = $state<PlanLimitError | null>(null);
	let focused = -1;

	function toKind(value: string | null): Kind {
		return value === 'images' || value === 'news' ? value : 'web';
	}

	const tabs = $derived([
		{ id: 'web', label: t('search.tabWeb'), href: link('web') },
		{ id: 'images', label: t('search.tabImages'), icon: IMAGES_ICON, href: link('images') },
		{ id: 'news', label: t('search.tabNews'), icon: NEWS_ICON, href: link('news') }
	]);

	function link(to: Kind): string {
		return `/search?${new URLSearchParams({ q, type: to })}`;
	}

	$effect(() => {
		if (!session.loaded || !session.user || !q.trim()) return;
		const controller = new AbortController();
		run(q, kind, controller.signal);
		return () => controller.abort();
	});

	async function run(text: string, tab: Kind, signal: AbortSignal) {
		loading = true;
		failed = '';
		limit = null;
		focused = -1;
		try {
			const answer = await api.search(text, tab, signal);
			if (answer.redirect) {
				location.replace(answer.redirect);
				return;
			}
			response = answer;
		} catch (error) {
			if (signal.aborted) return;
			response = null;
			if (error instanceof ApiError && error.status === 402) limit = planLimitFrom(error.body);
			else failed = error instanceof Error ? error.message : t('search.failed');
		} finally {
			if (!signal.aborted) loading = false;
		}
		usage = await api.usage().catch(() => usage);
	}

	function onKey(event: KeyboardEvent) {
		if (event.altKey || event.ctrlKey || event.metaKey) return;
		const typing = isTyping(event.target);
		if (typing && event.key !== 'ArrowDown') return;
		if (typing && (event.target as HTMLElement).getAttribute('aria-expanded') === 'true') return;
		const links = [...document.querySelectorAll<HTMLAnchorElement>('[data-result]')];
		const next = nextIndex(event.key, focused, links.length);
		if (next === null) return;
		event.preventDefault();
		focused = next;
		if (next === -1) document.querySelector<HTMLInputElement>('input[type=search]')?.focus();
		else links[next].focus();
	}

	const empty = $derived(
		response !== null && !loading && !(response.web?.length || response.images?.length || response.news?.length)
	);
</script>

<svelte:window onkeydown={onKey} />
<svelte:head><title>{q ? `${q} — Sonar` : 'Sonar'}</title></svelte:head>

<Shell wide>
	{#snippet search()}
		<SearchBox value={q} {kind} />
	{/snippet}

	<div class="flex flex-col gap-5">
		<div class="flex flex-wrap items-center justify-between gap-3">
			<Tabs items={tabs} value={kind} label={t('search.tabs')} />
			{#if usage}
				<Badge class="whitespace-nowrap">
					{response?.own_keys ? t('search.ownKeys') : usage.limit < 0 ? t('search.unlimited') : t('search.usage', { used: usage.used, limit: usage.limit })}
				</Badge>
			{/if}
		</div>

		{#if session.loaded && !session.user}
			<SignInCard />
		{:else if limit}
			<PlanLimit error={limit} />
		{:else if failed}
			<Alert tone="danger" title={t('search.failed')}>
				<p>{failed}</p>
				<Button size="sm" variant="secondary" onclick={() => run(q, kind, new AbortController().signal)}>{t('search.retry')}</Button>
			</Alert>
		{:else if loading && !response}
			<div class="flex flex-col gap-4" aria-busy="true" aria-label={t('shell.loading')}>
				{#each [0, 1, 2, 3] as row (row)}<Skeleton class="h-20 w-full" />{/each}
			</div>
		{:else if empty}
			<EmptyState title={t('search.noResults', { q })} description={t('search.noResultsHint')} />
		{:else if response}
			<div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_18rem]" class:opacity-60={loading}>
				<section aria-label={t('search.results')} class="min-w-0">
					<Results web={response.web} images={response.images} news={response.news} />
				</section>
				<aside class="flex flex-col gap-4">
					{#if session.features.summary && response.web?.length}
						<Summary query={response.query} />
					{/if}
					<p class="hidden text-fc-xs text-fc-fg-muted lg:block">{t('search.keyboardHint')}</p>
					<p class="text-fc-xs text-fc-fg-muted">{t('shell.poweredBy')}</p>
				</aside>
			</div>
		{/if}
	</div>
</Shell>

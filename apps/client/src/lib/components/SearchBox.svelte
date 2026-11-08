<script lang="ts">
	import { Icon, icons } from '@facile/muse';
	import { goto } from '$app/navigation';
	import { api, type Kind } from '$lib/api';
	import { hasBang } from '$lib/keyboard';
	import { prefs } from '$lib/prefs.svelte';
	import { session } from '$lib/session.svelte';
	import { t } from '$lib/i18n/index.svelte';

	let {
		value = '',
		kind = 'web',
		autofocus = false,
		large = false
	}: { value?: string; kind?: Kind; autofocus?: boolean; large?: boolean } = $props();

	let query = $state('');
	let typed = '';
	let items = $state<string[]>([]);
	let active = $state(-1);
	let open = $state(false);
	let input: HTMLInputElement | undefined = $state();
	let pending: AbortController | null = null;
	let timer: ReturnType<typeof setTimeout> | undefined;
	const listId = $props.id();

	$effect(() => {
		query = value;
	});

	$effect(() => {
		if (autofocus) input?.focus();
	});

	function onWindowKey(event: KeyboardEvent) {
		if (event.key !== '/' || document.activeElement === input) return;
		const target = event.target as HTMLElement | null;
		if (target && ['INPUT', 'TEXTAREA'].includes(target.tagName)) return;
		event.preventDefault();
		input?.focus();
		input?.select();
	}

	function fetchSuggestions(text: string) {
		clearTimeout(timer);
		pending?.abort();
		if (!prefs.suggest || !session.features.suggest || !session.user || text.trim().length < 2 || hasBang(text)) {
			items = [];
			return;
		}
		timer = setTimeout(async () => {
			pending = new AbortController();
			try {
				const [, list] = await api.suggest(text, pending.signal);
				items = list;
				open = items.length > 0;
				active = -1;
			} catch {
				items = [];
			}
		}, 140);
	}

	function submit(text = query) {
		const q = text.trim();
		open = false;
		items = [];
		if (!q) return;
		const target = `/search?${new URLSearchParams({ q, type: kind })}`;
		if (hasBang(q)) {
			location.assign(target);
			return;
		}
		goto(target);
	}

	function onKey(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			open = false;
			return;
		}
		if (!open || items.length === 0) return;
		if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
			event.preventDefault();
			const step = event.key === 'ArrowDown' ? 1 : -1;
			const n = items.length;
			active = ((active + 1 + step + n + 1) % (n + 1)) - 1;
			query = active >= 0 ? items[active] : typed;
		}
	}
</script>

<svelte:window onkeydown={onWindowKey} />

<form
	role="search"
	class="relative w-full"
	onsubmit={(event) => {
		event.preventDefault();
		submit();
	}}
>
	<label for="{listId}-input" class="sr-only">{t('search.label')}</label>
	<div class="flex items-center rounded-fc-pill bg-fc-field focus-within:outline focus-within:outline-2 focus-within:outline-fc-ring {large ? 'h-14 pl-5.5 pr-4 gap-3.5' : 'h-11 pl-3.5 pr-3 gap-2.5'}">
		<Icon icon={icons.search} size={large ? 22 : 18} class="shrink-0 text-fc-fg-muted {large ? 'translate-y-0.5' : 'translate-y-px'}" />
		<input
			bind:this={input}
			id="{listId}-input"
			type="search"
			name="q"
			enterkeyhint="search"
			autocomplete="off"
			spellcheck="false"
			maxlength="400"
			placeholder={t('search.placeholder')}
			class="h-full min-w-0 flex-1 bg-transparent p-0 outline-none placeholder:text-fc-fg-muted [&::-webkit-search-cancel-button]:appearance-none {large ? 'text-fc-md' : 'text-fc-sm'}"
			role="combobox"
			aria-expanded={open}
			aria-controls="{listId}-list"
			aria-autocomplete="list"
			aria-activedescendant={active >= 0 ? `${listId}-opt-${active}` : undefined}
			bind:value={query}
			oninput={() => {
				typed = query;
				fetchSuggestions(query);
			}}
			onkeydown={onKey}
			onblur={() => setTimeout(() => (open = false), 120)}
		/>
		<button type="submit" class="sr-only">{t('search.submit')}</button>
	</div>
	{#if open && items.length > 0}
		<ul
			id="{listId}-list"
			role="listbox"
			aria-label={t('search.suggestions')}
			class="absolute inset-x-0 top-full z-40 mt-2 overflow-hidden rounded-fc-lg bg-fc-bg py-2 shadow-lg"
		>
			{#each items as item, i (item)}
				<li
					id="{listId}-opt-{i}"
					role="option"
					aria-selected={i === active}
					class="flex min-h-11 cursor-pointer items-center gap-3 px-4 text-fc-sm {i === active ? 'bg-fc-surface' : 'hover:bg-fc-surface'}"
					onmousedown={(event) => {
						event.preventDefault();
						submit(item);
					}}
				>
					<Icon icon={icons.search} size={14} class="shrink-0 text-fc-fg-muted" />
					<span class="truncate">{item}</span>
				</li>
			{/each}
		</ul>
	{/if}
</form>

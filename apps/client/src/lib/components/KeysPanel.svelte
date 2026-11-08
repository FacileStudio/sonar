<script lang="ts">
	import { Alert, Button, Icon, SecretField, SettingsRow, SettingsSection, toast } from '@facile/muse';
	import { api, type SavedKey } from '$lib/api';
	import { ENGINE_ICONS } from '$lib/icons';
	import { t } from '$lib/i18n/index.svelte';

	let enabled = $state(true);
	let engines = $state<string[]>([]);
	let saved = $state<Record<string, SavedKey>>({});
	let drafts = $state<Record<string, string>>({});
	let busy = $state('');

	const names: Record<string, string> = {
		brave: 'Brave Search', exa: 'Exa', firecrawl: 'Firecrawl', linkup: 'Linkup',
		parallel: 'Parallel', serpapi: 'SerpAPI', tavily: 'Tavily'
	};

	async function load() {
		const answer = await api.keys();
		enabled = answer.enabled;
		engines = answer.engines;
		for (const engine of answer.engines) drafts[engine] ??= '';
		saved = Object.fromEntries(answer.keys.map((key) => [key.engine, key]));
	}

	$effect(() => {
		load().catch(() => (enabled = false));
	});

	async function save(engine: string) {
		busy = engine;
		try {
			await api.saveKey(engine, drafts[engine] ?? '');
			drafts[engine] = '';
			await load();
		} catch (error) {
			toast.danger(error instanceof Error ? `${t('settings.keyFailed')}: ${error.message}` : t('settings.keyFailed'));
		} finally {
			busy = '';
		}
	}

	async function remove(engine: string) {
		busy = engine;
		await api.removeKey(engine).catch(() => undefined);
		await load().catch(() => undefined);
		busy = '';
	}
</script>

<SettingsSection title={t('settings.keys')} description={t('settings.keysHelp')}>
	{#if !enabled}
		<Alert tone="info">{t('settings.keysDisabled')}</Alert>
	{:else}
		{#each engines as engine (engine)}
			<SettingsRow stacked>
				<div class="flex w-full flex-col gap-3">
					<div class="flex min-w-0 flex-col gap-1">
						<label for="{engine}-key" class="flex items-center gap-2 text-fc-sm font-medium text-fc-fg">
							<Icon icon={ENGINE_ICONS[engine]} size={20} class="text-fc-fg shrink-0" />
							<span>{names[engine] ?? engine}</span>
						</label>
						<p class="text-fc-xs text-fc-fg-muted">
							{saved[engine] ? t('settings.keySaved', { hint: saved[engine].hint }) : t('settings.keyNone')}
						</p>
					</div>
					<div class="flex flex-col gap-2 sm:flex-row sm:items-end">
						<SecretField
							id="{engine}-key"
							for="{engine}-key"
							bind:value={drafts[engine]}
							editable
							placeholder="••••••••"
							class="min-w-0 flex-1"
						/>
						<div class="flex gap-2">
							<Button size="sm" onclick={() => save(engine)} disabled={busy === engine || !drafts[engine]}>{t('settings.keySave')}</Button>
							{#if saved[engine]}
								<Button size="sm" variant="ghost-danger" onclick={() => remove(engine)} disabled={busy === engine}>{t('settings.keyRemove')}</Button>
							{/if}
						</div>
					</div>
				</div>
			</SettingsRow>
		{/each}
	{/if}
</SettingsSection>

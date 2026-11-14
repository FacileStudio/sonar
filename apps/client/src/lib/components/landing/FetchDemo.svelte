<script lang="ts">
	import { Bar, Caret, Stage, Terminal, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';

	const COMMAND = 'sonar fetch https://martinfowler.com/bliki/CircuitBreaker.html';
	const PAGE = [
		{ kind: 'h1', text: '# Circuit Breaker' },
		{ kind: 'bar', w: '92%' },
		{ kind: 'bar', w: '84%' },
		{ kind: 'bar', w: '61%' },
		{ kind: 'h2', text: '## Basic Circuit Breaker' },
		{ kind: 'bar', w: '88%' },
		{ kind: 'bar', w: '47%' }
	] as const;

	let command = $state('');
	let shown = $state(0);

	function reset() {
		command = '';
		shown = 0;
	}

	function poster() {
		command = COMMAND;
		shown = PAGE.length;
	}

	async function script(act: Act) {
		await act.wait(600);
		await act.type(COMMAND, (v) => (command = v), 40);
		await act.wait(700);
		for (let i = 1; i <= PAGE.length; i++) {
			shown = i;
			await act.wait(140);
		}
		await act.wait(2600);
		reset();
		await act.wait(600);
	}
</script>

<Stage
	width={440}
	height={320}
	compact={{ below: 400, width: 358, height: 340 }}
	label={t('landing.fetchDemo')}
	{script}
	{reset}
	{poster}
>
	{#snippet children({ compact })}
		<Terminal class="fd-term" style={compact ? 'top:12px;left:12px;right:12px;bottom:12px' : 'top:20px;left:20px;right:20px;bottom:20px'}>
			<div class="fd-line fd-wrap">
				<span class="fd-dim">~ $</span>
				{command}{#if shown === 0}<Caret />{/if}
			</div>
			{#each PAGE.slice(0, shown) as line, i (i)}
				{#if line.kind === 'bar'}
					<div class="fd-bar"><Bar w={line.w} /></div>
				{:else}
					<div class="fd-line fd-head" class:fd-gap={line.kind === 'h2'}>{line.text}</div>
				{/if}
			{/each}
			{#if shown === PAGE.length}
				<div class="fd-line fd-gap"><span class="fd-dim">~ $</span> <Caret /></div>
			{/if}
		</Terminal>
	{/snippet}
</Stage>

<style>
	:global(.fd-term) {
		--color-fc-fg: var(--color-fc-on-inverse);
	}
	.fd-line {
		white-space: nowrap;
	}
	.fd-wrap {
		white-space: normal;
		word-break: break-all;
		margin-bottom: 12px;
	}
	.fd-head {
		font-weight: 700;
	}
	.fd-gap {
		margin-top: 12px;
	}
	.fd-bar {
		display: flex;
		align-items: center;
		height: 20px;
	}
	.fd-dim {
		color: var(--color-fc-on-inverse-muted);
	}
</style>

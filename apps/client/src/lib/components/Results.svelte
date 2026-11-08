<script lang="ts">
	import { Badge } from '@facile/muse';
	import type { ImageResult, NewsResult, WebResult } from '$lib/api';

	let {
		web = [],
		images = [],
		news = []
	}: { web?: WebResult[]; images?: ImageResult[]; news?: NewsResult[] } = $props();

	function host(url: string): string {
		try {
			return new URL(url).hostname.replace(/^www\./, '');
		} catch {
			return url;
		}
	}
</script>

{#if web.length > 0}
	<ol class="flex flex-col gap-2">
		{#each web as result, i (result.url)}
			<li>
				<a
					href={result.url}
					rel="noopener noreferrer"
					data-result={i}
					class="group block rounded-fc-lg p-3 outline-none hover:bg-fc-surface focus-visible:bg-fc-surface focus-visible:outline-2 focus-visible:outline-fc-ring"
				>
					<span class="flex min-w-0 items-center gap-2 text-fc-xs text-fc-fg-muted">
						<span class="truncate">{host(result.url)}</span>
						<Badge class="shrink-0 whitespace-nowrap">{result.engine}</Badge>
					</span>
					<span class="mt-1 block text-fc-md font-medium group-hover:underline">{result.title}</span>
					{#if result.snippet}<span class="mt-1 line-clamp-3 text-fc-sm text-fc-fg-muted">{result.snippet}</span>{/if}
				</a>
			</li>
		{/each}
	</ol>
{:else if images.length > 0}
	<ul class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
		{#each images as image, i (image.thumbnail)}
			<li>
				<a
					href={image.url}
					rel="noopener noreferrer"
					data-result={i}
					class="group block overflow-hidden rounded-fc-lg bg-fc-component outline-none focus-visible:outline-2 focus-visible:outline-fc-ring"
				>
					<img src={image.thumbnail} alt={image.title} loading="lazy" class="aspect-[4/3] w-full object-cover" />
					<span class="block truncate px-3 py-2 text-fc-xs text-fc-fg-muted">{image.source || host(image.url)}</span>
				</a>
			</li>
		{/each}
	</ul>
{:else if news.length > 0}
	<ol class="flex flex-col gap-2">
		{#each news as item, i (item.url)}
			<li>
				<a
					href={item.url}
					rel="noopener noreferrer"
					data-result={i}
					class="group flex gap-4 rounded-fc-lg p-3 outline-none hover:bg-fc-surface focus-visible:bg-fc-surface focus-visible:outline-2 focus-visible:outline-fc-ring"
				>
					<span class="min-w-0 flex-1">
						<span class="flex min-w-0 items-center gap-2 text-fc-xs text-fc-fg-muted">
							<span class="truncate">{item.source || host(item.url)}</span>
							{#if item.age}<span class="shrink-0 whitespace-nowrap">· {item.age}</span>{/if}
						</span>
						<span class="mt-1 block text-fc-md font-medium group-hover:underline">{item.title}</span>
						{#if item.snippet}<span class="mt-1 line-clamp-2 text-fc-sm text-fc-fg-muted">{item.snippet}</span>{/if}
					</span>
					{#if item.thumbnail}
						<img src={item.thumbnail} alt="" loading="lazy" class="hidden h-20 w-28 shrink-0 rounded-fc-md object-cover sm:block" />
					{/if}
				</a>
			</li>
		{/each}
	</ol>
{/if}

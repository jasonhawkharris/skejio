<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import type { TourDate } from '$lib/types';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	// Callers who represent at least one artist (managers/agents/labels) see a
	// merged tourdates list with no indication of whose date is whose, so
	// group by artist whenever we know of any represented artists.
	let groupedByArtist = $derived.by(() => {
		const groups: { artistId: string; artistName: string; tourdates: TourDate[] }[] = [];
		const indexByArtistId: Record<string, number> = {};
		for (const td of data.tourdates) {
			let index = indexByArtistId[td.user_id];
			if (index === undefined) {
				index = groups.length;
				indexByArtistId[td.user_id] = index;
				groups.push({
					artistId: td.user_id,
					artistName: data.artistNames[td.user_id] ?? 'Unknown artist',
					tourdates: []
				});
			}
			groups[index].tourdates.push(td);
		}
		return groups;
	});

	let showArtistGrouping = $derived(Object.keys(data.artistNames).length > 0);
</script>

{#snippet tourdateTable(tourdates: TourDate[])}
	<table>
		<thead>
			<tr>
				<th>Date</th>
				<th>City</th>
				<th>State</th>
				<th>Venue</th>
				<th></th>
			</tr>
		</thead>
		<tbody>
			{#each tourdates as td (td.id)}
				<tr
					class="row-link"
					tabindex="0"
					role="link"
					aria-label={`View tourdate in ${td.city} on ${td.date}`}
					onclick={() => goto(resolve('/tourdates/[id]', { id: td.id }))}
					onkeydown={(e) => {
						if (e.key === 'Enter') goto(resolve('/tourdates/[id]', { id: td.id }));
					}}
				>
					<td>{td.date}</td>
					<td>{td.city}</td>
					<td>{td.state ?? ''}</td>
					<td>{td.venue}</td>
					<td class="actions">
						<a
							class="edit-link"
							href={resolve('/tourdates/[id]/edit', { id: td.id })}
							onclick={(e) => e.stopPropagation()}>Edit</a
						>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{/snippet}

<div class="page">
	<header class="topbar">
		<span class="logo">Skejio</span>
		<form method="POST" action="?/logout">
			<button type="submit" class="logout">Log out</button>
		</form>
	</header>

	<main class="content">
		<div class="page-header">
			<h1>Tourdates</h1>
			<a class="new-link" href={resolve('/tourdates/new')}>+ New tourdate</a>
		</div>

		{#if data.tourdates.length === 0}
			<div class="card empty">
				<p>No tourdates yet.</p>
			</div>
		{:else if showArtistGrouping}
			{#each groupedByArtist as group (group.artistId)}
				<section class="card">
					<h2>{group.artistName}</h2>
					{@render tourdateTable(group.tourdates)}
				</section>
			{/each}
		{:else}
			<section class="card">
				{@render tourdateTable(data.tourdates)}
			</section>
		{/if}
	</main>
</div>

<style>
	.page {
		min-height: 100dvh;
		background: var(--color-bg);
		color: var(--color-text);
	}

	.topbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1.5rem clamp(1.5rem, 5vw, 4rem);
		border-bottom: 1px solid var(--color-border);
	}

	.logo {
		font-family: var(--font-display);
		font-weight: 700;
		font-size: 1.15rem;
		letter-spacing: -0.01em;
		text-transform: uppercase;
	}

	.logout {
		font: inherit;
		font-size: 0.85rem;
		font-weight: 500;
		color: var(--color-text);
		background: transparent;
		padding: 0.5rem 1.1rem;
		border: 1px solid var(--color-border-strong);
		border-radius: var(--radius-sm);
		cursor: pointer;
		transition: background 0.1s ease;
	}

	.logout:hover {
		background: var(--color-surface-hover);
	}

	.content {
		max-width: 60rem;
		margin: 0 auto;
		padding: 2rem clamp(1.5rem, 5vw, 4rem) 4rem;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.page-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin: 0.5rem 0 0.5rem;
	}

	.content h1 {
		font-size: 1.5rem;
		letter-spacing: -0.01em;
		margin: 0;
	}

	.new-link {
		flex-shrink: 0;
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--color-accent-text);
		text-decoration: none;
		padding: 0.55rem 1.1rem;
		border-radius: var(--radius-sm);
		background: var(--color-accent);
		transition: background 0.1s ease;
	}

	.new-link:hover {
		background: var(--color-accent-hover);
	}

	.card {
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		padding: 1.5rem;
	}

	.card h2 {
		font-size: 1.05rem;
		margin: 0 0 1rem;
		color: var(--color-text);
	}

	.empty p {
		margin: 0;
		color: var(--color-text-muted);
		font-size: 0.9rem;
	}

	table {
		width: 100%;
		border-collapse: collapse;
	}

	thead th {
		text-align: left;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		font-size: 0.72rem;
		font-weight: 600;
		color: var(--color-text-muted);
		padding: 0 0.75rem 0.6rem;
		border-bottom: 1px solid var(--color-border);
	}

	tbody td {
		padding: 0.7rem 0.75rem;
		font-size: 0.9rem;
		border-bottom: 1px solid var(--color-border);
	}

	tbody tr:last-child td {
		border-bottom: none;
	}

	.row-link {
		cursor: pointer;
	}

	.row-link:hover td {
		background: var(--color-surface-hover);
	}

	.row-link:focus-visible {
		outline: 2px solid var(--color-accent);
		outline-offset: -2px;
	}

	.actions {
		text-align: right;
	}

	.edit-link {
		display: inline-block;
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--color-accent);
		text-decoration: none;
		padding: 0.3rem 0.75rem;
		border: 1px solid var(--color-border-strong);
		border-radius: var(--radius-sm);
		transition: background 0.1s ease;
	}

	.edit-link:hover {
		background: var(--color-surface-hover);
	}
</style>

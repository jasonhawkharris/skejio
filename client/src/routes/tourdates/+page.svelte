<script lang="ts">
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
				<tr>
					<td>{td.date}</td>
					<td>{td.city}</td>
					<td>{td.state ?? ''}</td>
					<td>{td.venue}</td>
					<td class="actions">
						<a class="edit-link" href={resolve('/tourdates/[id]/edit', { id: td.id })}>Edit</a>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{/snippet}

<div class="page">
	<div class="glow glow-a"></div>
	<div class="glow glow-b"></div>

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
		position: relative;
		overflow: hidden;
		min-height: 100dvh;
		background: #08080c;
		color: #e7e7ee;
	}

	.glow {
		position: absolute;
		border-radius: 50%;
		filter: blur(90px);
		pointer-events: none;
		z-index: 0;
	}

	.glow-a {
		width: 500px;
		height: 500px;
		top: -150px;
		left: -100px;
		background: radial-gradient(circle, rgba(139, 92, 246, 0.35), transparent 70%);
	}

	.glow-b {
		width: 600px;
		height: 600px;
		bottom: -200px;
		right: -150px;
		background: radial-gradient(circle, rgba(56, 189, 248, 0.22), transparent 70%);
	}

	.topbar,
	.content {
		position: relative;
		z-index: 1;
	}

	.topbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1.5rem clamp(1.5rem, 5vw, 4rem);
	}

	.logo {
		font-family: var(--font-display);
		font-weight: 700;
		font-size: 1.25rem;
		letter-spacing: -0.02em;
	}

	.logout {
		font: inherit;
		font-size: 0.9rem;
		font-weight: 500;
		color: #e7e7ee;
		background: transparent;
		padding: 0.5rem 1.1rem;
		border: 1px solid rgba(231, 231, 238, 0.2);
		border-radius: 999px;
		cursor: pointer;
		transition:
			border-color 0.15s ease,
			background 0.15s ease;
	}

	.logout:hover {
		border-color: rgba(231, 231, 238, 0.5);
		background: rgba(231, 231, 238, 0.06);
	}

	.content {
		max-width: 60rem;
		margin: 0 auto;
		padding: 0 clamp(1.5rem, 5vw, 4rem) 4rem;
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
		font-size: 1.75rem;
		letter-spacing: -0.02em;
		margin: 0;
	}

	.new-link {
		flex-shrink: 0;
		font-size: 0.85rem;
		font-weight: 600;
		color: white;
		text-decoration: none;
		padding: 0.55rem 1.1rem;
		border-radius: 999px;
		background: linear-gradient(135deg, #8b5cf6, #6366f1);
		box-shadow: 0 12px 30px -10px rgba(139, 92, 246, 0.55);
		transition:
			transform 0.15s ease,
			box-shadow 0.15s ease;
	}

	.new-link:hover {
		transform: translateY(-1px);
		box-shadow: 0 16px 34px -8px rgba(139, 92, 246, 0.65);
	}

	.card {
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 16px;
		padding: 1.5rem;
	}

	.card h2 {
		font-size: 1.05rem;
		margin: 0 0 1rem;
		color: #f4f4f8;
	}

	.empty p {
		margin: 0;
		color: #9d9dac;
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
		color: #9d9dac;
		padding: 0 0.75rem 0.6rem;
		border-bottom: 1px solid rgba(255, 255, 255, 0.08);
	}

	tbody td {
		padding: 0.7rem 0.75rem;
		font-size: 0.9rem;
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
	}

	tbody tr:last-child td {
		border-bottom: none;
	}

	tbody tr:hover td {
		background: rgba(255, 255, 255, 0.02);
	}

	.actions {
		text-align: right;
	}

	.edit-link {
		display: inline-block;
		font-size: 0.8rem;
		font-weight: 600;
		color: #a78bfa;
		text-decoration: none;
		padding: 0.3rem 0.75rem;
		border: 1px solid rgba(167, 139, 250, 0.35);
		border-radius: 999px;
		transition:
			border-color 0.15s ease,
			background 0.15s ease;
	}

	.edit-link:hover {
		border-color: rgba(167, 139, 250, 0.7);
		background: rgba(167, 139, 250, 0.1);
	}
</style>

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
					<td>
						<a href={resolve('/tourdates/[id]/edit', { id: td.id })}>Edit</a>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{/snippet}

<h1>Tourdates</h1>

<form method="POST" action="?/logout">
	<button type="submit">Log out</button>
</form>

{#if data.tourdates.length === 0}
	<p>No tourdates yet.</p>
{:else if showArtistGrouping}
	{#each groupedByArtist as group (group.artistId)}
		<section>
			<h2>{group.artistName}</h2>
			{@render tourdateTable(group.tourdates)}
		</section>
	{/each}
{:else}
	{@render tourdateTable(data.tourdates)}
{/if}

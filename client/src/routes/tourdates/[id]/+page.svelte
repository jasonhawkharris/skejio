<script lang="ts">
	import { resolve } from '$app/paths';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<div class="page">
	<div class="card">
		{#if data.artistName}
			<p class="eyebrow">{data.artistName}</p>
		{/if}
		<h1>{data.tourdate.city}{data.tourdate.state ? `, ${data.tourdate.state}` : ''}</h1>

		<dl>
			<div class="row">
				<dt>Date</dt>
				<dd>{data.tourdate.date}</dd>
			</div>
			<div class="row">
				<dt>Venue</dt>
				<dd>{data.tourdate.venue}</dd>
			</div>
			<div class="row">
				<dt>City</dt>
				<dd>{data.tourdate.city}</dd>
			</div>
			<div class="row">
				<dt>State</dt>
				<dd>{data.tourdate.state ?? '—'}</dd>
			</div>
		</dl>

		<div class="buttons">
			<a class="edit-link" href={resolve('/tourdates/[id]/edit', { id: data.tourdate.id })}>Edit</a>
			<a href={resolve('/tourdates')}>Back to tourdates</a>
		</div>
	</div>
</div>

<style>
	.page {
		min-height: 100dvh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1.5rem;
		background: var(--color-bg);
	}

	.card {
		width: 100%;
		max-width: 420px;
		background: var(--color-surface);
		border-radius: var(--radius-lg);
		border: 1px solid var(--color-border);
		padding: 2.5rem 2rem;
		color: var(--color-text);
	}

	.eyebrow {
		text-transform: uppercase;
		letter-spacing: 0.1em;
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--color-accent);
		margin: 0 0 0.5rem;
	}

	.card h1 {
		font-size: 1.4rem;
		letter-spacing: -0.01em;
		margin: 0 0 1.75rem;
		color: var(--color-text);
	}

	dl {
		margin: 0 0 2rem;
		display: flex;
		flex-direction: column;
		gap: 0.9rem;
	}

	.row {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		padding-bottom: 0.9rem;
		border-bottom: 1px solid var(--color-border);
	}

	.row:last-child {
		border-bottom: none;
		padding-bottom: 0;
	}

	dt {
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--color-text-muted);
	}

	dd {
		margin: 0;
		font-size: 0.95rem;
		color: var(--color-text);
		text-align: right;
	}

	.buttons {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.edit-link {
		flex: 1;
		text-align: center;
		padding: 0.7rem 1rem;
		border-radius: var(--radius-sm);
		background: var(--color-accent);
		color: var(--color-accent-text);
		font-weight: 600;
		text-decoration: none;
		transition: background 0.1s ease;
	}

	.edit-link:hover {
		background: var(--color-accent-hover);
	}

	.buttons a:not(.edit-link) {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--color-text-muted);
		text-decoration: none;
	}

	.buttons a:not(.edit-link):hover {
		color: var(--color-text);
	}
</style>

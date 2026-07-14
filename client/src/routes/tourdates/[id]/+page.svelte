<script lang="ts">
	import { resolve } from '$app/paths';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<div class="page">
	<div class="glow glow-a"></div>
	<div class="glow glow-b"></div>

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
		position: relative;
		overflow: hidden;
		min-height: 100dvh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1.5rem;
		background: #08080c;
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

	.card {
		position: relative;
		z-index: 1;
		width: 100%;
		max-width: 420px;
		background: rgba(255, 255, 255, 0.03);
		backdrop-filter: blur(20px);
		border-radius: 16px;
		border: 1px solid rgba(255, 255, 255, 0.08);
		box-shadow: 0 20px 60px -20px rgba(0, 0, 0, 0.6);
		padding: 2.5rem 2rem;
		color: #e7e7ee;
	}

	.eyebrow {
		text-transform: uppercase;
		letter-spacing: 0.1em;
		font-size: 0.75rem;
		font-weight: 600;
		color: #a78bfa;
		margin: 0 0 0.5rem;
	}

	.card h1 {
		font-size: 1.6rem;
		letter-spacing: -0.02em;
		margin: 0 0 1.75rem;
		color: #f4f4f8;
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
		border-bottom: 1px solid rgba(255, 255, 255, 0.08);
	}

	.row:last-child {
		border-bottom: none;
		padding-bottom: 0;
	}

	dt {
		font-size: 0.8rem;
		font-weight: 600;
		color: #9d9dac;
	}

	dd {
		margin: 0;
		font-size: 0.95rem;
		color: #e7e7ee;
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
		border-radius: 8px;
		background: linear-gradient(135deg, #8b5cf6, #6366f1);
		color: white;
		font-weight: 600;
		text-decoration: none;
		box-shadow: 0 12px 30px -10px rgba(139, 92, 246, 0.55);
		transition:
			transform 0.15s ease,
			box-shadow 0.15s ease;
	}

	.edit-link:hover {
		transform: translateY(-1px);
		box-shadow: 0 16px 34px -8px rgba(139, 92, 246, 0.65);
	}

	.buttons a:not(.edit-link) {
		font-size: 0.85rem;
		font-weight: 600;
		color: #9d9dac;
		text-decoration: none;
	}

	.buttons a:not(.edit-link):hover {
		color: #e7e7ee;
	}
</style>

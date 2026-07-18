<script lang="ts">
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();
</script>

<div class="auth-page">
	<div class="auth-card">
		<div class="brand">
			<h1>Skejio</h1>
			<p class="tagline">Organize your tour!</p>
		</div>

		{@render children()}
	</div>
</div>

<style>
	.auth-page {
		min-height: 100dvh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1.5rem;
		background: var(--color-bg);
	}

	.auth-card {
		width: 100%;
		max-width: 380px;
		background: var(--color-surface);
		border-radius: var(--radius-lg);
		border: 1px solid var(--color-border);
		padding: 2.5rem 2rem;
	}

	.brand {
		text-align: center;
		margin-bottom: 2rem;
	}

	.brand h1 {
		font-size: 1.5rem;
		font-weight: 700;
		letter-spacing: -0.01em;
		text-transform: uppercase;
		color: var(--color-text);
	}

	.tagline {
		margin: 0.25rem 0 0;
		color: var(--color-text-muted);
		font-size: 0.925rem;
	}
</style>

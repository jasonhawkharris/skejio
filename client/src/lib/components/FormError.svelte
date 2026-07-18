<script lang="ts">
	let { message, spaced = false }: { message?: string | null; spaced?: boolean } = $props();
</script>

{#if message}
	<p class="error" style={spaced ? 'margin-top: 1rem' : undefined}>{message}</p>
{/if}

<style>
	.error {
		margin: 0;
		padding: 0.6rem 0.75rem;
		border-radius: var(--radius-sm);
		background: var(--color-danger-bg);
		color: var(--color-danger-text);
		font-size: 0.85rem;
	}
</style>

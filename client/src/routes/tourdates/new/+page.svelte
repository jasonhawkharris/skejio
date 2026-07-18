<script lang="ts">
	import { resolve } from '$app/paths';
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<div class="page">
	<div class="card">
		<h1>New tourdate</h1>

		<form method="POST" class="new-form">
			{#if data.artists.length > 0}
				<label>
					<span>Artist</span>
					<select name="artist_id" required>
						<option value="" disabled selected>Choose an artist</option>
						{#each data.artists as artist (artist.user_id)}
							<option value={artist.user_id}>{artist.name}</option>
						{/each}
					</select>
				</label>
			{/if}

			<label>
				<span>Date</span>
				<input type="date" name="date" required />
			</label>
			<label>
				<span>City</span>
				<input type="text" name="city" required />
			</label>
			<label>
				<span>State</span>
				<input type="text" name="state" />
			</label>
			<label>
				<span>Venue</span>
				<input type="text" name="venue" required />
			</label>

			{#if form?.error}
				<p class="error">{form.error}</p>
			{/if}

			<div class="buttons">
				<button type="submit">Create</button>
				<a href={resolve('/tourdates')}>Cancel</a>
			</div>
		</form>
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
		max-width: 380px;
		background: var(--color-surface);
		border-radius: var(--radius-lg);
		border: 1px solid var(--color-border);
		padding: 2.5rem 2rem;
	}

	.card h1 {
		text-align: center;
		font-size: 1.4rem;
		letter-spacing: -0.01em;
		margin: 0 0 2rem;
		color: var(--color-text);
	}

	.new-form {
		display: flex;
		flex-direction: column;
		gap: 1.1rem;
	}

	label {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--color-text-muted);
	}

	input,
	select {
		font: inherit;
		padding: 0.65rem 0.75rem;
		border-radius: var(--radius-sm);
		border: 1px solid var(--color-border-strong);
		background: var(--color-bg);
		color: var(--color-text);
		transition: border-color 0.1s ease;
	}

	select option {
		background: var(--color-surface);
		color: var(--color-text);
	}

	input:focus,
	select:focus {
		outline: 2px solid var(--color-accent);
		outline-offset: -1px;
		border-color: var(--color-accent);
	}

	.buttons {
		margin-top: 0.5rem;
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	button {
		flex: 1;
		padding: 0.7rem 1rem;
		border: none;
		border-radius: var(--radius-sm);
		background: var(--color-accent);
		color: var(--color-accent-text);
		font: inherit;
		font-weight: 600;
		cursor: pointer;
		transition: background 0.1s ease;
	}

	button:hover {
		background: var(--color-accent-hover);
	}

	.buttons a {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--color-text-muted);
		text-decoration: none;
	}

	.buttons a:hover {
		color: var(--color-text);
	}

	.error {
		margin: 0;
		padding: 0.6rem 0.75rem;
		border-radius: var(--radius-sm);
		background: var(--color-danger-bg);
		color: var(--color-danger-text);
		font-size: 0.85rem;
	}
</style>

<script lang="ts">
	import { resolve } from '$app/paths';
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<div class="page">
	<div class="glow glow-a"></div>
	<div class="glow glow-b"></div>

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
		max-width: 380px;
		background: rgba(255, 255, 255, 0.03);
		backdrop-filter: blur(20px);
		border-radius: 16px;
		border: 1px solid rgba(255, 255, 255, 0.08);
		box-shadow: 0 20px 60px -20px rgba(0, 0, 0, 0.6);
		padding: 2.5rem 2rem;
	}

	.card h1 {
		text-align: center;
		font-size: 1.5rem;
		letter-spacing: -0.02em;
		margin: 0 0 2rem;
		color: #f4f4f8;
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
		color: #9d9dac;
	}

	input,
	select {
		font: inherit;
		padding: 0.65rem 0.75rem;
		border-radius: 8px;
		border: 1px solid rgba(255, 255, 255, 0.12);
		background: rgba(255, 255, 255, 0.04);
		color: #e7e7ee;
		transition:
			border-color 0.15s ease,
			box-shadow 0.15s ease,
			background 0.15s ease;
	}

	select option {
		background: #14141c;
		color: #e7e7ee;
	}

	input:focus,
	select:focus {
		outline: none;
		border-color: #a78bfa;
		box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.2);
		background: rgba(255, 255, 255, 0.07);
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
		border-radius: 8px;
		background: linear-gradient(135deg, #8b5cf6, #6366f1);
		color: white;
		font: inherit;
		font-weight: 600;
		cursor: pointer;
		box-shadow: 0 12px 30px -10px rgba(139, 92, 246, 0.55);
		transition:
			transform 0.15s ease,
			box-shadow 0.15s ease;
	}

	button:hover {
		transform: translateY(-1px);
		box-shadow: 0 16px 34px -8px rgba(139, 92, 246, 0.65);
	}

	.buttons a {
		font-size: 0.85rem;
		font-weight: 600;
		color: #9d9dac;
		text-decoration: none;
	}

	.buttons a:hover {
		color: #e7e7ee;
	}

	.error {
		margin: 0;
		padding: 0.6rem 0.75rem;
		border-radius: 8px;
		background: rgba(220, 38, 38, 0.15);
		color: #f87171;
		font-size: 0.85rem;
	}
</style>

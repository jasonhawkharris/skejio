<script lang="ts">
	import type { ActionData } from './$types';

	let { form }: { form: ActionData } = $props();
</script>

<div class="login-page">
	<div class="login-card">
		<div class="brand">
			<h1>Skejio</h1>
			<p class="tagline">Organize your tour!</p>
		</div>

		<form method="POST" class="login-form">
			<label>
				<span>Email</span>
				<input type="email" name="email" autocomplete="email" required />
			</label>
			<label>
				<span>Password</span>
				<input type="password" name="password" autocomplete="current-password" required />
			</label>

			{#if form?.error}
				<p class="error">{form.error}</p>
			{/if}

			<button type="submit">Log in</button>
		</form>
	</div>
</div>

<style>
	.login-page {
		min-height: 100dvh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1.5rem;
		background:
			radial-gradient(circle at 20% 20%, rgba(124, 58, 237, 0.12), transparent 45%),
			radial-gradient(circle at 80% 0%, rgba(59, 130, 246, 0.12), transparent 40%), var(--color-bg);
	}

	.login-card {
		width: 100%;
		max-width: 380px;
		background: var(--color-surface);
		border-radius: var(--radius);
		box-shadow: var(--shadow);
		border: 1px solid var(--color-border);
		padding: 2.5rem 2rem;
	}

	.brand {
		text-align: center;
		margin-bottom: 2rem;
	}

	.brand h1 {
		font-size: 1.75rem;
		font-weight: 700;
		letter-spacing: -0.02em;
		background: linear-gradient(135deg, var(--color-primary), #4338ca);
		background-clip: text;
		-webkit-background-clip: text;
		color: transparent;
	}

	.tagline {
		margin: 0.25rem 0 0;
		color: var(--color-text-muted);
		font-size: 0.925rem;
	}

	.login-form {
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

	input {
		font: inherit;
		padding: 0.65rem 0.75rem;
		border-radius: 8px;
		border: 1px solid var(--color-border);
		background: #fafafa;
		color: var(--color-text);
		transition:
			border-color 0.15s ease,
			box-shadow 0.15s ease,
			background 0.15s ease;
	}

	input:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: 0 0 0 3px rgba(109, 40, 217, 0.15);
		background: white;
	}

	button {
		margin-top: 0.5rem;
		padding: 0.7rem 1rem;
		border: none;
		border-radius: 8px;
		background: var(--color-primary);
		color: white;
		font: inherit;
		font-weight: 600;
		cursor: pointer;
		transition: background 0.15s ease;
	}

	button:hover {
		background: var(--color-primary-hover);
	}

	.error {
		margin: 0;
		padding: 0.6rem 0.75rem;
		border-radius: 8px;
		background: rgba(220, 38, 38, 0.08);
		color: var(--color-error);
		font-size: 0.85rem;
	}
</style>

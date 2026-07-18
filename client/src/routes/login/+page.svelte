<script lang="ts">
	import { resolve } from '$app/paths';
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

		<p class="switch">
			No account yet? <a href={resolve('/sign-up')}>Sign up</a>
		</p>
	</div>
</div>

<style>
	.login-page {
		min-height: 100dvh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1.5rem;
		background: var(--color-bg);
	}

	.login-card {
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
		border-radius: var(--radius-sm);
		border: 1px solid var(--color-border-strong);
		background: var(--color-bg);
		color: var(--color-text);
		transition: border-color 0.1s ease;
	}

	input:focus {
		outline: 2px solid var(--color-accent);
		outline-offset: -1px;
		border-color: var(--color-accent);
	}

	button {
		margin-top: 0.5rem;
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

	.error {
		margin: 0;
		padding: 0.6rem 0.75rem;
		border-radius: var(--radius-sm);
		background: var(--color-danger-bg);
		color: var(--color-danger-text);
		font-size: 0.85rem;
	}

	.switch {
		margin: 1.5rem 0 0;
		text-align: center;
		font-size: 0.85rem;
		color: var(--color-text-muted);
	}

	.switch a {
		color: var(--color-accent);
		font-weight: 600;
		text-decoration: none;
	}

	.switch a:hover {
		text-decoration: underline;
	}
</style>

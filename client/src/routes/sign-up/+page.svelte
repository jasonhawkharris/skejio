<script lang="ts">
	import { resolve } from '$app/paths';
	import AuthCard from '$lib/components/AuthCard.svelte';
	import FormError from '$lib/components/FormError.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import type { ActionData } from './$types';

	let { form }: { form: ActionData } = $props();
</script>

<AuthCard>
	<form method="POST" class="login-form">
		<FormField label="Name" name="name" autocomplete="name" required />
		<FormField label="Email" name="email" type="email" autocomplete="email" required />
		<FormField
			label="Password"
			name="password"
			type="password"
			autocomplete="new-password"
			required
		/>
		<label class="field">
			<span>Account type</span>
			<select name="user_type" required>
				<option value="" disabled selected>Choose one</option>
				<option value="ARTIST">Artist</option>
				<option value="MANAGER">Manager</option>
				<option value="AGENT">Agent</option>
				<option value="CREW">Crew</option>
				<option value="LABEL">Label</option>
			</select>
		</label>

		<FormError message={form?.error} />

		<button type="submit">Sign up</button>
	</form>

	<p class="switch">
		Already have an account? <a href={resolve('/login')}>Log in</a>
	</p>
</AuthCard>

<style>
	.login-form {
		display: flex;
		flex-direction: column;
		gap: 1.1rem;
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

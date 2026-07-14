<script lang="ts">
	import { resolve } from '$app/paths';
	import type { ActionData } from './$types';

	let { form }: { form: ActionData } = $props();
</script>

<div class="login-page">
	<div class="glow glow-a"></div>
	<div class="glow glow-b"></div>

	<div class="login-card">
		<div class="brand">
			<h1>Skejio</h1>
			<p class="tagline">Organize your tour!</p>
		</div>

		<form method="POST" class="login-form">
			<label>
				<span>Name</span>
				<input type="text" name="name" autocomplete="name" required />
			</label>
			<label>
				<span>Email</span>
				<input type="email" name="email" autocomplete="email" required />
			</label>
			<label>
				<span>Password</span>
				<input type="password" name="password" autocomplete="new-password" required />
			</label>
			<label>
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

			{#if form?.error}
				<p class="error">{form.error}</p>
			{/if}

			<button type="submit">Sign up</button>
		</form>

		<p class="switch">
			Already have an account? <a href={resolve('/login')}>Log in</a>
		</p>
	</div>
</div>

<style>
	.login-page {
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

	.login-card {
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

	.brand {
		text-align: center;
		margin-bottom: 2rem;
	}

	.brand h1 {
		font-size: 1.75rem;
		font-weight: 700;
		letter-spacing: -0.02em;
		background: linear-gradient(135deg, #ffffff 30%, #a78bfa 100%);
		background-clip: text;
		-webkit-background-clip: text;
		color: transparent;
	}

	.tagline {
		margin: 0.25rem 0 0;
		color: #a1a1ae;
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

	button {
		margin-top: 0.5rem;
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

	.error {
		margin: 0;
		padding: 0.6rem 0.75rem;
		border-radius: 8px;
		background: rgba(220, 38, 38, 0.15);
		color: #f87171;
		font-size: 0.85rem;
	}

	.switch {
		margin: 1.5rem 0 0;
		text-align: center;
		font-size: 0.85rem;
		color: #9d9dac;
	}

	.switch a {
		color: #a78bfa;
		font-weight: 600;
		text-decoration: none;
	}

	.switch a:hover {
		text-decoration: underline;
	}
</style>

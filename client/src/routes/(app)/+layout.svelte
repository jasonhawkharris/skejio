<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	let { children } = $props();

	const NAV_ITEMS = [
		{ label: 'Dashboard', href: '/dashboard' },
		{ label: 'Tours', href: '/tours' },
		{ label: 'Financials', href: '/financials' },
		{ label: 'Merch', href: '/merch' },
		{ label: 'Riders', href: '/riders' },
		{ label: 'Team', href: '/team' }
	] as const;
</script>

<div class="shell">
	<div class="glow glow-a"></div>
	<div class="glow glow-b"></div>

	<aside class="sidebar">
		<span class="logo">Skejio</span>

		<nav>
			<ul>
				{#each NAV_ITEMS as item (item.href)}
					<li>
						<a href={resolve(item.href)} class:active={page.url.pathname === item.href}>
							{item.label}
						</a>
					</li>
				{/each}
			</ul>
		</nav>

		<form method="POST" action={resolve('/logout')}>
			<button type="submit" class="logout">Log out</button>
		</form>
	</aside>

	<main class="content">
		{@render children()}
	</main>
</div>

<style>
	.shell {
		position: relative;
		overflow: hidden;
		min-height: 100dvh;
		display: flex;
		background: #08080c;
		color: #e7e7ee;
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

	.sidebar,
	.content {
		position: relative;
		z-index: 1;
	}

	.sidebar {
		flex-shrink: 0;
		width: 15rem;
		display: flex;
		flex-direction: column;
		gap: 2rem;
		padding: 1.75rem 1.25rem;
		border-right: 1px solid rgba(255, 255, 255, 0.08);
	}

	.logo {
		font-family: var(--font-display);
		font-weight: 700;
		font-size: 1.25rem;
		letter-spacing: -0.02em;
		padding: 0 0.5rem;
	}

	nav {
		flex: 1;
	}

	nav ul {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	nav a {
		display: block;
		padding: 0.6rem 0.75rem;
		border-radius: 10px;
		font-size: 0.92rem;
		font-weight: 500;
		color: #9d9dac;
		text-decoration: none;
		transition:
			background 0.15s ease,
			color 0.15s ease;
	}

	nav a:hover {
		background: rgba(255, 255, 255, 0.05);
		color: #e7e7ee;
	}

	nav a.active {
		background: rgba(139, 92, 246, 0.14);
		color: #f4f4f8;
	}

	.logout {
		font: inherit;
		font-size: 0.9rem;
		font-weight: 500;
		color: #e7e7ee;
		background: transparent;
		width: 100%;
		padding: 0.55rem 0.75rem;
		border: 1px solid rgba(231, 231, 238, 0.2);
		border-radius: 999px;
		cursor: pointer;
		transition:
			border-color 0.15s ease,
			background 0.15s ease;
	}

	.logout:hover {
		border-color: rgba(231, 231, 238, 0.5);
		background: rgba(231, 231, 238, 0.06);
	}

	.content {
		flex: 1;
		min-width: 0;
		padding: 2.5rem clamp(1.5rem, 4vw, 3.5rem) 4rem;
	}
</style>

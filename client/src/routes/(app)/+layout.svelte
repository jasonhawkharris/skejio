<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	let { children } = $props();

	const NAV_ITEMS = [
		{ label: 'Dashboard', href: '/dashboard' },
		{ label: 'Shows', href: '/shows' },
		{ label: 'Financials', href: '/financials' },
		{ label: 'Merch', href: '/merch' },
		{ label: 'Riders', href: '/riders' },
		{ label: 'Team', href: '/team' }
	] as const;
</script>

<div class="shell">
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
		min-height: 100dvh;
		display: flex;
		background: var(--color-bg);
		color: var(--color-text);
	}

	.sidebar {
		flex-shrink: 0;
		width: 15rem;
		display: flex;
		flex-direction: column;
		gap: 2rem;
		padding: 1.75rem 1.25rem;
		background: var(--color-surface);
		border-right: 1px solid var(--color-border);
	}

	.logo {
		font-family: var(--font-display);
		font-weight: 700;
		font-size: 1.15rem;
		letter-spacing: -0.01em;
		text-transform: uppercase;
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
		gap: 0.15rem;
	}

	nav a {
		display: block;
		padding: 0.6rem 0.75rem;
		border-radius: var(--radius-sm);
		border-left: 2px solid transparent;
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--color-text-muted);
		text-decoration: none;
		transition:
			background 0.1s ease,
			color 0.1s ease,
			border-color 0.1s ease;
	}

	nav a:hover {
		background: var(--color-surface-hover);
		color: var(--color-text);
	}

	nav a.active {
		background: var(--color-surface-hover);
		border-left-color: var(--color-accent);
		color: var(--color-text);
	}

	.logout {
		font: inherit;
		font-size: 0.85rem;
		font-weight: 500;
		color: var(--color-text);
		background: transparent;
		width: 100%;
		padding: 0.55rem 0.75rem;
		border: 1px solid var(--color-border-strong);
		border-radius: var(--radius-sm);
		cursor: pointer;
		transition:
			border-color 0.1s ease,
			background 0.1s ease;
	}

	.logout:hover {
		background: var(--color-surface-hover);
	}

	.content {
		flex: 1;
		min-width: 0;
		padding: 2.5rem clamp(1.5rem, 4vw, 3.5rem) 4rem;
	}
</style>

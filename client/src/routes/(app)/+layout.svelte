<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { initials } from '$lib/format';
	import type { Snippet } from 'svelte';
	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: Snippet } = $props();

	const NAV_ITEMS = [
		{ key: 'dashboard', label: 'Dashboard', href: '/dashboard' },
		{ key: 'shows', label: 'Shows', href: '/shows' },
		{ key: 'financials', label: 'Financials', href: '/financials' },
		{ key: 'merch', label: 'Merch', href: '/merch' },
		{ key: 'riders', label: 'Riders', href: '/riders' },
		{ key: 'team', label: 'Team', href: '/team' }
	] as const;

	let activeItem = $derived(NAV_ITEMS.find((item) => item.href === page.url.pathname));
</script>

{#snippet navIcon(key: string)}
	<svg
		width="16"
		height="16"
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="1.75"
		stroke-linecap="round"
		stroke-linejoin="round"
	>
		{#if key === 'dashboard'}
			<rect x="3" y="3" width="7" height="7" rx="1.5" />
			<rect x="14" y="3" width="7" height="7" rx="1.5" />
			<rect x="3" y="14" width="7" height="7" rx="1.5" />
			<rect x="14" y="14" width="7" height="7" rx="1.5" />
		{:else if key === 'shows'}
			<rect x="3" y="4" width="18" height="17" rx="2" />
			<line x1="3" y1="9" x2="21" y2="9" />
			<line x1="8" y1="2" x2="8" y2="6" />
			<line x1="16" y1="2" x2="16" y2="6" />
		{:else if key === 'financials'}
			<circle cx="12" cy="12" r="9" />
			<text
				x="12"
				y="16"
				text-anchor="middle"
				font-size="11"
				font-weight="700"
				fill="currentColor"
				stroke="none">$</text
			>
		{:else if key === 'merch'}
			<path d="M6 8h12l-1 12H7L6 8Z" />
			<path d="M9 8V6a3 3 0 0 1 6 0v2" />
		{:else if key === 'riders'}
			<path d="M7 3h7l5 5v13H7V3Z" />
			<path d="M14 3v5h5" />
			<line x1="9" y1="13" x2="15" y2="13" />
			<line x1="9" y1="17" x2="15" y2="17" />
		{:else if key === 'team'}
			<circle cx="9" cy="8" r="3" />
			<path d="M3 20c0-3.3 2.7-6 6-6s6 2.7 6 6" />
			<circle cx="17" cy="9" r="2.5" />
			<path d="M15.5 14.2c2.5.4 4.5 2.6 4.5 5.3" />
		{:else if key === 'logout'}
			<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
			<polyline points="16 17 21 12 16 7" />
			<line x1="21" y1="12" x2="9" y2="12" />
		{/if}
	</svg>
{/snippet}

<div class="shell">
	<aside class="sidebar">
		<div class="brand">
			<span class="mark">S</span>
			<span class="wordmark">Skejio</span>
		</div>

		<nav>
			<ul>
				{#each NAV_ITEMS as item (item.href)}
					<li>
						<a href={resolve(item.href)} class:active={page.url.pathname === item.href}>
							{@render navIcon(item.key)}
							{item.label}
						</a>
					</li>
				{/each}
			</ul>
		</nav>

		<div class="footer">
			<div class="user">
				<span class="avatar">{initials(data.userName ?? '?')}</span>
				<span class="user-name">{data.userName ?? 'Account'}</span>
			</div>

			<form method="POST" action={resolve('/logout')}>
				<button type="submit" class="logout" aria-label="Log out" title="Log out">
					{@render navIcon('logout')}
				</button>
			</form>
		</div>
	</aside>

	<div class="main">
		<header class="topbar">
			<span class="crumb-root">Skejio</span>
			{#if activeItem}
				<span class="crumb-sep">/</span>
				<span class="crumb-current">{activeItem.label}</span>
			{/if}
		</header>

		<main class="content">
			{@render children()}
		</main>
	</div>
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
		gap: 1.5rem;
		padding: 1rem 0.75rem;
		background: var(--color-surface);
		border-right: 1px solid var(--color-border);
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.25rem 0.5rem;
	}

	.mark {
		flex-shrink: 0;
		width: 1.5rem;
		height: 1.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: var(--radius-sm);
		background: var(--color-accent);
		color: var(--color-accent-text);
		font-family: var(--font-display);
		font-weight: 700;
		font-size: 0.8rem;
	}

	.wordmark {
		font-family: var(--font-display);
		font-weight: 700;
		font-size: 0.85rem;
		letter-spacing: -0.01em;
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
		gap: 0.05rem;
	}

	nav a {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.4rem 0.5rem;
		border-radius: var(--radius-sm);
		font-size: 0.8125rem;
		font-weight: 500;
		color: var(--color-text-muted);
		text-decoration: none;
		transition:
			background 0.1s ease,
			color 0.1s ease;
	}

	nav a :global(svg) {
		flex-shrink: 0;
	}

	nav a:hover {
		background: var(--color-surface-hover);
		color: var(--color-text);
	}

	nav a.active {
		background: var(--color-active);
		color: var(--color-text);
	}

	.footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
		padding: 0.6rem 0.5rem 0.25rem;
		border-top: 1px solid var(--color-border);
	}

	.user {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		min-width: 0;
	}

	.avatar {
		flex-shrink: 0;
		width: 1.6rem;
		height: 1.6rem;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 999px;
		background: var(--color-active);
		font-size: 0.65rem;
		font-weight: 700;
		color: var(--color-text);
	}

	.user-name {
		min-width: 0;
		font-size: 0.8125rem;
		font-weight: 500;
		color: var(--color-text);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.logout {
		flex-shrink: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 1.75rem;
		height: 1.75rem;
		color: var(--color-text-muted);
		background: transparent;
		border: none;
		border-radius: var(--radius-sm);
		cursor: pointer;
		transition:
			background 0.1s ease,
			color 0.1s ease;
	}

	.logout:hover {
		background: var(--color-surface-hover);
		color: var(--color-text);
	}

	.main {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
	}

	.topbar {
		flex-shrink: 0;
		display: flex;
		align-items: center;
		gap: 0.4rem;
		height: 2.75rem;
		padding: 0 1.25rem;
		border-bottom: 1px solid var(--color-border);
		font-size: 0.8125rem;
	}

	.crumb-root {
		color: var(--color-text-muted);
	}

	.crumb-sep {
		color: var(--color-text-faint);
	}

	.crumb-current {
		color: var(--color-text);
		font-weight: 600;
	}

	.content {
		flex: 1;
		min-width: 0;
		padding: 2rem clamp(1.5rem, 4vw, 3.5rem) 4rem;
	}
</style>

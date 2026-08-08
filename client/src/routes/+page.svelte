<script lang="ts">
	import { resolve } from '$app/paths';
</script>

<div class="landing">
	<header class="nav">
		<span class="logo">Skejio</span>
		<nav class="nav-links">
			<a href={resolve('/login')}>Log in</a>
			<a class="nav-cta" href={resolve('/sign-up')}>Sign up</a>
		</nav>
	</header>

	<main class="hero">
		<p class="eyebrow">Tour management, simplified</p>
		<h1>Run your tour<br />like a pro.</h1>
		<p class="subtext">
			Every date, city, and venue in one place — with scoped, revocable access for the managers,
			agents, and labels who work the tour with you.
		</p>
		<a class="cta" href={resolve('/sign-up')}>Get started →</a>
	</main>

	<section class="features">
		<div class="feature">
			<h2>Own your dates</h2>
			<p>Artists create and control every tour date they own, always.</p>
		</div>
		<div class="feature">
			<h2>Delegate access</h2>
			<p>Grant managers, agents, and labels scoped access to the artists they represent.</p>
		</div>
		<div class="feature">
			<h2>One shared calendar</h2>
			<p>Everyone works off the same schedule — no spreadsheets, no confusion.</p>
		</div>
	</section>

	<footer class="footer">
		<p>Skejio — Organize your tour!</p>
	</footer>
</div>

<style>
	.landing {
		min-height: 100dvh;
		background: var(--color-bg);
		color: var(--color-text);
		font-family: var(--font-body);
		display: flex;
		flex-direction: column;
	}

	.nav {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1.5rem clamp(1.5rem, 5vw, 4rem);
		border-bottom: 1px solid var(--color-border);
	}

	.logo {
		font-family: var(--font-display);
		font-weight: 700;
		font-size: 1.15rem;
		letter-spacing: -0.01em;
		text-transform: uppercase;
	}

	.nav-links {
		display: flex;
		align-items: center;
		gap: 1.25rem;
	}

	.nav-links a {
		color: var(--color-text);
		text-decoration: none;
		font-size: 0.9rem;
		font-weight: 500;
	}

	.nav-cta {
		padding: 0.5rem 1.1rem;
		border: 1px solid var(--color-border-strong);
		border-radius: var(--radius-sm);
		transition: background 0.1s ease;
	}

	.nav-cta:hover {
		background: var(--color-surface-hover);
	}

	.hero {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		text-align: center;
		padding: 3rem 1.5rem 4rem;
	}

	.eyebrow {
		text-transform: uppercase;
		letter-spacing: 0.14em;
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--color-accent);
		margin: 0 0 1.25rem;
	}

	.hero h1 {
		font-family: var(--font-display);
		font-size: clamp(2.5rem, 6vw, 4.25rem);
		line-height: 1.05;
		letter-spacing: -0.02em;
		margin: 0 0 1.25rem;
		color: var(--color-text);
	}

	.subtext {
		max-width: 34rem;
		color: var(--color-text-muted);
		font-size: 1.05rem;
		line-height: 1.6;
		margin: 0 0 2.25rem;
	}

	.cta {
		display: inline-block;
		padding: 0.85rem 1.75rem;
		border-radius: var(--radius-sm);
		background: var(--color-accent);
		color: var(--color-accent-text);
		font-weight: 600;
		text-decoration: none;
		transition: background 0.1s ease;
	}

	.cta:hover {
		background: var(--color-accent-hover);
	}

	.features {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 1.5rem;
		max-width: 68rem;
		margin: 0 auto;
		padding: 0 clamp(1.5rem, 5vw, 4rem) 5rem;
		width: 100%;
	}

	.feature {
		padding: 1.75rem;
		border-radius: var(--radius-lg);
		background: var(--color-surface);
		border: 1px solid var(--color-border);
	}

	.feature h2 {
		font-size: 1.05rem;
		margin: 0 0 0.5rem;
		color: var(--color-text);
	}

	.feature p {
		margin: 0;
		color: var(--color-text-muted);
		font-size: 0.9rem;
		line-height: 1.55;
	}

	.footer {
		padding: 1.5rem;
		text-align: center;
		font-size: 0.8rem;
		color: var(--color-text-muted);
		border-top: 1px solid var(--color-border);
	}

	@media (max-width: 480px) {
		.nav {
			padding: 1.25rem 1.25rem;
		}

		.nav-links {
			gap: 0.75rem;
		}

		.nav-cta {
			padding: 0.45rem 0.85rem;
		}

		.hero {
			padding: 2rem 1.25rem 3rem;
		}

		.subtext {
			font-size: 0.95rem;
		}

		.features {
			grid-template-columns: 1fr;
		}
	}
</style>

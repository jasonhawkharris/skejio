<script lang="ts">
	import { resolve } from '$app/paths';
</script>

<div class="landing">
	<div class="glow glow-a"></div>
	<div class="glow glow-b"></div>

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
		position: relative;
		min-height: 100dvh;
		overflow: hidden;
		background: #08080c;
		color: #e7e7ee;
		font-family: var(--font-body);
		display: flex;
		flex-direction: column;
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

	.nav,
	.hero,
	.features,
	.footer {
		position: relative;
		z-index: 1;
	}

	.nav {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1.5rem clamp(1.5rem, 5vw, 4rem);
	}

	.logo {
		font-family: var(--font-display);
		font-weight: 700;
		font-size: 1.25rem;
		letter-spacing: -0.02em;
	}

	.nav-links {
		display: flex;
		align-items: center;
		gap: 1.25rem;
	}

	.nav-links a {
		color: #e7e7ee;
		text-decoration: none;
		font-size: 0.9rem;
		font-weight: 500;
	}

	.nav-cta {
		padding: 0.5rem 1.1rem;
		border: 1px solid rgba(231, 231, 238, 0.2);
		border-radius: 999px;
		transition:
			border-color 0.15s ease,
			background 0.15s ease;
	}

	.nav-cta:hover {
		border-color: rgba(231, 231, 238, 0.5);
		background: rgba(231, 231, 238, 0.06);
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
		color: #a78bfa;
		margin: 0 0 1.25rem;
	}

	.hero h1 {
		font-family: var(--font-display);
		font-size: clamp(2.5rem, 6vw, 4.25rem);
		line-height: 1.05;
		letter-spacing: -0.03em;
		margin: 0 0 1.25rem;
		background: linear-gradient(135deg, #ffffff 30%, #a78bfa 100%);
		background-clip: text;
		-webkit-background-clip: text;
		color: transparent;
	}

	.subtext {
		max-width: 34rem;
		color: #a1a1ae;
		font-size: 1.05rem;
		line-height: 1.6;
		margin: 0 0 2.25rem;
	}

	.cta {
		display: inline-block;
		padding: 0.85rem 1.75rem;
		border-radius: 999px;
		background: linear-gradient(135deg, #8b5cf6, #6366f1);
		color: white;
		font-weight: 600;
		text-decoration: none;
		box-shadow: 0 12px 30px -10px rgba(139, 92, 246, 0.55);
		transition:
			transform 0.15s ease,
			box-shadow 0.15s ease;
	}

	.cta:hover {
		transform: translateY(-1px);
		box-shadow: 0 16px 34px -8px rgba(139, 92, 246, 0.65);
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
		border-radius: 16px;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.08);
	}

	.feature h2 {
		font-size: 1.05rem;
		margin: 0 0 0.5rem;
		color: #f4f4f8;
	}

	.feature p {
		margin: 0;
		color: #9d9dac;
		font-size: 0.9rem;
		line-height: 1.55;
	}

	.footer {
		padding: 1.5rem;
		text-align: center;
		font-size: 0.8rem;
		color: #6b6b78;
		border-top: 1px solid rgba(255, 255, 255, 0.06);
	}
</style>

<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const MONTHS = ['JAN', 'FEB', 'MAR', 'APR', 'MAY', 'JUN', 'JUL', 'AUG', 'SEP', 'OCT', 'NOV', 'DEC'];

	function dateBadge(dateStr: string) {
		const [, m, d] = dateStr.split('-').map(Number);
		return { month: MONTHS[m - 1], day: d };
	}

	function formatShowTime(t: string | null): string | null {
		if (!t) return null;
		const [hStr, mStr] = t.split(':');
		let h = Number(hStr);
		const period = h >= 12 ? 'PM' : 'AM';
		h = h % 12 || 12;
		return `${h}:${mStr} ${period}`;
	}

	const ROLE_LABELS: Record<string, string> = { MANAGER: 'Manager', AGENT: 'Agent', LABEL: 'Label' };

	function initials(name: string): string {
		return name
			.split(/\s+/)
			.filter(Boolean)
			.slice(0, 2)
			.map((part) => part[0]?.toUpperCase())
			.join('');
	}
</script>

<div class="page">
	<div class="glow glow-a"></div>
	<div class="glow glow-b"></div>

	<header class="topbar">
		<span class="logo">Skejio</span>
		<form method="POST" action="?/logout">
			<button type="submit" class="logout">Log out</button>
		</form>
	</header>

	<main class="content">
		<section class="frame">
			<div class="section-header">
				<h1>Upcoming shows</h1>
				<a class="all-link" href={resolve('/tourdates')}>View all tourdates →</a>
			</div>

			{#if data.upcoming.length === 0}
				<div class="card empty">
					<p>No upcoming shows on the books.</p>
					<a class="new-link" href={resolve('/tourdates/new')}>+ Add a tourdate</a>
				</div>
			{:else}
				<ul class="show-list">
					{#each data.upcoming as td (td.id)}
						{@const badge = dateBadge(td.date)}
						{@const showTime = formatShowTime(td.show_start)}
						<li>
							<button
								type="button"
								class="show-row"
								onclick={() => goto(resolve('/tourdates/[id]', { id: td.id }))}
							>
								<span class="date-badge">
									<span class="badge-month">{badge.month}</span>
									<span class="badge-day">{badge.day}</span>
								</span>
								<span class="show-info">
									<span class="venue">{td.venue}</span>
									<span class="location">{td.city}{td.state ? `, ${td.state}` : ''}</span>
								</span>
								{#if showTime}
									<span class="show-time">{showTime}</span>
								{/if}
							</button>
						</li>
					{/each}
				</ul>
			{/if}
		</section>

		<section class="frame">
			<div class="section-header">
				<h2>My team</h2>
			</div>

			{#if data.team.length === 0}
				<div class="card empty">
					<p>No representatives yet.</p>
				</div>
			{:else}
				<ul class="team-list">
					{#each data.team as rep (rep.relationship_id)}
						<li class="team-row">
							<span class="avatar">{initials(rep.name)}</span>
							<span class="rep-name">{rep.name}</span>
							<span class="role-badge">{ROLE_LABELS[rep.user_type] ?? rep.user_type}</span>
						</li>
					{/each}
				</ul>
			{/if}
		</section>
	</main>
</div>

<style>
	.page {
		position: relative;
		overflow: hidden;
		min-height: 100dvh;
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

	.topbar,
	.content {
		position: relative;
		z-index: 1;
	}

	.topbar {
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

	.logout {
		font: inherit;
		font-size: 0.9rem;
		font-weight: 500;
		color: #e7e7ee;
		background: transparent;
		padding: 0.5rem 1.1rem;
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
		max-width: 42rem;
		margin: 0 auto;
		padding: 0 clamp(1.5rem, 5vw, 4rem) 4rem;
		display: flex;
		flex-direction: column;
		gap: 2.5rem;
	}

	.frame {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.section-header {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.content h1 {
		font-size: 1.75rem;
		letter-spacing: -0.02em;
		margin: 0;
	}

	.content h2 {
		font-size: 1.1rem;
		letter-spacing: -0.01em;
		margin: 0;
		color: #f4f4f8;
	}

	.all-link {
		font-size: 0.85rem;
		font-weight: 500;
		color: #9d9dac;
		text-decoration: none;
	}

	.all-link:hover {
		color: #e7e7ee;
	}

	.card {
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 16px;
		padding: 1.5rem;
	}

	.empty {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 0.9rem;
	}

	.empty p {
		margin: 0;
		color: #9d9dac;
		font-size: 0.9rem;
	}

	.new-link {
		font-size: 0.85rem;
		font-weight: 600;
		color: white;
		text-decoration: none;
		padding: 0.55rem 1.1rem;
		border-radius: 999px;
		background: linear-gradient(135deg, #8b5cf6, #6366f1);
		box-shadow: 0 12px 30px -10px rgba(139, 92, 246, 0.55);
	}

	.show-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.65rem;
	}

	.show-row {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 1.1rem;
		padding: 0.85rem 1.1rem;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 14px;
		font: inherit;
		color: inherit;
		text-align: left;
		cursor: pointer;
		transition:
			border-color 0.15s ease,
			background 0.15s ease;
	}

	.show-row:hover {
		background: rgba(255, 255, 255, 0.05);
		border-color: rgba(167, 139, 250, 0.4);
	}

	.show-row:focus-visible {
		outline: 2px solid #a78bfa;
		outline-offset: 2px;
	}

	.date-badge {
		flex-shrink: 0;
		width: 3.25rem;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 0.4rem 0;
		border-radius: 10px;
		background: rgba(139, 92, 246, 0.12);
		border: 1px solid rgba(167, 139, 250, 0.3);
	}

	.badge-month {
		font-size: 0.65rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		color: #a78bfa;
	}

	.badge-day {
		font-family: var(--font-display);
		font-size: 1.15rem;
		font-weight: 700;
		line-height: 1.1;
		color: #f4f4f8;
	}

	.show-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
	}

	.venue {
		font-weight: 600;
		font-size: 0.95rem;
		color: #f4f4f8;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.location {
		font-size: 0.82rem;
		color: #9d9dac;
	}

	.show-time {
		flex-shrink: 0;
		font-size: 0.82rem;
		font-weight: 600;
		color: #9d9dac;
	}

	.team-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.65rem;
	}

	.team-row {
		display: flex;
		align-items: center;
		gap: 0.9rem;
		padding: 0.75rem 1.1rem;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 14px;
	}

	.avatar {
		flex-shrink: 0;
		width: 2.25rem;
		height: 2.25rem;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 999px;
		background: rgba(56, 189, 248, 0.14);
		border: 1px solid rgba(125, 211, 252, 0.3);
		font-size: 0.78rem;
		font-weight: 700;
		color: #7dd3fc;
	}

	.rep-name {
		flex: 1;
		min-width: 0;
		font-weight: 600;
		font-size: 0.92rem;
		color: #f4f4f8;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.role-badge {
		flex-shrink: 0;
		font-size: 0.72rem;
		font-weight: 600;
		letter-spacing: 0.03em;
		color: #9d9dac;
		padding: 0.3rem 0.7rem;
		border-radius: 999px;
		border: 1px solid rgba(255, 255, 255, 0.12);
	}
</style>

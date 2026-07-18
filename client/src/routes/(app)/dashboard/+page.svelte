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
</div>

<style>
	.page {
		max-width: 42rem;
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

	.page h1 {
		font-size: 1.5rem;
		letter-spacing: -0.01em;
		margin: 0;
	}

	.page h2 {
		font-size: 1.05rem;
		letter-spacing: -0.01em;
		margin: 0;
		color: var(--color-text);
	}

	.all-link {
		font-size: 0.85rem;
		font-weight: 500;
		color: var(--color-text-muted);
		text-decoration: none;
	}

	.all-link:hover {
		color: var(--color-text);
	}

	.card {
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
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
		color: var(--color-text-muted);
		font-size: 0.9rem;
	}

	.new-link {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--color-accent-text);
		text-decoration: none;
		padding: 0.55rem 1.1rem;
		border-radius: var(--radius-sm);
		background: var(--color-accent);
	}

	.new-link:hover {
		background: var(--color-accent-hover);
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
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font: inherit;
		color: inherit;
		text-align: left;
		cursor: pointer;
		transition: border-color 0.1s ease;
	}

	.show-row:hover {
		border-color: var(--color-border-strong);
	}

	.show-row:focus-visible {
		outline: 2px solid var(--color-accent);
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
		border-radius: var(--radius-sm);
		background: var(--color-bg);
		border: 1px solid var(--color-border-strong);
	}

	.badge-month {
		font-size: 0.65rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		color: var(--color-accent);
	}

	.badge-day {
		font-family: var(--font-display);
		font-size: 1.15rem;
		font-weight: 700;
		line-height: 1.1;
		color: var(--color-text);
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
		color: var(--color-text);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.location {
		font-size: 0.82rem;
		color: var(--color-text-muted);
	}

	.show-time {
		flex-shrink: 0;
		font-size: 0.82rem;
		font-weight: 600;
		color: var(--color-text-muted);
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
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
	}

	.avatar {
		flex-shrink: 0;
		width: 2.25rem;
		height: 2.25rem;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 999px;
		background: var(--color-bg);
		border: 1px solid var(--color-border-strong);
		font-size: 0.78rem;
		font-weight: 700;
		color: var(--color-accent);
	}

	.rep-name {
		flex: 1;
		min-width: 0;
		font-weight: 600;
		font-size: 0.92rem;
		color: var(--color-text);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.role-badge {
		flex-shrink: 0;
		font-size: 0.72rem;
		font-weight: 600;
		letter-spacing: 0.03em;
		color: var(--color-text-muted);
		padding: 0.3rem 0.7rem;
		border-radius: var(--radius-sm);
		border: 1px solid var(--color-border-strong);
	}
</style>

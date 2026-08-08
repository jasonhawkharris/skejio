<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { formatClockTime, initials } from '$lib/format';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const MONTHS = [
		'JAN',
		'FEB',
		'MAR',
		'APR',
		'MAY',
		'JUN',
		'JUL',
		'AUG',
		'SEP',
		'OCT',
		'NOV',
		'DEC'
	];

	function dateBadge(dateStr: string) {
		const [, m, d] = dateStr.split('-').map(Number);
		return { month: MONTHS[m - 1], day: d };
	}

	function formatShowTime(t: string | null): string | null {
		return t ? formatClockTime(t) : null;
	}

	const ROLE_LABELS: Record<string, string> = {
		MANAGER: 'Manager',
		AGENT: 'Agent',
		LABEL: 'Label'
	};
</script>

<div class="page">
	<section class="frame">
		<div class="section-header">
			<h1>Upcoming shows</h1>
			<a class="all-link" href={resolve('/shows')}>View all tourdates →</a>
		</div>

		{#if data.upcoming.length === 0}
			<div class="empty">
				<p>No upcoming shows on the books.</p>
				<a class="new-link" href={resolve('/shows')}>+ Add a tourdate</a>
			</div>
		{:else}
			<ul class="show-list">
				{#each data.upcoming as td (td.id)}
					{@const badge = dateBadge(td.date)}
					{@const showTime = formatShowTime(td.show_start)}
					<li>
						<button type="button" class="show-row" onclick={() => goto(resolve('/shows'))}>
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
			<div class="empty">
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
		font-size: 1.15rem;
		letter-spacing: -0.01em;
		margin: 0;
	}

	.page h2 {
		font-size: 0.95rem;
		letter-spacing: -0.01em;
		margin: 0;
		color: var(--color-text);
	}

	.all-link {
		font-size: 0.8rem;
		font-weight: 500;
		color: var(--color-text-muted);
		text-decoration: none;
	}

	.all-link:hover {
		color: var(--color-text);
	}

	.empty {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 0.75rem;
		padding: 0.5rem 0;
	}

	.empty p {
		margin: 0;
		color: var(--color-text-muted);
		font-size: 0.85rem;
	}

	.new-link {
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--color-accent-text);
		text-decoration: none;
		padding: 0.4rem 0.85rem;
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
		gap: 0.1rem;
	}

	.show-row {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 0.5rem 0.6rem;
		background: transparent;
		border: none;
		border-radius: var(--radius-sm);
		font: inherit;
		color: inherit;
		text-align: left;
		cursor: pointer;
		transition: background 0.1s ease;
	}

	.show-row:hover {
		background: var(--color-surface-hover);
	}

	.show-row:focus-visible {
		outline: 2px solid var(--color-accent);
		outline-offset: -2px;
	}

	.date-badge {
		flex-shrink: 0;
		width: 2.75rem;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 0.3rem 0;
		border-radius: var(--radius-sm);
		background: var(--color-bg);
		border: 1px solid var(--color-border);
	}

	.badge-month {
		font-size: 0.6rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		color: var(--color-text-faint);
	}

	.badge-day {
		font-family: var(--font-display);
		font-size: 1rem;
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
		font-weight: 500;
		font-size: 0.85rem;
		color: var(--color-text);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.location {
		font-size: 0.78rem;
		color: var(--color-text-muted);
	}

	.show-time {
		flex-shrink: 0;
		font-size: 0.78rem;
		font-weight: 500;
		color: var(--color-text-faint);
	}

	.team-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.1rem;
	}

	.team-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.5rem 0.6rem;
		border-radius: var(--radius-sm);
		transition: background 0.1s ease;
	}

	.team-row:hover {
		background: var(--color-surface-hover);
	}

	.avatar {
		flex-shrink: 0;
		width: 1.9rem;
		height: 1.9rem;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 999px;
		background: var(--color-active);
		font-size: 0.7rem;
		font-weight: 700;
		color: var(--color-text);
	}

	.rep-name {
		flex: 1;
		min-width: 0;
		font-weight: 500;
		font-size: 0.85rem;
		color: var(--color-text);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.role-badge {
		flex-shrink: 0;
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.03em;
		color: var(--color-text-faint);
	}

	@media (max-width: 480px) {
		.page {
			gap: 2rem;
		}

		.show-row {
			gap: 0.6rem;
			padding: 0.5rem 0.4rem;
		}

		.date-badge {
			width: 2.35rem;
		}

		.show-time {
			display: none;
		}
	}
</style>

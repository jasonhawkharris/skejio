<script lang="ts">
	import { enhance } from '$app/forms';
	import { untrack } from 'svelte';
	import type { TourDate } from '$lib/types';
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	// data.shows is sorted soonest-first, so the first entry is the next show.
	// untrack: capture that once on mount, rather than re-selecting it every
	// time data reloads and clobbering whatever the user has since clicked.
	let selectedId = $state<string | null>(untrack(() => data.shows[0]?.id ?? null));
	let editingShow = $state<TourDate | null>(null);
	let showAdvanced = $state(false);
	let creating = $state(false);
	let showCreateAdvanced = $state(false);

	let selectedShow = $derived(data.shows.find((s) => s.id === selectedId) ?? null);

	type EditableFields = Pick<
		TourDate,
		| 'date'
		| 'city'
		| 'state'
		| 'venue'
		| 'poc_name'
		| 'poc_number'
		| 'poc_email'
		| 'promoter_name'
		| 'promoter_number'
		| 'promoter_email'
		| 'doors'
		| 'show_start'
		| 'show_end'
		| 'load_in'
		| 'sound_check'
		| 'advance'
	>;

	const BLANK_SHOW: EditableFields = {
		date: '',
		city: '',
		state: null,
		venue: '',
		poc_name: null,
		poc_number: null,
		poc_email: null,
		promoter_name: null,
		promoter_number: null,
		promoter_email: null,
		doors: null,
		show_start: null,
		show_end: null,
		load_in: null,
		sound_check: null,
		advance: null
	};

	function formatState(city: string, state: string | null): string {
		return state ? `${city}, ${state}` : city;
	}

	const CLOCK_FIELDS = new Set<keyof TourDate>(['doors', 'show_start', 'show_end', 'load_in', 'sound_check']);

	// Backend clock fields come back as 24h "HH:MM" - render US-style 12h with AM/PM.
	function formatClockTime(t: string): string {
		const [hStr, mStr] = t.split(':');
		let h = Number(hStr);
		const period = h >= 12 ? 'PM' : 'AM';
		h = h % 12 || 12;
		return `${h}:${mStr} ${period}`;
	}

	function detailValue(key: keyof TourDate, value: string | null): string {
		if (value == null) return '—';
		return CLOCK_FIELDS.has(key) ? formatClockTime(value) : value;
	}

	// Formats as a US number with area code, e.g. "(555) 123-4567", as the
	// user types - digits beyond 10 (a stray country code, etc.) are dropped
	// rather than silently reflowing the format.
	function formatPhoneInput(e: Event & { currentTarget: HTMLInputElement }) {
		const digits = e.currentTarget.value.replace(/\D/g, '').slice(0, 10);
		if (digits.length > 6) {
			e.currentTarget.value = `(${digits.slice(0, 3)}) ${digits.slice(3, 6)}-${digits.slice(6)}`;
		} else if (digits.length > 3) {
			e.currentTarget.value = `(${digits.slice(0, 3)}) ${digits.slice(3)}`;
		} else if (digits.length > 0) {
			e.currentTarget.value = `(${digits}`;
		} else {
			e.currentTarget.value = '';
		}
	}

	function openEdit(show: TourDate) {
		editingShow = show;
		showAdvanced = false;
	}

	function closeEdit() {
		editingShow = null;
	}

	function openCreate() {
		creating = true;
		showCreateAdvanced = false;
	}

	function closeCreate() {
		creating = false;
	}

	const DETAIL_FIELDS: { label: string; key: keyof TourDate }[] = [
		{ label: 'Date', key: 'date' },
		{ label: 'Venue', key: 'venue' },
		{ label: 'City', key: 'city' },
		{ label: 'State', key: 'state' },
		{ label: 'POC name', key: 'poc_name' },
		{ label: 'POC number', key: 'poc_number' },
		{ label: 'POC email', key: 'poc_email' },
		{ label: 'Promoter name', key: 'promoter_name' },
		{ label: 'Promoter number', key: 'promoter_number' },
		{ label: 'Promoter email', key: 'promoter_email' },
		{ label: 'Doors', key: 'doors' },
		{ label: 'Show start', key: 'show_start' },
		{ label: 'Show end', key: 'show_end' },
		{ label: 'Load-in', key: 'load_in' },
		{ label: 'Sound check', key: 'sound_check' },
		{ label: 'Advance', key: 'advance' }
	];
</script>

<div class="split">
	<section class="panel list-panel">
		<div class="panel-header">
			<h1>Upcoming shows</h1>
			<button type="button" class="new-link" onclick={openCreate}>+ New tourdate</button>
		</div>

		{#if data.shows.length === 0}
			<p class="empty">No upcoming shows on the books.</p>
		{:else}
			<ul class="show-list">
				{#each data.shows as show (show.id)}
					<li>
						<button
							type="button"
							class="show-row"
							class:selected={show.id === selectedId}
							onclick={() => (selectedId = show.id)}
						>
							<span class="show-main">
								<span class="show-date">{show.date}</span>
								<span class="show-venue">{show.venue}</span>
								<span class="show-location">{formatState(show.city, show.state)}</span>
							</span>
							<span
								role="button"
								tabindex="0"
								class="edit-btn"
								onclick={(e) => {
									e.stopPropagation();
									openEdit(show);
								}}
								onkeydown={(e) => {
									if (e.key === 'Enter' || e.key === ' ') {
										e.preventDefault();
										e.stopPropagation();
										openEdit(show);
									}
								}}
							>
								Edit
							</span>
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</section>

	<section class="panel detail-panel">
		{#if selectedShow}
			<h2>{formatState(selectedShow.city, selectedShow.state)}</h2>
			<dl>
				{#each DETAIL_FIELDS as field (field.key)}
					<div class="row">
						<dt>{field.label}</dt>
						<dd>{detailValue(field.key, selectedShow[field.key])}</dd>
					</div>
				{/each}
			</dl>
		{:else}
			<p class="empty">Select a show to view details.</p>
		{/if}
	</section>
</div>

{#snippet basicFields(values: EditableFields)}
	<div class="field-grid">
		<label>
			<span>Date</span>
			<input type="date" name="date" value={values.date} required />
		</label>
		<label>
			<span>City</span>
			<input type="text" name="city" value={values.city} required />
		</label>
		<label>
			<span>State</span>
			<input type="text" name="state" value={values.state ?? ''} />
		</label>
		<label>
			<span>Venue</span>
			<input type="text" name="venue" value={values.venue} required />
		</label>
	</div>
{/snippet}

{#snippet advancedFields(values: EditableFields)}
	<div class="field-grid">
		<label>
			<span>POC name</span>
			<input type="text" name="poc_name" value={values.poc_name ?? ''} />
		</label>
		<label>
			<span>POC number</span>
			<input
				type="text"
				inputmode="tel"
				name="poc_number"
				value={values.poc_number ?? ''}
				oninput={formatPhoneInput}
			/>
		</label>
		<label>
			<span>POC email</span>
			<input type="email" name="poc_email" value={values.poc_email ?? ''} />
		</label>
		<label>
			<span>Promoter name</span>
			<input type="text" name="promoter_name" value={values.promoter_name ?? ''} />
		</label>
		<label>
			<span>Promoter number</span>
			<input
				type="text"
				inputmode="tel"
				name="promoter_number"
				value={values.promoter_number ?? ''}
				oninput={formatPhoneInput}
			/>
		</label>
		<label>
			<span>Promoter email</span>
			<input type="email" name="promoter_email" value={values.promoter_email ?? ''} />
		</label>
		<label>
			<span>Doors</span>
			<input type="time" name="doors" value={values.doors ?? ''} />
		</label>
		<label>
			<span>Show start</span>
			<input type="time" name="show_start" value={values.show_start ?? ''} />
		</label>
		<label>
			<span>Show end</span>
			<input type="time" name="show_end" value={values.show_end ?? ''} />
		</label>
		<label>
			<span>Load-in</span>
			<input type="time" name="load_in" value={values.load_in ?? ''} />
		</label>
		<label>
			<span>Sound check</span>
			<input type="time" name="sound_check" value={values.sound_check ?? ''} />
		</label>
		<label class="span-2">
			<span>Advance</span>
			<input type="text" name="advance" value={values.advance ?? ''} />
		</label>
	</div>
{/snippet}

{#if editingShow}
	{@const show = editingShow}
	<div
		class="modal-backdrop"
		role="presentation"
		onclick={closeEdit}
		onkeydown={(e) => e.key === 'Escape' && closeEdit()}
	>
		<div
			class="modal"
			role="dialog"
			aria-modal="true"
			aria-label="Edit show"
			tabindex="-1"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
		>
			<h2>Edit show</h2>

			<form
				method="POST"
				action="?/update"
				use:enhance={() => {
					return async ({ result, update }) => {
						if (result.type === 'success') {
							closeEdit();
						}
						await update();
					};
				}}
			>
				<input type="hidden" name="id" value={show.id} />

				{@render basicFields(show)}

				{#if form?.error}
					<p class="error">{form.error}</p>
				{/if}

				<button type="button" class="advanced-toggle" onclick={() => (showAdvanced = !showAdvanced)}>
					{showAdvanced ? '− Hide advanced fields' : '+ Show advanced fields'}
				</button>

				{#if showAdvanced}
					{@render advancedFields(show)}
				{/if}

				<div class="buttons">
					<button type="submit" class="primary">Save</button>
					<button type="button" onclick={closeEdit}>Cancel</button>
				</div>
			</form>
		</div>
	</div>
{/if}

{#if creating}
	<div
		class="modal-backdrop"
		role="presentation"
		onclick={closeCreate}
		onkeydown={(e) => e.key === 'Escape' && closeCreate()}
	>
		<div
			class="modal"
			role="dialog"
			aria-modal="true"
			aria-label="New tourdate"
			tabindex="-1"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
		>
			<h2>New tourdate</h2>

			<form
				method="POST"
				action="?/create"
				use:enhance={() => {
					return async ({ result, update }) => {
						if (result.type === 'success') {
							closeCreate();
						}
						await update();
					};
				}}
			>
				{#if data.artists.length > 0}
					<label class="artist-field">
						<span>Artist</span>
						<select name="artist_id" required>
							<option value="" disabled selected>Choose an artist</option>
							{#each data.artists as artist (artist.user_id)}
								<option value={artist.user_id}>{artist.name}</option>
							{/each}
						</select>
					</label>
				{/if}

				{@render basicFields(BLANK_SHOW)}

				{#if form?.error}
					<p class="error">{form.error}</p>
				{/if}

				<button
					type="button"
					class="advanced-toggle"
					onclick={() => (showCreateAdvanced = !showCreateAdvanced)}
				>
					{showCreateAdvanced ? '− Hide advanced fields' : '+ Show advanced fields'}
				</button>

				{#if showCreateAdvanced}
					{@render advancedFields(BLANK_SHOW)}
				{/if}

				<div class="buttons">
					<button type="submit" class="primary">Create</button>
					<button type="button" onclick={closeCreate}>Cancel</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<style>
	.split {
		display: flex;
		gap: 1.5rem;
		height: calc(100dvh - 6.5rem);
	}

	.panel {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 1rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		padding: 1.5rem;
		overflow-y: auto;
	}

	.panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
	}

	.panel h1 {
		font-size: 1.3rem;
		letter-spacing: -0.01em;
		margin: 0;
	}

	.new-link {
		flex-shrink: 0;
		font: inherit;
		font-size: 0.82rem;
		font-weight: 600;
		color: var(--color-accent-text);
		text-decoration: none;
		padding: 0.5rem 0.9rem;
		border: none;
		border-radius: var(--radius-sm);
		background: var(--color-accent);
		cursor: pointer;
		transition: background 0.1s ease;
	}

	.new-link:hover {
		background: var(--color-accent-hover);
	}

	.panel h2 {
		font-size: 1.2rem;
		letter-spacing: -0.01em;
		margin: 0;
	}

	.empty {
		margin: 0;
		color: var(--color-text-muted);
		font-size: 0.9rem;
	}

	.show-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.show-row {
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.75rem 1rem;
		background: var(--color-bg);
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

	.show-row.selected {
		border-color: var(--color-accent);
	}

	.show-main {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
		min-width: 0;
	}

	.show-date {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--color-accent);
	}

	.show-venue {
		font-weight: 600;
		font-size: 0.92rem;
		color: var(--color-text);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.show-location {
		font-size: 0.8rem;
		color: var(--color-text-muted);
	}

	.edit-btn {
		flex-shrink: 0;
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--color-text);
		padding: 0.35rem 0.85rem;
		border: 1px solid var(--color-border-strong);
		border-radius: var(--radius-sm);
		cursor: pointer;
	}

	.edit-btn:hover {
		background: var(--color-surface-hover);
	}

	dl {
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.row {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		padding-bottom: 0.75rem;
		border-bottom: 1px solid var(--color-border);
	}

	.row:last-child {
		border-bottom: none;
		padding-bottom: 0;
	}

	dt {
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--color-text-muted);
	}

	dd {
		margin: 0;
		font-size: 0.9rem;
		color: var(--color-text);
		text-align: right;
	}

	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.6);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1.5rem;
		z-index: 10;
	}

	.modal {
		width: 100%;
		max-width: 52rem;
		max-height: 90dvh;
		overflow-y: auto;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		padding: 2rem;
	}

	.modal h2 {
		font-size: 1.2rem;
		letter-spacing: -0.01em;
		margin: 0 0 1.5rem;
	}

	.field-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 1rem;
	}

	.span-2 {
		grid-column: 1 / -1;
	}

	label {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--color-text-muted);
	}

	.artist-field {
		margin-bottom: 1rem;
	}

	input,
	select {
		font: inherit;
		padding: 0.6rem 0.7rem;
		border-radius: var(--radius-sm);
		border: 1px solid var(--color-border-strong);
		background: var(--color-bg);
		color: var(--color-text);
	}

	select option {
		background: var(--color-surface);
		color: var(--color-text);
	}

	input:focus,
	select:focus {
		outline: 2px solid var(--color-accent);
		outline-offset: -1px;
		border-color: var(--color-accent);
	}

	.advanced-toggle {
		align-self: flex-start;
		margin: 1.25rem 0;
		font: inherit;
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--color-accent);
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
	}

	.buttons {
		margin-top: 1.5rem;
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.buttons button {
		flex: 1;
		padding: 0.65rem 1rem;
		border-radius: var(--radius-sm);
		font: inherit;
		font-weight: 600;
		cursor: pointer;
	}

	.buttons .primary {
		border: none;
		background: var(--color-accent);
		color: var(--color-accent-text);
	}

	.buttons .primary:hover {
		background: var(--color-accent-hover);
	}

	.buttons button:not(.primary) {
		background: transparent;
		border: 1px solid var(--color-border-strong);
		color: var(--color-text);
	}

	.buttons button:not(.primary):hover {
		background: var(--color-surface-hover);
	}

	.error {
		margin: 1rem 0 0;
		padding: 0.6rem 0.75rem;
		border-radius: var(--radius-sm);
		background: var(--color-danger-bg);
		color: var(--color-danger-text);
		font-size: 0.85rem;
	}
</style>

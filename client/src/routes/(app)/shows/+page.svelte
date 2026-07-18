<script lang="ts">
	import { enhance } from '$app/forms';
	import { untrack } from 'svelte';
	import FormError from '$lib/components/FormError.svelte';
	import FormField from '$lib/components/FormField.svelte';
	import { formatClockTime } from '$lib/format';
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

	const CLOCK_FIELDS = new Set<keyof TourDate>([
		'doors',
		'show_start',
		'show_end',
		'load_in',
		'sound_check'
	]);

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
		<FormField label="Date" name="date" type="date" value={values.date} required />
		<FormField label="City" name="city" value={values.city} required />
		<FormField label="State" name="state" value={values.state} />
		<FormField label="Venue" name="venue" value={values.venue} required />
	</div>
{/snippet}

{#snippet advancedFields(values: EditableFields)}
	<div class="field-grid">
		<FormField label="POC name" name="poc_name" value={values.poc_name} />
		<FormField
			label="POC number"
			name="poc_number"
			inputmode="tel"
			value={values.poc_number}
			oninput={formatPhoneInput}
		/>
		<FormField label="POC email" name="poc_email" type="email" value={values.poc_email} />
		<FormField label="Promoter name" name="promoter_name" value={values.promoter_name} />
		<FormField
			label="Promoter number"
			name="promoter_number"
			inputmode="tel"
			value={values.promoter_number}
			oninput={formatPhoneInput}
		/>
		<FormField
			label="Promoter email"
			name="promoter_email"
			type="email"
			value={values.promoter_email}
		/>
		<FormField label="Doors" name="doors" type="time" value={values.doors} />
		<FormField label="Show start" name="show_start" type="time" value={values.show_start} />
		<FormField label="Show end" name="show_end" type="time" value={values.show_end} />
		<FormField label="Load-in" name="load_in" type="time" value={values.load_in} />
		<FormField label="Sound check" name="sound_check" type="time" value={values.sound_check} />
		<FormField label="Advance" name="advance" value={values.advance} wide />
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

				<FormError message={form?.error} spaced />

				<button
					type="button"
					class="advanced-toggle"
					onclick={() => (showAdvanced = !showAdvanced)}
				>
					{showAdvanced ? '− Hide advanced fields' : '+ Show advanced fields'}
				</button>

				{#if showAdvanced}
					{@render advancedFields(show)}
				{/if}

				<div class="buttons">
					<button type="button" onclick={closeEdit}>Cancel</button>
					<button type="submit" class="primary">Save</button>
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
					<label class="field artist-field">
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

				<FormError message={form?.error} spaced />

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
					<button type="button" onclick={closeCreate}>Cancel</button>
					<button type="submit" class="primary">Create</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<style>
	.split {
		display: flex;
		height: calc(100dvh - 8.75rem);
	}

	.panel {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		padding: 0 1.5rem;
		overflow-y: auto;
	}

	.list-panel {
		padding-left: 0;
		border-right: 1px solid var(--color-border);
	}

	.panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
	}

	.panel h1 {
		font-size: 1.05rem;
		letter-spacing: -0.01em;
		margin: 0;
	}

	.new-link {
		flex-shrink: 0;
		font: inherit;
		font-size: 0.78rem;
		font-weight: 600;
		color: var(--color-accent-text);
		text-decoration: none;
		padding: 0.4rem 0.75rem;
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
		font-size: 1.05rem;
		letter-spacing: -0.01em;
		margin: 0;
	}

	.empty {
		margin: 0;
		color: var(--color-text-muted);
		font-size: 0.85rem;
	}

	.show-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.05rem;
	}

	.show-row {
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: space-between;
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

	.show-row:hover .edit-btn {
		opacity: 1;
	}

	.show-row.selected {
		background: var(--color-active);
	}

	.show-main {
		display: flex;
		flex-direction: column;
		gap: 0.1rem;
		min-width: 0;
	}

	.show-date {
		font-size: 0.7rem;
		font-weight: 600;
		color: var(--color-text-faint);
	}

	.show-venue {
		font-weight: 500;
		font-size: 0.85rem;
		color: var(--color-text);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.show-location {
		font-size: 0.76rem;
		color: var(--color-text-muted);
	}

	.edit-btn {
		flex-shrink: 0;
		font-size: 0.76rem;
		font-weight: 600;
		color: var(--color-text-muted);
		padding: 0.3rem 0.6rem;
		border: none;
		border-radius: var(--radius-sm);
		cursor: pointer;
		opacity: 0;
		transition:
			opacity 0.1s ease,
			background 0.1s ease,
			color 0.1s ease;
	}

	.edit-btn:hover {
		background: var(--color-active);
		color: var(--color-text);
	}

	.edit-btn:focus-visible {
		opacity: 1;
		outline: 2px solid var(--color-accent);
		outline-offset: 1px;
	}

	dl {
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}

	.row {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		padding-bottom: 0.6rem;
		border-bottom: 1px solid var(--color-border);
	}

	.row:last-child {
		border-bottom: none;
		padding-bottom: 0;
	}

	dt {
		font-size: 0.78rem;
		font-weight: 500;
		color: var(--color-text-muted);
	}

	dd {
		margin: 0;
		font-size: 0.85rem;
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
		border-radius: var(--radius-md);
		padding: 1.5rem;
	}

	.modal h2 {
		font-size: 0.95rem;
		font-weight: 600;
		letter-spacing: -0.01em;
		margin: 0 0 1.25rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid var(--color-border);
	}

	.field-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 0.85rem;
	}

	.artist-field {
		margin-bottom: 0.85rem;
	}

	.advanced-toggle {
		align-self: flex-start;
		margin: 1rem 0;
		font: inherit;
		font-size: 0.78rem;
		font-weight: 600;
		color: var(--color-accent);
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
	}

	.buttons {
		margin-top: 1.25rem;
		padding-top: 1rem;
		border-top: 1px solid var(--color-border);
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 0.5rem;
	}

	.buttons button {
		padding: 0.4rem 0.85rem;
		border-radius: var(--radius-sm);
		font: inherit;
		font-size: 0.8rem;
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
		border: none;
		color: var(--color-text-muted);
	}

	.buttons button:not(.primary):hover {
		background: var(--color-surface-hover);
		color: var(--color-text);
	}
</style>

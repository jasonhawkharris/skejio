<script lang="ts">
	import { resolve } from '$app/paths';
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<h1>Edit tourdate</h1>

<form method="POST">
	<label>
		Date
		<input type="date" name="date" value={data.tourdate.date} required />
	</label>
	<label>
		City
		<input type="text" name="city" value={data.tourdate.city} required />
	</label>
	<label>
		State
		<input type="text" name="state" value={data.tourdate.state ?? ''} />
	</label>
	<label>
		Venue
		<input type="text" name="venue" value={data.tourdate.venue} required />
	</label>
	<button type="submit">Save</button>
	<a href={resolve('/tourdates')}>Cancel</a>
</form>

{#if form?.error}
	<p>{form.error}</p>
{/if}

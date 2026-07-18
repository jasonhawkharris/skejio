<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';

	interface Props {
		label: string;
		name: string;
		type?: string;
		value?: string | null;
		required?: boolean;
		autocomplete?: HTMLInputAttributes['autocomplete'];
		inputmode?: HTMLInputAttributes['inputmode'];
		oninput?: (e: Event & { currentTarget: HTMLInputElement }) => void;
		wide?: boolean;
	}

	let {
		label,
		name,
		type = 'text',
		value = '',
		required = false,
		autocomplete,
		inputmode,
		oninput,
		wide = false
	}: Props = $props();
</script>

<label class="field" style={wide ? 'grid-column: 1 / -1' : undefined}>
	<span>{label}</span>
	<input {type} {name} value={value ?? ''} {required} {autocomplete} {inputmode} {oninput} />
</label>

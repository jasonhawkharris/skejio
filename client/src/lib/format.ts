// Backend clock fields come back as 24h "HH:MM" - render US-style 12h with AM/PM.
export function formatClockTime(t: string): string {
	const [hStr, mStr] = t.split(':');
	let h = Number(hStr);
	const period = h >= 12 ? 'PM' : 'AM';
	h = h % 12 || 12;
	return `${h}:${mStr} ${period}`;
}

export function initials(name: string): string {
	return name
		.split(/\s+/)
		.filter(Boolean)
		.slice(0, 2)
		.map((part) => part[0]?.toUpperCase())
		.join('');
}

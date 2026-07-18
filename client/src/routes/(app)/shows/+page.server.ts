import { error, fail, redirect } from '@sveltejs/kit';
import { BackendError, backendFetch } from '$lib/server/backend';
import type { RepresentedUser, TourDate } from '$lib/types';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, fetch, cookies }) => {
	if (!locals.token) {
		redirect(303, '/login');
	}
	const token = locals.token;

	try {
		const [tourdates, artists] = await Promise.all([
			backendFetch<TourDate[]>(fetch, '/tourdates', { token }),
			// 403s for a caller who isn't a manager/agent/label - they only ever
			// create tourdates for themselves, so there's nothing to pick from.
			backendFetch<RepresentedUser[]>(fetch, '/represented-artists', { token }).catch((err) => {
				if (err instanceof BackendError && err.status === 403) return [];
				throw err;
			})
		]);

		const today = new Date().toISOString().slice(0, 10);
		const shows = tourdates
			.filter((td) => td.date >= today)
			.sort((a, b) => (a.date < b.date ? -1 : 1));
		return { shows, artists };
	} catch (err) {
		if (err instanceof BackendError && err.status === 401) {
			cookies.delete('session_token', { path: '/' });
			cookies.delete('user_type', { path: '/' });
			cookies.delete('user_name', { path: '/' });
			redirect(303, '/login');
		}
		throw err;
	}
};

// Blank means "leave the field empty" for a nullable column - form fields have
// no way to distinguish that from omission, so every edit sends a full
// replacement value for every nullable field rather than relying on the
// backend's PATCH omitted-vs-null semantics.
function nullableField(form: FormData, key: string): string | null {
	const v = form.get(key);
	return typeof v === 'string' && v.trim() !== '' ? v.trim() : null;
}

export const actions: Actions = {
	create: async ({ request, locals, fetch, cookies }) => {
		if (!locals.token) {
			redirect(303, '/login');
		}
		const token = locals.token;

		const form = await request.formData();
		const date = form.get('date');
		const city = form.get('city');
		const venue = form.get('venue');
		const artistId = form.get('artist_id');

		if (typeof date !== 'string' || !date) {
			return fail(400, { error: 'date is required' });
		}
		if (typeof city !== 'string' || !city) {
			return fail(400, { error: 'city is required' });
		}
		if (typeof venue !== 'string' || !venue) {
			return fail(400, { error: 'venue is required' });
		}

		const body: Record<string, unknown> = {
			date,
			city,
			venue,
			state: nullableField(form, 'state'),
			poc_name: nullableField(form, 'poc_name'),
			poc_number: nullableField(form, 'poc_number'),
			poc_email: nullableField(form, 'poc_email'),
			promoter_name: nullableField(form, 'promoter_name'),
			promoter_number: nullableField(form, 'promoter_number'),
			promoter_email: nullableField(form, 'promoter_email'),
			doors: nullableField(form, 'doors'),
			show_start: nullableField(form, 'show_start'),
			show_end: nullableField(form, 'show_end'),
			load_in: nullableField(form, 'load_in'),
			sound_check: nullableField(form, 'sound_check'),
			advance: nullableField(form, 'advance')
		};
		if (typeof artistId === 'string' && artistId) {
			body.artist_id = artistId;
		}

		try {
			await backendFetch(fetch, '/tourdates', { method: 'POST', token, body });
		} catch (err) {
			if (err instanceof BackendError && err.status === 401) {
				cookies.delete('session_token', { path: '/' });
				cookies.delete('user_type', { path: '/' });
				cookies.delete('user_name', { path: '/' });
				redirect(303, '/login');
			}
			if (err instanceof BackendError && err.status === 403) {
				error(403, 'you do not have access to create tourdates for this artist');
			}
			if (err instanceof BackendError) {
				const errBody = err.body as { error?: string } | null;
				return fail(err.status, { error: errBody?.error ?? 'failed to create tourdate' });
			}
			return fail(500, { error: 'unable to reach the server' });
		}

		return { success: true };
	},

	update: async ({ request, locals, fetch, cookies }) => {
		if (!locals.token) {
			redirect(303, '/login');
		}
		const token = locals.token;

		const form = await request.formData();
		const id = form.get('id');
		const date = form.get('date');
		const city = form.get('city');
		const venue = form.get('venue');

		if (typeof id !== 'string' || !id) {
			return fail(400, { error: 'missing tourdate id' });
		}
		if (typeof date !== 'string' || !date) {
			return fail(400, { error: 'date is required' });
		}
		if (typeof city !== 'string' || !city) {
			return fail(400, { error: 'city is required' });
		}
		if (typeof venue !== 'string' || !venue) {
			return fail(400, { error: 'venue is required' });
		}

		const body = {
			date,
			city,
			venue,
			state: nullableField(form, 'state'),
			poc_name: nullableField(form, 'poc_name'),
			poc_number: nullableField(form, 'poc_number'),
			poc_email: nullableField(form, 'poc_email'),
			promoter_name: nullableField(form, 'promoter_name'),
			promoter_number: nullableField(form, 'promoter_number'),
			promoter_email: nullableField(form, 'promoter_email'),
			doors: nullableField(form, 'doors'),
			show_start: nullableField(form, 'show_start'),
			show_end: nullableField(form, 'show_end'),
			load_in: nullableField(form, 'load_in'),
			sound_check: nullableField(form, 'sound_check'),
			advance: nullableField(form, 'advance')
		};

		try {
			await backendFetch(fetch, `/tourdates/${id}`, { method: 'PATCH', token, body });
		} catch (err) {
			if (err instanceof BackendError && err.status === 401) {
				cookies.delete('session_token', { path: '/' });
				cookies.delete('user_type', { path: '/' });
				cookies.delete('user_name', { path: '/' });
				redirect(303, '/login');
			}
			if (err instanceof BackendError && err.status === 404) {
				error(404, 'tourdate not found');
			}
			if (err instanceof BackendError) {
				const errBody = err.body as { error?: string } | null;
				return fail(err.status, { error: errBody?.error ?? 'failed to update tourdate' });
			}
			return fail(500, { error: 'unable to reach the server' });
		}

		return { success: true };
	}
};

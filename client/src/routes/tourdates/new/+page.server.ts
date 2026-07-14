import { error, fail, redirect } from '@sveltejs/kit';
import { BackendError, backendFetch } from '$lib/server/backend';
import type { RepresentedUser } from '$lib/types';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, fetch }) => {
	if (!locals.token) {
		redirect(303, '/login');
	}

	// 403s for a caller who isn't a manager/agent/label - they only ever
	// create tourdates for themselves, so there's nothing to pick from.
	const artists = await backendFetch<RepresentedUser[]>(fetch, '/represented-artists', {
		token: locals.token
	}).catch((err) => {
		if (err instanceof BackendError && err.status === 403) return [];
		throw err;
	});

	return { artists };
};

export const actions: Actions = {
	default: async ({ request, locals, fetch, cookies }) => {
		if (!locals.token) {
			redirect(303, '/login');
		}
		const token = locals.token;

		const form = await request.formData();
		const date = form.get('date');
		const city = form.get('city');
		const state = form.get('state');
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

		const stateValue = typeof state === 'string' && state.trim() !== '' ? state.trim() : null;
		const body: Record<string, unknown> = { date, city, state: stateValue, venue };
		if (typeof artistId === 'string' && artistId) {
			body.artist_id = artistId;
		}

		try {
			await backendFetch(fetch, '/tourdates', {
				method: 'POST',
				token,
				body
			});
		} catch (err) {
			if (err instanceof BackendError && err.status === 401) {
				cookies.delete('session_token', { path: '/' });
				redirect(303, '/login');
			}
			if (err instanceof BackendError && err.status === 403) {
				error(403, 'you do not have access to create tourdates for this artist');
			}
			if (err instanceof BackendError) {
				const body = err.body as { error?: string } | null;
				return fail(err.status, { error: body?.error ?? 'failed to create tourdate' });
			}
			return fail(500, { error: 'unable to reach the server' });
		}

		redirect(303, '/tourdates');
	}
};

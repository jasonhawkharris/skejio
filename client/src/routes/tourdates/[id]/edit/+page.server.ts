import { error, fail, redirect } from '@sveltejs/kit';
import { BackendError, backendFetch } from '$lib/server/backend';
import type { TourDate } from '$lib/types';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, fetch, cookies, params }) => {
	if (!locals.token) {
		redirect(303, '/login');
	}

	try {
		const tourdate = await backendFetch<TourDate>(fetch, `/tourdates/${params.id}`, {
			token: locals.token
		});
		return { tourdate };
	} catch (err) {
		if (err instanceof BackendError && err.status === 401) {
			cookies.delete('session_token', { path: '/' });
			redirect(303, '/login');
		}
		if (err instanceof BackendError && err.status === 404) {
			error(404, 'tourdate not found');
		}
		throw err;
	}
};

export const actions: Actions = {
	default: async ({ request, locals, fetch, cookies, params }) => {
		if (!locals.token) {
			redirect(303, '/login');
		}
		const token = locals.token;

		const form = await request.formData();
		const date = form.get('date');
		const city = form.get('city');
		const state = form.get('state');
		const venue = form.get('venue');

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

		try {
			await backendFetch(fetch, `/tourdates/${params.id}`, {
				method: 'PATCH',
				token,
				body: { date, city, state: stateValue, venue }
			});
		} catch (err) {
			if (err instanceof BackendError && err.status === 401) {
				cookies.delete('session_token', { path: '/' });
				redirect(303, '/login');
			}
			if (err instanceof BackendError && err.status === 404) {
				error(404, 'tourdate not found');
			}
			if (err instanceof BackendError) {
				const body = err.body as { error?: string } | null;
				return fail(err.status, { error: body?.error ?? 'failed to update tourdate' });
			}
			return fail(500, { error: 'unable to reach the server' });
		}

		redirect(303, '/tourdates');
	}
};

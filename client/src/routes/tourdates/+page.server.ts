import { redirect } from '@sveltejs/kit';
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
			// see their own tourdates, so there's nothing to group by artist.
			backendFetch<RepresentedUser[]>(fetch, '/represented-artists', { token }).catch((err) => {
				if (err instanceof BackendError && err.status === 403) return [];
				throw err;
			})
		]);

		const artistNames: Record<string, string> = Object.fromEntries(
			artists.map((a) => [a.user_id, a.name])
		);

		return { tourdates, artistNames };
	} catch (err) {
		if (err instanceof BackendError && err.status === 401) {
			cookies.delete('session_token', { path: '/' });
			redirect(303, '/login');
		}
		throw err;
	}
};

export const actions: Actions = {
	logout: async ({ locals, cookies, fetch }) => {
		if (locals.token) {
			await backendFetch(fetch, '/logout', { method: 'POST', token: locals.token }).catch(() => {});
		}
		cookies.delete('session_token', { path: '/' });
		redirect(303, '/login');
	}
};

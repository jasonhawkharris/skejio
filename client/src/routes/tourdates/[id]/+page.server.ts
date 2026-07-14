import { error, redirect } from '@sveltejs/kit';
import { BackendError, backendFetch } from '$lib/server/backend';
import type { RepresentedUser, TourDate } from '$lib/types';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, fetch, cookies, params }) => {
	if (!locals.token) {
		redirect(303, '/login');
	}
	const token = locals.token;

	try {
		const [tourdate, artists] = await Promise.all([
			backendFetch<TourDate>(fetch, `/tourdates/${params.id}`, { token }),
			// 403s for a caller who isn't a manager/agent/label - they only ever
			// see their own tourdates, so there's no artist name to resolve.
			backendFetch<RepresentedUser[]>(fetch, '/represented-artists', { token }).catch((err) => {
				if (err instanceof BackendError && err.status === 403) return [];
				throw err;
			})
		]);

		const artistName = artists.find((a) => a.user_id === tourdate.user_id)?.name ?? null;

		return { tourdate, artistName };
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

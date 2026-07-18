import { redirect } from '@sveltejs/kit';
import { BackendError, backendFetch } from '$lib/server/backend';
import type { RepresentedUser, TourDate } from '$lib/types';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, fetch, cookies }) => {
	if (!locals.token) {
		redirect(303, '/login');
	}
	// Only artists have a dashboard so far - everyone else still lands on
	// /tourdates. locals.userType is only unset for sessions predating this
	// cookie, so let those through rather than bouncing a real artist out.
	if (locals.userType && locals.userType !== 'ARTIST') {
		redirect(303, '/tourdates');
	}
	const token = locals.token;

	try {
		const [tourdates, team] = await Promise.all([
			backendFetch<TourDate[]>(fetch, '/tourdates', { token }),
			// 403s for a caller whose (possibly stale) session predates the
			// user_type cookie and turns out not to be an artist - the /dashboard
			// guard above only catches that case once a fresh login sets the
			// cookie, so fall back to an empty team rather than erroring here.
			backendFetch<RepresentedUser[]>(fetch, '/representatives', { token }).catch((err) => {
				if (err instanceof BackendError && err.status === 403) return [];
				throw err;
			})
		]);

		const today = new Date().toISOString().slice(0, 10);
		const upcoming = tourdates.filter((td) => td.date >= today).sort((a, b) => (a.date < b.date ? -1 : 1));

		return { upcoming, team };
	} catch (err) {
		if (err instanceof BackendError && err.status === 401) {
			cookies.delete('session_token', { path: '/' });
			cookies.delete('user_type', { path: '/' });
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
		cookies.delete('user_type', { path: '/' });
		redirect(303, '/login');
	}
};

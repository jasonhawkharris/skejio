import { redirect } from '@sveltejs/kit';
import { backendFetch } from '$lib/server/backend';
import type { RequestHandler } from './$types';

// A plain endpoint (not a page form action) so every page under the sidebar
// shell can point its logout button at the same "/logout" path regardless of
// which route it's currently rendering.
export const POST: RequestHandler = async ({ locals, cookies, fetch }) => {
	if (locals.token) {
		await backendFetch(fetch, '/logout', { method: 'POST', token: locals.token }).catch(() => {});
	}
	cookies.delete('session_token', { path: '/' });
	cookies.delete('user_type', { path: '/' });
	redirect(303, '/login');
};

import { dev } from '$app/environment';
import { fail, redirect } from '@sveltejs/kit';
import { BackendError, backendFetch } from '$lib/server/backend';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	if (locals.token) {
		redirect(303, locals.userType === 'ARTIST' ? '/dashboard' : '/tourdates');
	}
};

export const actions: Actions = {
	default: async ({ request, cookies, fetch }) => {
		const form = await request.formData();
		const email = form.get('email');
		const password = form.get('password');

		if (typeof email !== 'string' || typeof password !== 'string' || !email || !password) {
			return fail(400, { error: 'email and password are required' });
		}

		let result: { token: string; expires_at: string; user_type: string };
		try {
			result = await backendFetch(fetch, '/login', {
				method: 'POST',
				body: { email, password }
			});
		} catch (err) {
			if (err instanceof BackendError) {
				const body = err.body as { error?: string } | null;
				return fail(err.status, { error: body?.error ?? 'login failed' });
			}
			return fail(500, { error: 'unable to reach the server' });
		}

		const cookieOpts = {
			path: '/',
			httpOnly: true,
			sameSite: 'lax' as const,
			secure: !dev,
			expires: new Date(result.expires_at)
		};
		cookies.set('session_token', result.token, cookieOpts);
		cookies.set('user_type', result.user_type, cookieOpts);

		redirect(303, result.user_type === 'ARTIST' ? '/dashboard' : '/tourdates');
	}
};

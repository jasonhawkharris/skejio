import { dev } from '$app/environment';
import { fail, redirect } from '@sveltejs/kit';
import { BackendError, backendFetch } from '$lib/server/backend';
import type { Actions, PageServerLoad } from './$types';

const VALID_USER_TYPES = ['ARTIST', 'MANAGER', 'AGENT', 'CREW', 'LABEL'];

export const load: PageServerLoad = async ({ locals }) => {
	if (locals.token) {
		redirect(303, locals.userType === 'ARTIST' ? '/dashboard' : '/tourdates');
	}
};

export const actions: Actions = {
	default: async ({ request, cookies, fetch }) => {
		const form = await request.formData();
		const name = form.get('name');
		const email = form.get('email');
		const password = form.get('password');
		const userType = form.get('user_type');

		if (typeof name !== 'string' || !name) {
			return fail(400, { error: 'name is required' });
		}
		if (typeof email !== 'string' || !email) {
			return fail(400, { error: 'email is required' });
		}
		if (typeof password !== 'string' || !password) {
			return fail(400, { error: 'password is required' });
		}
		if (typeof userType !== 'string' || !VALID_USER_TYPES.includes(userType)) {
			return fail(400, { error: 'a valid account type is required' });
		}

		try {
			await backendFetch(fetch, '/sign-up', {
				method: 'POST',
				body: { name, email, password, user_type: userType }
			});
		} catch (err) {
			if (err instanceof BackendError) {
				const body = err.body as { error?: string } | null;
				return fail(err.status, { error: body?.error ?? 'sign-up failed' });
			}
			return fail(500, { error: 'unable to reach the server' });
		}

		let loginResult: { token: string; expires_at: string; user_type: string };
		try {
			loginResult = await backendFetch(fetch, '/login', {
				method: 'POST',
				body: { email, password }
			});
		} catch {
			redirect(303, '/login');
		}

		const cookieOpts = {
			path: '/',
			httpOnly: true,
			sameSite: 'lax' as const,
			secure: !dev,
			expires: new Date(loginResult.expires_at)
		};
		cookies.set('session_token', loginResult.token, cookieOpts);
		cookies.set('user_type', loginResult.user_type, cookieOpts);

		redirect(303, loginResult.user_type === 'ARTIST' ? '/dashboard' : '/tourdates');
	}
};

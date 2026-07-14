import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.token = event.cookies.get('session_token');
	return resolve(event);
};

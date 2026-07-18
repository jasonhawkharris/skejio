import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.token = event.cookies.get('session_token');
	event.locals.userType = event.cookies.get('user_type');
	event.locals.userName = event.cookies.get('user_name');
	return resolve(event);
};

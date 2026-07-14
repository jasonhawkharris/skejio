import { env } from '$env/dynamic/private';

const BACKEND_URL = env.BACKEND_URL ?? 'http://localhost:8080';

export class BackendError extends Error {
	status: number;
	body: unknown;

	constructor(status: number, body: unknown) {
		super(`backend request failed with status ${status}`);
		this.status = status;
		this.body = body;
	}
}

type BackendFetchOptions = Omit<RequestInit, 'body'> & {
	token?: string;
	body?: unknown;
};

// backendFetch proxies a request to the Go API from server-side code only
// (load functions, form actions) - never called from the browser, so the
// session token never leaves the server.
export async function backendFetch<T>(
	fetch: typeof globalThis.fetch,
	path: string,
	{ token, body, headers, ...rest }: BackendFetchOptions = {}
): Promise<T> {
	const response = await fetch(`${BACKEND_URL}${path}`, {
		...rest,
		headers: {
			'Content-Type': 'application/json',
			...(token ? { Authorization: `Bearer ${token}` } : {}),
			...headers
		},
		body: body !== undefined ? JSON.stringify(body) : undefined
	});

	if (response.status === 204) {
		return null as T;
	}

	const data = await response.json().catch(() => null);

	if (!response.ok) {
		throw new BackendError(response.status, data);
	}

	return data as T;
}

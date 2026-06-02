import type { CommandEntry, RconSession, User } from './types';

const BASE = typeof window !== 'undefined' && import.meta.env.DEV ? 'http://localhost:8080' : '';

async function request<T>(
	method: string,
	path: string,
	body?: unknown
): Promise<{ data?: T; error?: string }> {
	try {
		const res = await fetch(BASE + path, {
			method,
			credentials: 'include',
			headers: body ? { 'Content-Type': 'application/json' } : {},
			body: body ? JSON.stringify(body) : undefined
		});
		const json = await res.json();
		if (!res.ok) return { error: json.error ?? 'request failed' };
		return { data: json as T };
	} catch {
		return { error: 'network error' };
	}
}

async function upload<T>(path: string, file: File): Promise<{ data?: T; error?: string }> {
	try {
		const form = new FormData();
		form.append('file', file);
		const res = await fetch(BASE + path, { method: 'POST', credentials: 'include', body: form });
		const json = await res.json();
		if (!res.ok) return { error: json.error ?? 'upload failed' };
		return { data: json as T };
	} catch {
		return { error: 'network error' };
	}
}

export const api = {
	auth: {
		login: (username: string, password: string) =>
			request<{ user: User }>('POST', '/api/auth/login', { username, password }),
		logout: () => request('POST', '/api/auth/logout'),
		me: () => request<{ user: User }>('GET', '/api/auth/me')
	},

	sessions: {
		list: () => request<{ sessions: RconSession[] }>('GET', '/api/sessions'),
		create: (data: { name: string; icon: string; host: string; port: number; password: string }) =>
			request<RconSession>('POST', '/api/sessions', data),
		update: (
			id: string,
			data: { name: string; icon: string; host: string; port: number; password?: string }
		) => request('PUT', `/api/sessions/${id}`, data),
		delete: (id: string) => request('DELETE', `/api/sessions/${id}`),
		test: (id: string) =>
			request<{ success: boolean; message: string }>('POST', `/api/sessions/${id}/test`),
		connect: (id: string) =>
			request<{ success: boolean; message?: string }>('POST', `/api/sessions/${id}/connect`),
		disconnect: (id: string) => request('POST', `/api/sessions/${id}/disconnect`),
		history: (id: string, limit = 200) =>
			request<{ history: CommandEntry[] }>('GET', `/api/sessions/${id}/history?limit=${limit}`),
		clearHistory: (id: string) => request('DELETE', `/api/sessions/${id}/history`)
	},

	icons: {
		upload: (file: File) => upload<{ url: string }>('/api/icons', file)
	}
};

export function createWsUrl(sessionId: string): string {
	const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	const host = import.meta.env.DEV ? 'localhost:8080' : window.location.host;
	return `${proto}//${host}/api/sessions/${sessionId}/ws`;
}

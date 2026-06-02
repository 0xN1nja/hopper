export interface User {
	id: string;
	username: string;
	createdAt: string;
}

export interface RconSession {
	id: string;
	name: string;
	icon: string;
	host: string;
	port: number;
	createdAt: string;
	connected: boolean;
	subscribers: number;
}

export function isImageUrl(s: string): boolean {
	return (
		s.startsWith('/') ||
		s.startsWith('http://') ||
		s.startsWith('https://') ||
		s.startsWith('data:')
	);
}

export interface CommandEntry {
	id: string;
	rconSessionId: string;
	userId: string;
	command: string;
	output: string;
	executedAt: string;
}

export interface WsMessage {
	type: 'connected' | 'disconnected' | 'output' | 'error';
	id?: string;
	command?: string;
	output?: string;
	message?: string;
	timestamp?: string;
	userId?: string;
}

import { api } from '$lib/api';
import type { User } from '$lib/types';

let user = $state<User | null>(null);
let loading = $state(true);

async function init() {
	const { data } = await api.auth.me();
	user = data?.user ?? null;
	loading = false;
}

export const authStore = {
	get user() {
		return user;
	},
	get loading() {
		return loading;
	},
	init,
	setUser(u: User | null) {
		user = u;
	}
};

<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { authStore } from '$lib/stores/auth.svelte';
	import { Loader2 } from 'lucide-svelte';

	let username = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit() {
		if (!username || !password) return;
		loading = true;
		error = '';
		const { data, error: err } = await api.auth.login(username, password);
		loading = false;
		if (err || !data) {
			error = err ?? 'Login failed';
			return;
		}
		authStore.setUser(data.user);
		goto('/');
	}
</script>

<div class="flex min-h-screen flex-col items-center justify-center bg-zinc-50 px-4">
	<div class="mb-8 flex flex-col items-center gap-3">
		<img src="/hopper.png" alt="Hopper" class="h-14 w-14 object-contain" />
		<div class="text-center">
			<h1 class="text-xl font-semibold tracking-tight text-zinc-900">Hopper</h1>
			<p class="mt-1 text-sm text-zinc-500">RCON management, self-hosted.</p>
		</div>
	</div>

	<div
		class="w-full max-w-sm rounded-2xl border border-zinc-200 bg-white p-6 shadow-[0_1px_3px_rgba(0,0,0,0.04)]"
	>
		<form
			onsubmit={(e) => {
				e.preventDefault();
				handleSubmit();
			}}
			class="flex flex-col gap-4"
		>
			<div>
				<label for="username" class="mb-1.5 block text-xs font-medium text-zinc-600">Username</label
				>
				<input
					id="username"
					type="text"
					bind:value={username}
					required
					placeholder="admin"
					autocomplete="username"
					class="h-10 w-full rounded-lg border border-zinc-200 bg-white px-3 text-sm text-zinc-900 placeholder-zinc-400 transition-colors focus:border-blue-500 focus:outline-none"
				/>
			</div>

			<div>
				<label for="password" class="mb-1.5 block text-xs font-medium text-zinc-600">Password</label
				>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					placeholder="••••••••"
					autocomplete="current-password"
					class="h-10 w-full rounded-lg border border-zinc-200 bg-white px-3 text-sm text-zinc-900 placeholder-zinc-400 transition-colors focus:border-blue-500 focus:outline-none"
				/>
			</div>

			{#if error}
				<p class="rounded-lg border border-red-100 bg-red-50 px-3 py-2 text-xs text-red-600">
					{error}
				</p>
			{/if}

			<button
				type="submit"
				disabled={loading}
				class="flex h-10 items-center justify-center gap-2 rounded-xl bg-blue-600 text-sm font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-60"
			>
				{#if loading}
					<Loader2 size={14} class="animate-spin" />
				{/if}
				Sign in
			</button>
		</form>
	</div>

	<p class="mt-6 text-xs text-zinc-400">
		Account credentials are configured via environment variables.
	</p>
</div>

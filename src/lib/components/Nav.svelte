<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { authStore } from '$lib/stores/auth.svelte';
	import { LogOut } from 'lucide-svelte';

	async function handleLogout() {
		await api.auth.logout();
		authStore.setUser(null);
		goto('/login');
	}
</script>

<header class="sticky top-0 z-40 border-b border-zinc-200 bg-white/95 backdrop-blur-sm">
	<div class="mx-auto flex h-14 max-w-screen-xl items-center justify-between px-6">
		<a href="/" class="flex items-center gap-2.5">
			<img src="/hopper.png" alt="Hopper" class="h-7 w-7 object-contain" />
			<span class="text-sm font-semibold tracking-tight text-zinc-900">Hopper</span>
		</a>

		{#if authStore.user}
			<div class="flex items-center gap-3">
				<span class="text-sm text-zinc-500">{authStore.user.username}</span>
				<button
					onclick={handleLogout}
					class="flex h-8 w-8 items-center justify-center rounded-lg text-zinc-400 transition-colors duration-150 hover:bg-zinc-100 hover:text-zinc-700"
					title="Sign out"
				>
					<LogOut size={15} />
				</button>
			</div>
		{/if}
	</div>
</header>

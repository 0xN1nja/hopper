<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import Terminal from '$lib/components/Terminal.svelte';
	import { authStore } from '$lib/stores/auth.svelte';
	import { isImageUrl, type RconSession } from '$lib/types';
	import { ArrowLeft, Loader2, PlugZap, WifiOff } from 'lucide-svelte';
	import { onMount } from 'svelte';

	let sessionId = $derived($page.params.id ?? '');
	let session = $state<RconSession | null>(null);
	let loading = $state(true);
	let disconnecting = $state(false);
	let currentUser = $derived(authStore.user);

	onMount(async () => {
		const { data } = await api.sessions.list();
		if (data?.sessions) {
			session = data.sessions.find((s) => s.id === sessionId) ?? null;
		}
		loading = false;
		if (!session) goto('/');
	});

	async function handleDisconnect() {
		if (!session) return;
		disconnecting = true;
		await api.sessions.disconnect(session.id);
		disconnecting = false;
		goto('/');
	}
</script>

<div class="flex h-screen flex-col overflow-hidden bg-zinc-50">
	<div class="flex items-center justify-between border-b border-zinc-200 bg-white px-6 py-3">
		<div class="flex items-center gap-3">
			<button
				onclick={() => goto('/')}
				class="flex h-8 w-8 items-center justify-center rounded-lg text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-700"
			>
				<ArrowLeft size={15} />
			</button>
			{#if session}
				{#if session.icon && isImageUrl(session.icon)}
					<img src={session.icon} alt="" class="h-7 w-7 rounded-lg object-cover" />
				{:else if session.icon}
					<span class="text-lg">{session.icon}</span>
				{/if}
				<div>
					<h1 class="text-sm font-semibold tracking-tight text-zinc-900">{session.name}</h1>
					<p class="font-mono text-xs text-zinc-400">{session.host}:{session.port}</p>
				</div>
			{:else if loading}
				<div class="h-4 w-32 animate-pulse rounded bg-zinc-200"></div>
			{/if}
		</div>

		{#if session}
			<div class="flex items-center gap-2">
				{#if session.connected}
					<span class="flex items-center gap-1.5 text-xs font-medium text-emerald-600">
						<PlugZap size={12} />
						Active
					</span>
				{:else}
					<span class="flex items-center gap-1.5 text-xs text-zinc-400">
						<WifiOff size={12} />
						Disconnected
					</span>
				{/if}
				<button
					onclick={handleDisconnect}
					disabled={disconnecting}
					class="flex h-8 items-center gap-1.5 rounded-lg border border-zinc-200 px-3 text-xs font-medium text-zinc-600 transition-colors hover:bg-zinc-100 disabled:opacity-50"
				>
					{#if disconnecting}
						<Loader2 size={11} class="animate-spin" />
					{:else}
						<WifiOff size={11} />
					{/if}
					Disconnect
				</button>
			</div>
		{/if}
	</div>

	<div class="flex-1 overflow-hidden p-4">
		{#if !loading && session && currentUser}
			<Terminal {sessionId} userId={currentUser.id} />
		{:else if loading}
			<div class="flex h-full items-center justify-center">
				<Loader2 size={18} class="animate-spin text-zinc-400" />
			</div>
		{/if}
	</div>
</div>

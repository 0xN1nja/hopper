<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { isImageUrl, type RconSession } from '$lib/types';
	import { Loader2, Plug, PlugZap, Trash2, Wifi, WifiOff } from 'lucide-svelte';

	let {
		session,
		ondelete,
		onedit,
		ondisconnect
	}: {
		session: RconSession;
		ondelete: () => void;
		onedit: () => void;
		ondisconnect: () => void;
	} = $props();

	let confirmingDelete = $state(false);
	let testing = $state(false);
	let connecting = $state(false);
	let disconnecting = $state(false);
	let testMessage = $state('');
	let testOk = $state(false);

	async function handleTest() {
		testing = true;
		testMessage = '';
		const { data } = await api.sessions.test(session.id);
		testing = false;
		testOk = data?.success ?? false;
		testMessage = data?.message ?? (testOk ? 'Reachable' : 'Unreachable');
	}

	async function handleConnect() {
		connecting = true;
		const { data } = await api.sessions.connect(session.id);
		connecting = false;
		if (data?.success) {
			goto(`/session/${session.id}`);
		}
	}

	async function handleDisconnect() {
		disconnecting = true;
		await api.sessions.disconnect(session.id);
		disconnecting = false;
		ondisconnect();
	}
</script>

<div class="group relative rounded-2xl border border-zinc-200 bg-white p-5">
	<div class="flex items-start justify-between">
		<div class="flex items-center gap-3">
			<div
				class="flex h-10 w-10 items-center justify-center overflow-hidden rounded-xl border border-zinc-100 bg-zinc-50"
			>
				{#if session.icon && isImageUrl(session.icon)}
					<img src={session.icon} alt="" class="h-full w-full object-cover" />
				{:else if session.icon}
					<span class="text-xl">{session.icon}</span>
				{:else}
					<span class="text-zinc-400">
						<svg
							width="18"
							height="18"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="1.5"
							><rect x="2" y="3" width="20" height="14" rx="2" /><path d="M8 21h8M12 17v4" /></svg
						>
					</span>
				{/if}
			</div>
			<div>
				<h3 class="text-sm font-semibold tracking-tight text-zinc-900">{session.name}</h3>
				<p class="mt-0.5 font-mono text-xs text-zinc-400">{session.host}:{session.port}</p>
			</div>
		</div>

		<div class="flex items-center gap-1">
			{#if confirmingDelete}
				<div class="flex items-center gap-1.5 rounded-lg border border-red-100 bg-red-50 px-2 py-1">
					<span class="text-xs text-red-600">Remove?</span>
					<button
						onclick={ondelete}
						class="text-xs font-medium text-red-600 transition-colors hover:text-red-700"
					>
						Yes
					</button>
					<span class="text-red-300">·</span>
					<button
						onclick={() => (confirmingDelete = false)}
						class="text-xs text-zinc-400 transition-colors hover:text-zinc-600"
					>
						Cancel
					</button>
				</div>
			{:else}
				<div class="flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
					<button
						onclick={onedit}
						class="flex h-7 w-7 items-center justify-center rounded-lg text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-600"
						title="Edit session"
					>
						<svg
							width="13"
							height="13"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z" />
						</svg>
					</button>
					<button
						onclick={() => (confirmingDelete = true)}
						class="flex h-7 w-7 items-center justify-center rounded-lg text-zinc-400 transition-colors hover:bg-red-50 hover:text-red-500"
						title="Delete session"
					>
						<Trash2 size={13} />
					</button>
				</div>
			{/if}
		</div>
	</div>

	<div class="mt-4 space-y-2">
		<div class="flex items-center gap-2">
			{#if session.connected}
				<span class="flex items-center gap-1.5 text-xs font-medium text-emerald-600">
					<span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
					Connected
					{#if session.subscribers > 0}
						<span class="text-zinc-400">· {session.subscribers} active</span>
					{/if}
				</span>
				<div class="ml-auto flex gap-2">
					<button
						onclick={handleDisconnect}
						disabled={disconnecting}
						class="flex h-7 items-center gap-1.5 rounded-lg border border-zinc-200 px-3 text-xs font-medium text-zinc-600 transition-colors hover:bg-zinc-50 disabled:opacity-50"
					>
						{#if disconnecting}
							<Loader2 size={11} class="animate-spin" />
						{:else}
							<WifiOff size={11} />
						{/if}
						Disconnect
					</button>
					<button
						onclick={() => goto(`/session/${session.id}`)}
						class="flex h-7 items-center gap-1.5 rounded-lg bg-blue-600 px-3 text-xs font-medium text-white transition-colors hover:bg-blue-700"
					>
						<PlugZap size={12} />
						Open
					</button>
				</div>
			{:else}
				<span class="flex items-center gap-1.5 text-xs text-zinc-400">
					<WifiOff size={12} />
					Disconnected
				</span>
				<div class="ml-auto flex gap-2">
					<button
						onclick={handleTest}
						disabled={testing}
						class="flex h-7 items-center gap-1.5 rounded-lg border border-zinc-200 px-3 text-xs font-medium text-zinc-600 transition-colors hover:border-zinc-300 hover:bg-zinc-50 disabled:opacity-50"
					>
						{#if testing}
							<Loader2 size={11} class="animate-spin" />
						{:else}
							<Wifi size={11} />
						{/if}
						Test
					</button>
					<button
						onclick={handleConnect}
						disabled={connecting}
						class="flex h-7 items-center gap-1.5 rounded-lg bg-zinc-900 px-3 text-xs font-medium text-white transition-colors hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-40"
					>
						{#if connecting}
							<Loader2 size={11} class="animate-spin" />
						{:else}
							<Plug size={11} />
						{/if}
						Connect
					</button>
				</div>
			{/if}
		</div>

		{#if testMessage}
			<p class="text-xs {testOk ? 'text-emerald-600' : 'text-red-500'}">{testMessage}</p>
		{/if}
	</div>
</div>

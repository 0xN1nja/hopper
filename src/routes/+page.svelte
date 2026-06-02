<script lang="ts">
	import { api } from '$lib/api';
	import Modal from '$lib/components/Modal.svelte';
	import SessionCard from '$lib/components/SessionCard.svelte';
	import SessionForm from '$lib/components/SessionForm.svelte';
	import { authStore } from '$lib/stores/auth.svelte';
	import type { RconSession } from '$lib/types';
	import { Loader2, Plus, Server, WifiOff } from 'lucide-svelte';
	import { onMount } from 'svelte';

	let sessions = $state<RconSession[]>([]);
	let loading = $state(true);
	let createOpen = $state(false);
	let editOpen = $state(false);
	let editingSession = $state<RconSession | null>(null);
	let submitting = $state(false);
	let disconnectingAll = $state(false);

	let newName = $state('');
	let newIcon = $state('');
	let newHost = $state('');
	let newPort = $state(25575);
	let newPassword = $state('');

	let editName = $state('');
	let editIcon = $state('');
	let editHost = $state('');
	let editPort = $state(25575);
	let editPassword = $state('');

	let anyConnected = $derived(sessions.some((s) => s.connected));

	onMount(loadSessions);

	async function loadSessions() {
		loading = true;
		const { data } = await api.sessions.list();
		sessions = data?.sessions ?? [];
		loading = false;
	}

	function openCreate() {
		newName = '';
		newIcon = '';
		newHost = '';
		newPort = 25575;
		newPassword = '';
		createOpen = true;
	}

	function openEdit(session: RconSession) {
		editingSession = session;
		editName = session.name;
		editIcon = session.icon || '';
		editHost = session.host;
		editPort = session.port;
		editPassword = '';
		editOpen = true;
	}

	async function handleCreate() {
		submitting = true;
		const { data, error } = await api.sessions.create({
			name: newName,
			icon: newIcon,
			host: newHost,
			port: newPort,
			password: newPassword
		});
		submitting = false;
		if (error) return;
		sessions = [data as RconSession, ...sessions];
		createOpen = false;
	}

	async function handleEdit() {
		if (!editingSession) return;
		submitting = true;
		const { error } = await api.sessions.update(editingSession.id, {
			name: editName,
			icon: editIcon,
			host: editHost,
			port: editPort,
			password: editPassword
		});
		submitting = false;
		if (error) return;
		sessions = sessions.map((s) =>
			s.id === editingSession!.id
				? { ...s, name: editName, icon: editIcon, host: editHost, port: editPort }
				: s
		);
		editOpen = false;
	}

	async function handleDelete(id: string) {
		await api.sessions.delete(id);
		sessions = sessions.filter((s) => s.id !== id);
	}

	async function handleDisconnect(id: string) {
		sessions = sessions.map((s) => (s.id === id ? { ...s, connected: false, subscribers: 0 } : s));
	}

	async function disconnectAll() {
		disconnectingAll = true;
		await Promise.all(
			sessions.filter((s) => s.connected).map((s) => api.sessions.disconnect(s.id))
		);
		sessions = sessions.map((s) => ({ ...s, connected: false, subscribers: 0 }));
		disconnectingAll = false;
	}
</script>

<main class="mx-auto max-w-screen-xl px-6 py-10">
	<div class="mb-8 flex items-end justify-between">
		<div>
			<h1 class="text-xl font-semibold tracking-tight text-zinc-900">Sessions</h1>
			<p class="mt-1 text-sm text-zinc-500">Manage and connect to your RCON servers.</p>
		</div>
		<div class="flex items-center gap-2">
			<button
				onclick={disconnectAll}
				disabled={!anyConnected || disconnectingAll}
				class="flex h-9 items-center gap-2 rounded-xl border border-zinc-200 px-4 text-sm font-medium text-zinc-600 transition-colors hover:bg-zinc-50 disabled:cursor-not-allowed disabled:opacity-40"
			>
				{#if disconnectingAll}
					<Loader2 size={14} class="animate-spin" />
				{:else}
					<WifiOff size={14} />
				{/if}
				Disconnect all
			</button>
			<button
				onclick={openCreate}
				class="flex h-9 items-center gap-2 rounded-xl bg-blue-600 px-4 text-sm font-medium text-white transition-colors hover:bg-blue-700"
			>
				<Plus size={15} />
				New Session
			</button>
		</div>
	</div>

	{#if loading || authStore.loading}
		<div class="flex items-center gap-2 text-sm text-zinc-400">
			<Loader2 size={14} class="animate-spin" />
			Loading…
		</div>
	{:else if sessions.length === 0}
		<div
			class="flex flex-col items-center justify-center rounded-2xl border border-dashed border-zinc-200 py-20 text-center"
		>
			<div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-zinc-100">
				<Server size={22} class="text-zinc-400" />
			</div>
			<p class="mt-4 text-sm font-medium text-zinc-700">No sessions yet</p>
			<p class="mt-1 text-sm text-zinc-400">Create your first RCON session to get started.</p>
			<button
				onclick={openCreate}
				class="mt-6 flex h-9 items-center gap-2 rounded-xl bg-blue-600 px-4 text-sm font-medium text-white transition-colors hover:bg-blue-700"
			>
				<Plus size={15} />
				New Session
			</button>
		</div>
	{:else}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each sessions as session (session.id)}
				<SessionCard
					{session}
					ondelete={() => handleDelete(session.id)}
					onedit={() => openEdit(session)}
					ondisconnect={() => handleDisconnect(session.id)}
				/>
			{/each}
		</div>
	{/if}
</main>

<Modal bind:open={createOpen} title="New Session">
	{#snippet children()}
		<SessionForm
			bind:name={newName}
			bind:icon={newIcon}
			bind:host={newHost}
			bind:port={newPort}
			bind:password={newPassword}
			{submitting}
			onsubmit={handleCreate}
		/>
	{/snippet}
</Modal>

<Modal bind:open={editOpen} title="Edit Session">
	{#snippet children()}
		<SessionForm
			bind:name={editName}
			bind:icon={editIcon}
			bind:host={editHost}
			bind:port={editPort}
			bind:password={editPassword}
			passwordPlaceholder="Leave blank to keep current"
			{submitting}
			onsubmit={handleEdit}
		/>
	{/snippet}
</Modal>

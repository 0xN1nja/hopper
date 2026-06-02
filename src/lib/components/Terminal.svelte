<script lang="ts">
	import { api, createWsUrl } from '$lib/api';
	import type { CommandEntry, WsMessage } from '$lib/types';
	import { ChevronUp, Eraser, Loader2, WifiOff } from 'lucide-svelte';

	let { sessionId, userId }: { sessionId: string; userId: string } = $props();

	type OutputEntry = {
		id: string;
		command?: string;
		output: string;
		timestamp: Date;
		isError?: boolean;
		userId?: string;
	};

	let entries = $state<OutputEntry[]>([]);
	let input = $state('');
	let connected = $state(false);
	let connecting = $state(false);
	let historyIndex = $state(-1);
	let inputHistory = $state<string[]>([]);

	let outputEl: HTMLDivElement;
	let ws: WebSocket | null = null;

	$effect(() => {
		loadHistory();
		connectWs();
		return () => {
			ws?.close();
		};
	});

	$effect(() => {
		if (entries.length && outputEl) {
			scrollToBottom();
		}
	});

	function scrollToBottom() {
		requestAnimationFrame(() => {
			if (outputEl) outputEl.scrollTop = outputEl.scrollHeight;
		});
	}

	async function loadHistory() {
		const { data } = await api.sessions.history(sessionId);
		if (data?.history) {
			entries = data.history.map((e: CommandEntry) => ({
				id: e.id,
				command: e.command,
				output: e.output,
				timestamp: new Date(e.executedAt),
				userId: e.userId
			}));
		}
	}

	function connectWs() {
		connecting = true;
		const url = createWsUrl(sessionId);
		ws = new WebSocket(url);

		ws.onopen = () => {
			connecting = false;
		};

		ws.onmessage = (e) => {
			const msg: WsMessage = JSON.parse(e.data);
			handleMessage(msg);
		};

		ws.onerror = () => {
			connecting = false;
			connected = false;
		};

		ws.onclose = () => {
			connecting = false;
			connected = false;
			ws = null;
		};
	}

	function handleMessage(msg: WsMessage) {
		if (msg.type === 'connected') {
			connected = true;
			connecting = false;
		} else if (msg.type === 'disconnected') {
			connected = false;
		} else if (msg.type === 'output') {
			const existing = entries.findIndex((e) => e.id === msg.id);
			if (existing === -1) {
				entries = [
					...entries,
					{
						id: msg.id!,
						command: msg.command,
						output: msg.output ?? '',
						timestamp: new Date(msg.timestamp ?? Date.now()),
						userId: msg.userId
					}
				];
			}
		} else if (msg.type === 'error') {
			entries = [
				...entries,
				{
					id: crypto.randomUUID(),
					output: msg.message ?? 'unknown error',
					timestamp: new Date(),
					isError: true
				}
			];
		}
	}

	function sendCommand() {
		const cmd = input.trim();
		if (!cmd || !ws || ws.readyState !== WebSocket.OPEN) return;

		ws.send(JSON.stringify({ type: 'command', data: cmd }));
		inputHistory = [cmd, ...inputHistory.slice(0, 99)];
		input = '';
		historyIndex = -1;
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			sendCommand();
			return;
		}

		if (e.key === 'ArrowUp') {
			e.preventDefault();
			if (historyIndex < inputHistory.length - 1) {
				historyIndex++;
				input = inputHistory[historyIndex];
			}
			return;
		}

		if (e.key === 'ArrowDown') {
			e.preventDefault();
			if (historyIndex > 0) {
				historyIndex--;
				input = inputHistory[historyIndex];
			} else {
				historyIndex = -1;
				input = '';
			}
			return;
		}
	}

	function onInput() {
		historyIndex = -1;
	}

	async function handleClear() {
		await api.sessions.clearHistory(sessionId);
		entries = [];
	}

	function formatTime(d: Date) {
		return d.toLocaleTimeString(undefined, { hour12: false });
	}

	function isOwnEntry(entry: OutputEntry) {
		return entry.userId === userId;
	}
</script>

<div class="flex h-full flex-col overflow-hidden rounded-2xl border border-zinc-200 bg-zinc-950">
	<div class="flex items-center justify-between border-b border-zinc-800 px-4 py-2.5">
		<div class="flex items-center gap-2">
			{#if connecting}
				<span class="flex items-center gap-1.5 text-xs text-zinc-500">
					<Loader2 size={11} class="animate-spin text-zinc-500" />
					Connecting…
				</span>
			{:else if connected}
				<span class="flex items-center gap-1.5 text-xs text-emerald-500">
					<span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
					Connected
				</span>
			{:else}
				<span class="flex items-center gap-1.5 text-xs text-zinc-500">
					<WifiOff size={11} />
					Disconnected
				</span>
			{/if}
		</div>
		<button
			onclick={handleClear}
			class="flex items-center gap-1.5 rounded-lg px-2.5 py-1 text-xs text-zinc-500 transition-colors hover:bg-zinc-800 hover:text-zinc-300"
		>
			<Eraser size={11} />
			Clear
		</button>
	</div>

	<div
		bind:this={outputEl}
		class="terminal-font flex-1 overflow-y-auto p-4 text-zinc-300 scroll-smooth"
	>
		{#if entries.length === 0}
			<p class="text-xs text-zinc-600">No output yet. Run a command to get started.</p>
		{/if}
		{#each entries as entry (entry.id)}
			<div class="mb-3 {isOwnEntry(entry) ? '' : 'opacity-80'}">
				{#if entry.command}
					<div class="flex items-baseline gap-2">
						<span class="shrink-0 text-zinc-600">{formatTime(entry.timestamp)}</span>
						<span class="text-zinc-400">›</span>
						<span class="text-zinc-200">{entry.command}</span>
					</div>
				{/if}
				{#if entry.output}
					<div
						class="mt-0.5 whitespace-pre-wrap pl-18 {entry.isError
							? 'text-red-400'
							: 'text-zinc-400'}"
					>
						{entry.output}
					</div>
				{/if}
			</div>
		{/each}
	</div>

	<div class="border-t border-zinc-800">
		<div class="flex items-center gap-2 p-3">
			<span class="terminal-font shrink-0 text-zinc-600">›</span>
			<input
				bind:value={input}
				onkeydown={onKeydown}
				oninput={onInput}
				disabled={!connected}
				placeholder={connected ? 'Enter command…' : 'Not connected'}
				class="terminal-font min-w-0 flex-1 bg-transparent text-zinc-200 placeholder-zinc-700 focus:outline-none disabled:cursor-not-allowed"
				autocomplete="off"
				autocorrect="off"
				autocapitalize="off"
				spellcheck={false}
			/>
			{#if inputHistory.length > 0}
				<ChevronUp size={13} class="shrink-0 text-zinc-700" />
			{/if}
		</div>
	</div>
</div>

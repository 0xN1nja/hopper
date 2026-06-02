<script lang="ts">
	import { api } from '$lib/api';
	import { isImageUrl } from '$lib/types';
	import { Loader2 } from 'lucide-svelte';

	let {
		name = $bindable(''),
		icon = $bindable(''),
		host = $bindable(''),
		port = $bindable(25575),
		password = $bindable(''),
		passwordPlaceholder = 'RCON password',
		submitting = false,
		onsubmit
	}: {
		name: string;
		icon: string;
		host: string;
		port: number;
		password: string;
		passwordPlaceholder?: string;
		submitting?: boolean;
		onsubmit: () => void;
	} = $props();

	let uploading = $state(false);
	let urlDraft = $state(icon.startsWith('http://') || icon.startsWith('https://') ? icon : '');
	let fileInput: HTMLInputElement;

	$effect(() => {
		if (!icon) urlDraft = '';
	});

	async function onFileChange(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		uploading = true;
		const { data } = await api.icons.upload(file);
		uploading = false;
		if (data?.url) {
			icon = data.url;
			urlDraft = '';
		}
		(e.target as HTMLInputElement).value = '';
	}

	function applyUrl() {
		const url = urlDraft.trim();
		if (url) icon = url;
	}
</script>

<form
	onsubmit={(e) => {
		e.preventDefault();
		onsubmit();
	}}
	class="flex flex-col gap-4"
>
	<div>
		<label for="sf-name" class="mb-1.5 block text-xs font-medium text-zinc-600">Name</label>
		<input
			id="sf-name"
			bind:value={name}
			required
			placeholder="My Minecraft Server"
			class="h-9 w-full rounded-lg border border-zinc-200 bg-white px-3 text-sm text-zinc-900 placeholder-zinc-400 transition-colors focus:border-blue-500"
		/>
	</div>

	<div>
		<span class="mb-1.5 block text-xs font-medium text-zinc-600">Icon</span>
		<div class="flex items-center gap-2">
			<button
				type="button"
				onclick={() => fileInput.click()}
				class="group relative flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-zinc-200 bg-zinc-50 transition-colors hover:border-zinc-300 hover:bg-zinc-100"
				title="Upload image"
			>
				{#if uploading}
					<Loader2 size={13} class="animate-spin text-zinc-400" />
				{:else if icon && isImageUrl(icon)}
					<img src={icon} alt="" class="h-full w-full object-cover" />
					<div
						class="absolute inset-0 flex items-center justify-center bg-black/0 transition-colors group-hover:bg-black/35"
					>
						<svg
							width="11"
							height="11"
							viewBox="0 0 24 24"
							fill="none"
							stroke="white"
							stroke-width="2.5"
							stroke-linecap="round"
							stroke-linejoin="round"
							class="opacity-0 transition-opacity group-hover:opacity-100"
						>
							<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
							<polyline points="17 8 12 3 7 8" />
							<line x1="12" y1="3" x2="12" y2="15" />
						</svg>
					</div>
				{:else}
					<svg
						width="14"
						height="14"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="1.5"
						class="text-zinc-400"
					>
						<rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
						<circle cx="8.5" cy="8.5" r="1.5" />
						<polyline points="21 15 16 10 5 21" />
					</svg>
				{/if}
			</button>

			<input
				type="text"
				bind:value={urlDraft}
				onkeydown={(e) => {
					if (e.key === 'Enter') {
						e.preventDefault();
						applyUrl();
					}
				}}
				placeholder="or paste an image URL"
				class="h-9 min-w-0 flex-1 rounded-lg border border-zinc-200 bg-white px-3 text-sm text-zinc-900 placeholder-zinc-400 transition-colors focus:border-blue-500"
			/>
			{#if urlDraft.trim()}
				<button
					type="button"
					onclick={applyUrl}
					class="flex h-9 shrink-0 items-center rounded-lg bg-zinc-900 px-3 text-xs font-medium text-white transition-colors hover:bg-zinc-800"
				>
					Fetch
				</button>
			{/if}
		</div>
		<input
			bind:this={fileInput}
			type="file"
			accept="image/*"
			class="hidden"
			onchange={onFileChange}
		/>
	</div>

	<div class="grid grid-cols-3 gap-3">
		<div class="col-span-2">
			<label for="sf-host" class="mb-1.5 block text-xs font-medium text-zinc-600">Host</label>
			<input
				id="sf-host"
				bind:value={host}
				required
				placeholder="127.0.0.1"
				class="h-9 w-full rounded-lg border border-zinc-200 bg-white px-3 font-mono text-sm text-zinc-900 placeholder-zinc-400 transition-colors focus:border-blue-500"
			/>
		</div>
		<div>
			<label for="sf-port" class="mb-1.5 block text-xs font-medium text-zinc-600">Port</label>
			<input
				id="sf-port"
				bind:value={port}
				required
				type="number"
				min="1"
				max="65535"
				placeholder="25575"
				class="h-9 w-full rounded-lg border border-zinc-200 bg-white px-3 font-mono text-sm text-zinc-900 placeholder-zinc-400 transition-colors focus:border-blue-500"
			/>
		</div>
	</div>

	<div>
		<label for="sf-password" class="mb-1.5 block text-xs font-medium text-zinc-600">Password</label>
		<input
			id="sf-password"
			bind:value={password}
			type="password"
			placeholder={passwordPlaceholder}
			class="h-9 w-full rounded-lg border border-zinc-200 bg-white px-3 text-sm text-zinc-900 placeholder-zinc-400 transition-colors focus:border-blue-500"
		/>
	</div>

	<button
		type="submit"
		disabled={submitting}
		class="flex h-9 items-center justify-center gap-2 rounded-lg bg-blue-600 text-sm font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-60"
	>
		{#if submitting}
			<Loader2 size={14} class="animate-spin" />
		{/if}
		Save Session
	</button>
</form>

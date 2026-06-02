<script lang="ts">
	import { X } from 'lucide-svelte';

	let {
		open = $bindable(false),
		title,
		children
	}: {
		open: boolean;
		title: string;
		children: import('svelte').Snippet;
	} = $props();

	function close() {
		open = false;
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}
</script>

<svelte:window onkeydown={onKeydown} />

{#if open}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4"
		role="dialog"
		aria-modal="true"
	>
		<button
			class="absolute inset-0 bg-zinc-950/30"
			onclick={close}
			tabindex="-1"
			aria-label="Close modal"
		></button>

		<div
			class="relative w-full max-w-md rounded-2xl border border-zinc-200 bg-white shadow-[0_8px_32px_rgba(0,0,0,0.08)]"
		>
			<div class="flex items-center justify-between border-b border-zinc-100 px-6 py-4">
				<h2 class="text-sm font-semibold tracking-tight text-zinc-900">{title}</h2>
				<button
					onclick={close}
					class="flex h-7 w-7 items-center justify-center rounded-lg text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-600"
				>
					<X size={14} />
				</button>
			</div>
			<div class="px-6 py-5">
				{@render children()}
			</div>
		</div>
	</div>
{/if}

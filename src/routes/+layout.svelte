<script lang="ts">
	import '../app.css';
	import Nav from '$lib/components/Nav.svelte';
	import { authStore } from '$lib/stores/auth.svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let { children }: { children: import('svelte').Snippet } = $props();

	let isLoginPage = $derived($page.url.pathname === '/login');
	let showNav = $derived(!isLoginPage && !$page.url.pathname.startsWith('/session/'));

	onMount(async () => {
		await authStore.init();
		if (!authStore.user && !isLoginPage) {
			goto('/login');
		} else if (authStore.user && isLoginPage) {
			goto('/');
		}
	});
</script>

{#if showNav}
	<Nav />
{/if}

{@render children()}

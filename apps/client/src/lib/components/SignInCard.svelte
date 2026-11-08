<script lang="ts">
	import { Card, SignInWithFacile } from '@facile/muse';
	import { SIGN_IN_HREF } from '$lib/api';
	import { t } from '$lib/i18n/index.svelte';

	function remember() {
		try {
			sessionStorage.setItem('sonar-return', location.pathname + location.search);
		} catch {
			return;
		}
	}
</script>

<Card class="flex flex-col items-center gap-4 p-6 text-center">
	<p class="max-w-md text-fc-sm text-fc-fg-muted">{t('shell.signInLead', { count: 20 })}</p>
	<SignInWithFacile href={SIGN_IN_HREF} label={t('shell.signIn')} onclick={remember} />
</Card>

import { registerIcons } from '@facile/muse';

/** Sonar's own glyphs, drawn for Sonar and bundled so nothing is fetched at runtime. */
export const SONAR_ICON = 'sonar:mark';
export const IMAGES_ICON = 'sonar:images';
export const NEWS_ICON = 'sonar:news';
export const SPARK_ICON = 'sonar:spark';

const stroke = 'fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"';

/** Registers the glyphs muse lacks. */
export function registerAppIcons() {
	registerIcons({
		[SONAR_ICON]: {
			width: 24,
			height: 24,
			body: `<g ${stroke}><path d="M12 12 19 5"/><path d="M20.5 12A8.5 8.5 0 1 1 12 3.5" opacity=".5"/><path d="M16 12a4 4 0 1 1-4-4"/></g><circle cx="12" cy="12" r="1.6" fill="currentColor"/>`
		},
		[IMAGES_ICON]: {
			width: 24,
			height: 24,
			body: `<g ${stroke}><rect x="3" y="4" width="18" height="16" rx="3"/><path d="m3 16 5-5 4 4 3-3 6 6"/><circle cx="16" cy="9" r="1.5"/></g>`
		},
		[NEWS_ICON]: {
			width: 24,
			height: 24,
			body: `<g ${stroke}><rect x="4" y="3" width="16" height="18" rx="3"/><path d="M8 8h8M8 12h8M8 16h5"/></g>`
		},
		[SPARK_ICON]: {
			width: 24,
			height: 24,
			body: `<g ${stroke}><path d="M12 3c.6 4.2 2.8 6.4 7 7-4.2.6-6.4 2.8-7 7-.6-4.2-2.8-6.4-7-7 4.2-.6 6.4-2.8 7-7Z"/><path d="M19 16v4M17 18h4"/></g>`
		}
	});
}

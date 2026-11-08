import type { Catalogs, Message } from '@facile/muse';
import { search } from './search';
import { settings } from './settings';
import { shell } from './shell';
import { terms } from './terms';
import type { Row } from '../row';

const rows = { ...shell, ...search, ...settings, ...terms };

export type AppKey = keyof typeof rows;
type AppCatalog = Record<AppKey, Message>;

function column(index: number): AppCatalog {
	const entries = Object.entries(rows as Record<string, Row>).map(([key, row]) => [key, row[index]]);
	return Object.fromEntries(entries) as AppCatalog;
}

/** The app's catalogs: English is the source, and every row carries all four locales. */
export const catalogs: Catalogs<AppCatalog> = {
	en: column(0),
	fr: column(1),
	es: column(2),
	de: column(3)
};

import type { Row } from '../row';

/** Strings around the terms of use, whose body lives in $lib/legal. */
export const terms = {
	'terms.draft': [
		'Draft for review. These terms have not been checked by a lawyer and are not in force yet.',
		'Projet en relecture. Ces conditions n’ont pas été vérifiées par un juriste et ne sont pas encore en vigueur.',
		'Borrador en revisión. Estas condiciones no han sido revisadas por un abogado y aún no están en vigor.',
		'Entwurf zur Prüfung. Diese Bedingungen wurden nicht juristisch geprüft und gelten noch nicht.'
	],
	'terms.languageNote': [
		'',
		'',
		'Este texto solo existe en francés y en inglés. La versión francesa prevalece.',
		'Dieser Text liegt nur auf Französisch und Englisch vor. Die französische Fassung ist maßgeblich.'
	]
} as const satisfies Record<string, Row>;

import type { Row } from '../row';

/** Strings of the header, the footer and the sign-in screens. */
export const shell = {
	'shell.tagline': [
		'Private web search. No history, no tracking.',
		'Recherche web privée. Pas d’historique, pas de pistage.',
		'Búsqueda web privada. Sin historial ni rastreo.',
		'Private Websuche. Kein Verlauf, kein Tracking.'
	],
	'shell.home': ['Home', 'Accueil', 'Inicio', 'Startseite'],
	'shell.signIn': ['Sign in with Facile', 'Se connecter avec Facile', 'Iniciar sesión con Facile', 'Mit Facile anmelden'],
	'shell.signInLead': [
		'Sonar is part of Facile. Sign in to search: the free plan includes {count} searches a day.',
		'Sonar fait partie de Facile. Connectez-vous pour chercher : l’offre gratuite inclut {count} recherches par jour.',
		'Sonar forma parte de Facile. Inicia sesión para buscar: el plan gratuito incluye {count} búsquedas al día.',
		'Sonar gehört zu Facile. Melde dich an, um zu suchen: Der kostenlose Plan umfasst {count} Suchen pro Tag.'
	],
	'shell.terms': ['Terms of use', 'Conditions d’utilisation', 'Condiciones de uso', 'Nutzungsbedingungen'],
	'shell.bangs': ['!bangs', '!bangs', '!bangs', '!bangs'],
	'shell.api': ['API', 'API', 'API', 'API'],
	'shell.settings': ['Settings', 'Paramètres', 'Ajustes', 'Einstellungen'],
	'shell.install': ['Add to browser', 'Ajouter au navigateur', 'Añadir al navegador', 'Zum Browser hinzufügen'],
	'shell.facile': ['Facile', 'Facile', 'Facile', 'Facile'],
	'shell.privacyLine': [
		'Sonar never stores what you search. It keeps one number per day: how many searches you made.',
		'Sonar n’enregistre jamais vos recherches. Il garde un seul nombre par jour : combien de recherches vous avez faites.',
		'Sonar nunca guarda lo que buscas. Solo guarda un número al día: cuántas búsquedas hiciste.',
		'Sonar speichert nie, wonach du suchst. Es behält eine Zahl pro Tag: wie oft du gesucht hast.'
	],
	'shell.poweredBy': [
		'Results from Brave Search and other engines.',
		'Résultats de Brave Search et d’autres moteurs.',
		'Resultados de Brave Search y otros motores.',
		'Ergebnisse von Brave Search und anderen Suchmaschinen.'
	],
	'shell.loading': ['Loading…', 'Chargement…', 'Cargando…', 'Wird geladen…']
} as const satisfies Record<string, Row>;

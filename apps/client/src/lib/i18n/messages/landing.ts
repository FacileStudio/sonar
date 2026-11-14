import type { Row } from '../row';

/** Strings of the signed-out landing page and its demos. */
export const landing = {
	'landing.title': ['Search the web without a trace', 'Cherchez sur le web sans laisser de trace', 'Busca en la web sin dejar rastro', 'Im Web suchen, ohne Spuren'],
	'landing.lead': [
		'Several engines, one merged list. No history, no tracking.',
		'Plusieurs moteurs, une seule liste. Pas d’historique, pas de pistage.',
		'Varios motores, una sola lista. Sin historial ni rastreo.',
		'Mehrere Suchmaschinen, eine Liste. Kein Verlauf, kein Tracking.'
	],
	'landing.installCli': ['Install the CLI', 'Installer la CLI', 'Instalar la CLI', 'CLI installieren'],
	'landing.navFeatures': ['Features', 'Fonctionnalités', 'Funciones', 'Funktionen'],
	'landing.navCli': ['CLI', 'CLI', 'CLI', 'CLI'],
	'landing.featuresTitle': ['Why Sonar', 'Pourquoi Sonar', 'Por qué Sonar', 'Warum Sonar'],
	'landing.enginesTitle': ['Several engines at once', 'Plusieurs moteurs à la fois', 'Varios motores a la vez', 'Mehrere Suchmaschinen zugleich'],
	'landing.enginesBody': [
		'Sonar asks Brave and other engines, then merges duplicates into one ranked list.',
		'Sonar interroge Brave et d’autres moteurs, puis fusionne les doublons en une seule liste.',
		'Sonar consulta Brave y otros motores y une los duplicados en una sola lista.',
		'Sonar fragt Brave und weitere Suchmaschinen und fasst Duplikate zu einer Liste zusammen.'
	],
	'landing.privacyTitle': ['Nothing stored', 'Rien n’est enregistré', 'Nada se guarda', 'Nichts wird gespeichert'],
	'landing.bangsTitle': ['!bangs', '!bangs', '!bangs', '!bangs'],
	'landing.bangsBody': [
		'Type !w or !gh before a query to jump straight to Wikipedia or GitHub.',
		'Tapez !w ou !gh avant une recherche pour aller droit sur Wikipédia ou GitHub.',
		'Escribe !w o !gh antes de la búsqueda para ir directo a Wikipedia o GitHub.',
		'Tippe !w oder !gh vor die Suche, um direkt zu Wikipedia oder GitHub zu springen.'
	],
	'landing.keysTitle': ['Your own keys', 'Vos propres clés', 'Tus propias claves', 'Eigene Schlüssel'],
	'landing.keysBody': [
		'Add your own API keys. Searches on them skip the daily quota.',
		'Ajoutez vos propres clés d’API. Les recherches faites avec ne comptent pas dans le quota.',
		'Añade tus propias claves de API. Esas búsquedas no cuentan para la cuota diaria.',
		'Füge eigene API-Schlüssel hinzu. Suchen damit zählen nicht zum Tageskontingent.'
	],
	'landing.cliTitle': ['In your terminal', 'Dans votre terminal', 'En tu terminal', 'Im Terminal'],
	'landing.cliLead': [
		'Search from the shell and fetch any page as markdown.',
		'Cherchez depuis le shell et récupérez n’importe quelle page en markdown.',
		'Busca desde la terminal y obtén cualquier página en markdown.',
		'Suche in der Shell und hol jede Seite als Markdown.'
	],
	'landing.closingTitle': ['Start searching', 'Commencez à chercher', 'Empieza a buscar', 'Jetzt suchen'],
	'landing.engines': ['{count} engines', '{count} moteurs', '{count} motores', '{count} Suchmaschinen'],
	'landing.searchDemo': [
		'A search for circuit breaker pattern returns one list merged from three engines.',
		'Une recherche « circuit breaker pattern » renvoie une liste fusionnée depuis trois moteurs.',
		'Una búsqueda de circuit breaker pattern devuelve una lista combinada de tres motores.',
		'Eine Suche nach circuit breaker pattern liefert eine Liste aus drei Suchmaschinen.'
	],
	'landing.fetchDemo': [
		'sonar fetch turns a web page into markdown in the terminal.',
		'sonar fetch transforme une page web en markdown dans le terminal.',
		'sonar fetch convierte una página web en markdown en la terminal.',
		'sonar fetch macht im Terminal aus einer Webseite Markdown.'
	]
} as const satisfies Record<string, Row>;

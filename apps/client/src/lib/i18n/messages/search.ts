import type { Row } from '../row';

/** Strings of the search box and the result pages. */
export const search = {
	'search.placeholder': ['Search the web', 'Rechercher sur le web', 'Buscar en la web', 'Im Web suchen'],
	'search.label': ['Search', 'Rechercher', 'Buscar', 'Suchen'],
	'search.submit': ['Search', 'Rechercher', 'Buscar', 'Suchen'],
	'search.suggestions': ['Suggestions', 'Suggestions', 'Sugerencias', 'Vorschläge'],
	'search.tabWeb': ['Web', 'Web', 'Web', 'Web'],
	'search.tabImages': ['Images', 'Images', 'Imágenes', 'Bilder'],
	'search.tabNews': ['News', 'Actualités', 'Noticias', 'Nachrichten'],
	'search.tabs': ['Result type', 'Type de résultat', 'Tipo de resultado', 'Ergebnistyp'],
	'search.noResults': ['No results for “{q}”', 'Aucun résultat pour « {q} »', 'Sin resultados para «{q}»', 'Keine Ergebnisse für „{q}“'],
	'search.noResultsHint': [
		'Try other words, or another tab.',
		'Essayez d’autres mots, ou un autre onglet.',
		'Prueba otras palabras u otra pestaña.',
		'Versuche andere Wörter oder einen anderen Tab.'
	],
	'search.failed': ['The search failed', 'La recherche a échoué', 'La búsqueda falló', 'Die Suche ist fehlgeschlagen'],
	'search.retry': ['Try again', 'Réessayer', 'Reintentar', 'Erneut versuchen'],
	'search.ownKeys': ['Your keys', 'Vos clés', 'Tus claves', 'Deine Schlüssel'],
	'search.usage': ['{used} of {limit} searches today', '{used} recherches sur {limit} aujourd’hui', '{used} de {limit} búsquedas hoy', '{used} von {limit} Suchen heute'],
	'search.unlimited': ['Unlimited searches', 'Recherches illimitées', 'Búsquedas ilimitadas', 'Unbegrenzte Suchen'],
	'search.keyboardHint': [
		'Press / to search, ↑ ↓ to move, Enter to open.',
		'Appuyez sur / pour chercher, ↑ ↓ pour vous déplacer, Entrée pour ouvrir.',
		'Pulsa / para buscar, ↑ ↓ para moverte, Intro para abrir.',
		'Drücke / zum Suchen, ↑ ↓ zum Bewegen, Enter zum Öffnen.'
	],
	'search.results': ['Results', 'Résultats', 'Resultados', 'Ergebnisse'],
	'search.summary': ['AI summary', 'Résumé IA', 'Resumen IA', 'KI-Zusammenfassung'],
	'search.summarize': ['Summarize with AI', 'Résumer avec l’IA', 'Resumir con IA', 'Mit KI zusammenfassen'],
	'search.summaryNote': [
		'Written by Bulle from the sources listed above. Check them.',
		'Écrit par Bulle à partir des sources listées ci-dessus. Vérifiez-les.',
		'Escrito por Bulle a partir de las fuentes de arriba. Compruébalas.',
		'Von Bulle aus den oben genannten Quellen geschrieben. Prüfe sie.'
	],
	'search.summaryFailed': ['The summary failed', 'Le résumé a échoué', 'El resumen falló', 'Die Zusammenfassung ist fehlgeschlagen'],
	'search.redirecting': ['Opening {name}…', 'Ouverture de {name}…', 'Abriendo {name}…', '{name} wird geöffnet…'],
	'search.emptyTitle': ['Search without a trace', 'Cherchez sans laisser de trace', 'Busca sin dejar rastro', 'Suchen ohne Spuren'],
	'search.source': ['Source', 'Source', 'Fuente', 'Quelle']
} as const satisfies Record<string, Row>;

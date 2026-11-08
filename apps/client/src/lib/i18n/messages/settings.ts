import type { Row } from '../row';

/** Strings of the settings page and the bang list. */
export const settings = {
	'settings.title': ['Settings', 'Paramètres', 'Ajustes', 'Einstellungen'],
	'settings.profile': ['Profile', 'Profil', 'Perfil', 'Profil'],
	'settings.preferences': ['Preferences', 'Préférences', 'Preferencias', 'Einstellungen'],
	'settings.keys': ['Engine keys', 'Clés de moteurs', 'Claves de motores', 'Suchmaschinen-Schlüssel'],
	'settings.subscription': ['Subscription', 'Abonnement', 'Suscripción', 'Abo'],
	'settings.suggestions': ['Suggestions while typing', 'Suggestions pendant la saisie', 'Sugerencias al escribir', 'Vorschläge beim Tippen'],
	'settings.suggestionsHelp': [
		'What you type is sent to DuckDuckGo’s suggestion service, without your account or address.',
		'Ce que vous tapez est envoyé au service de suggestions de DuckDuckGo, sans votre compte ni votre adresse.',
		'Lo que escribes se envía al servicio de sugerencias de DuckDuckGo, sin tu cuenta ni tu dirección.',
		'Was du tippst, geht an den Vorschlagsdienst von DuckDuckGo, ohne dein Konto oder deine Adresse.'
	],
	'settings.keysHelp': [
		'Bring your own API keys. Searches that run on them do not count against your daily quota. Keys are encrypted on the server and never shown again.',
		'Utilisez vos propres clés d’API. Les recherches faites avec elles ne comptent pas dans votre quota quotidien. Les clés sont chiffrées sur le serveur et ne sont plus jamais affichées.',
		'Usa tus propias claves de API. Las búsquedas con ellas no cuentan en tu cuota diaria. Las claves se cifran en el servidor y no se vuelven a mostrar.',
		'Nutze eigene API-Schlüssel. Suchen damit zählen nicht zu deinem Tageskontingent. Schlüssel werden auf dem Server verschlüsselt und nie wieder angezeigt.'
	],
	'settings.keysDisabled': [
		'This server cannot store keys yet.',
		'Ce serveur ne peut pas encore enregistrer de clés.',
		'Este servidor aún no puede guardar claves.',
		'Dieser Server kann noch keine Schlüssel speichern.'
	],
	'settings.keySaved': ['Saved, ends in {hint}', 'Enregistrée, se termine par {hint}', 'Guardada, termina en {hint}', 'Gespeichert, endet auf {hint}'],
	'settings.keyNone': ['No key', 'Aucune clé', 'Sin clave', 'Kein Schlüssel'],
	'settings.keySave': ['Save', 'Enregistrer', 'Guardar', 'Speichern'],
	'settings.keyRemove': ['Remove', 'Retirer', 'Quitar', 'Entfernen'],
	'settings.keyFailed': ['The key was not saved', 'La clé n’a pas été enregistrée', 'La clave no se guardó', 'Der Schlüssel wurde nicht gespeichert'],
	'settings.plan': ['Plan', 'Offre', 'Plan', 'Plan'],
	'settings.planFree': ['Free', 'Gratuit', 'Gratis', 'Kostenlos'],
	'settings.planPaid': ['Paid', 'Payant', 'De pago', 'Bezahlt'],
	'settings.searchesToday': ['Searches today', 'Recherches aujourd’hui', 'Búsquedas hoy', 'Suchen heute'],
	'settings.manage': ['Manage on Facile', 'Gérer sur Facile', 'Gestionar en Facile', 'Auf Facile verwalten'],
	'bangs.title': ['!bang shortcuts', 'Raccourcis !bang', 'Atajos !bang', '!bang-Kürzel'],
	'bangs.lead': [
		'Type a bang anywhere in a query. Internal bangs stay in Sonar.',
		'Tapez un bang n’importe où dans une recherche. Les bangs internes restent dans Sonar.',
		'Escribe un bang en cualquier parte de la búsqueda. Los bangs internos se quedan en Sonar.',
		'Schreibe ein Bang irgendwo in die Suche. Interne Bangs bleiben in Sonar.'
	],
	'bangs.internal': ['In Sonar', 'Dans Sonar', 'En Sonar', 'In Sonar'],
	'bangs.external': ['Other sites', 'Autres sites', 'Otros sitios', 'Andere Seiten'],
	'bangs.privacy': [
		'An external bang sends your query straight to that site. That site sees it and its own privacy policy applies.',
		'Un bang externe envoie votre recherche directement à ce site. Ce site la voit et sa propre politique de confidentialité s’applique.',
		'Un bang externo envía tu búsqueda directamente a ese sitio. Ese sitio la ve y se aplica su propia política de privacidad.',
		'Ein externes Bang schickt deine Suche direkt an diese Seite. Sie sieht sie, und ihre eigene Datenschutzerklärung gilt.'
	]
} as const satisfies Record<string, Row>;

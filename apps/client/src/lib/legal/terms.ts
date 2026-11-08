/** One section of the terms: a heading and its paragraphs. */
export type TermsSection = { heading: string; body: string[] };

/** The terms of use in one language. */
export type Terms = { title: string; updated: string; sections: TermsSection[] };

const fr: Terms = {
	title: 'Conditions générales d’utilisation de Sonar',
	updated: 'Projet du 8 novembre 2026',
	sections: [
		{
			heading: '1. Objet',
			body: [
				'Sonar (sonar.facile.studio) est un moteur de recherche web édité par Facile Studio. Il interroge des moteurs tiers, dont Brave Search, et affiche leurs résultats. Les présentes conditions s’ajoutent aux conditions générales du compte Facile.'
			]
		},
		{
			heading: '2. Compte et offres',
			body: [
				'L’utilisation de Sonar demande un compte Facile. L’offre gratuite donne un nombre limité de recherches par jour ; les offres payantes en donnent davantage. Le compteur repart à zéro chaque jour à minuit UTC.',
				'Vous pouvez enregistrer vos propres clés d’API de moteurs. Les recherches faites avec vos clés ne comptent pas dans votre quota ; vous restez seul responsable du respect des conditions du fournisseur de chaque clé.'
			]
		},
		{
			heading: '3. Données personnelles',
			body: [
				'Sonar n’enregistre pas vos recherches. Il ne garde qu’un compteur par jour et par personne, supprimé après deux jours. Les journaux techniques ne contiennent pas le texte des recherches.',
				'Vos requêtes sont transmises aux moteurs interrogés, sans votre identité. Les suggestions de saisie sont envoyées à DuckDuckGo ; vous pouvez les désactiver dans les paramètres. Les images des résultats passent par notre serveur : le site d’origine ne voit pas votre adresse.',
				'Un !bang externe (par exemple !g) envoie votre recherche directement au site choisi, qui la voit et applique sa propre politique de confidentialité.'
			]
		},
		{
			heading: '4. Résultats de recherche et engagements de l’utilisateur',
			body: [
				'Les résultats appartiennent à leurs fournisseurs (notamment Brave Software, Inc.) et aux tiers qui les publient. Ils vous sont fournis pour votre usage personnel de Sonar. En utilisant Sonar, vous vous engagez à ne pas, et à ne laisser personne :',
				'a) stocker, mettre en cache ou constituer une base de données des résultats, en tout ou en partie, au-delà de ce qu’exige leur affichage ;',
				'b) créer des œuvres dérivées des résultats ;',
				'c) redistribuer, revendre, publier, céder ou concéder en sous-licence les résultats, ou l’accès au service ;',
				'd) utiliser les résultats pour créer, évaluer, entraîner, ré-entraîner, affiner, comparer ou améliorer des modèles ou services d’intelligence artificielle ;',
				'e) contourner les limites de débit ou de service, notamment en créant plusieurs comptes, ou interroger Sonar de façon automatisée et massive hors de l’API prévue ;',
				'f) dissimuler votre identité ou celle d’une application pour obtenir un accès ;',
				'g) utiliser Sonar ou ses résultats en violation de la loi ou des droits de tiers, ou pour des contenus ou activités illicites, malveillants, diffamatoires ou haineux.',
				'Ces engagements reprennent ceux que Brave Software impose à ses clients pour leurs utilisateurs. Ils restent applicables à toute donnée issue de Sonar après la fin de votre utilisation : vous devez alors cesser tout usage et supprimer les résultats que vous auriez conservés.'
			]
		},
		{
			heading: '5. Résumé par IA',
			body: [
				'Les offres payantes peuvent demander un résumé des résultats, rédigé par Bulle. Ce résumé peut être inexact ou incomplet. Il ne constitue pas un conseil professionnel, juridique, médical ou financier ; vérifiez toujours les sources.'
			]
		},
		{
			heading: '6. Suspension',
			body: [
				'Facile Studio surveille l’usage du service pour faire respecter ces conditions et celles de ses fournisseurs. En cas de manquement, il peut restreindre ou suspendre votre accès sans préavis.'
			]
		},
		{
			heading: '7. Responsabilité',
			body: [
				'Sonar et ses résultats sont fournis « en l’état ». Facile Studio ne garantit ni l’exactitude, ni l’exhaustivité, ni la disponibilité des résultats, qui proviennent de tiers.'
			]
		},
		{
			heading: '8. Droit applicable et contact',
			body: [
				'Ces conditions sont régies par le droit français. Pour toute question : contact@facile.studio.'
			]
		}
	]
};

const en: Terms = {
	title: 'Sonar terms of use',
	updated: 'Draft of 8 November 2026',
	sections: [
		{
			heading: '1. Purpose',
			body: [
				'Sonar (sonar.facile.studio) is a web search engine run by Facile Studio. It queries third-party engines, including Brave Search, and shows their results. These terms add to the general terms of your Facile account.'
			]
		},
		{
			heading: '2. Account and plans',
			body: [
				'Sonar needs a Facile account. The free plan gives a limited number of searches per day; paid plans give more. The counter resets every day at midnight UTC.',
				'You can save your own engine API keys. Searches on your keys do not count against your quota, and you alone answer for following each key provider’s terms.'
			]
		},
		{
			heading: '3. Personal data',
			body: [
				'Sonar does not record your searches. It keeps one counter per person and day, deleted after two days. Technical logs do not contain the text of searches.',
				'Your queries go to the engines Sonar asks, without your identity. Typing suggestions go to DuckDuckGo; you can turn them off in settings. Result images pass through our server, so the original site does not see your address.',
				'An external !bang (for example !g) sends your search straight to the chosen site, which sees it and applies its own privacy policy.'
			]
		},
		{
			heading: '4. Search results and your commitments',
			body: [
				'Results belong to their providers (including Brave Software, Inc.) and to the third parties who publish them. They are provided for your personal use of Sonar. By using Sonar you agree not to, and not to let anyone:',
				'a) store, cache or build a database of results, in whole or in part, beyond what showing them requires;',
				'b) create derivative works of the results;',
				'c) redistribute, resell, publish, assign or sublicense the results, or access to the service;',
				'd) use the results to create, evaluate, train, re-train, fine-tune, benchmark or improve artificial intelligence models or services;',
				'e) get around rate or service limits, including by creating several accounts, or query Sonar in bulk by automated means outside the provided API;',
				'f) hide your identity or an application’s identity to obtain access;',
				'g) use Sonar or its results against the law or the rights of others, or for unlawful, malicious, defamatory or hateful content or activity.',
				'These commitments mirror those Brave Software requires its customers to impose on their users. They keep applying to any data from Sonar after you stop using it: you must then stop all use and delete any results you kept.'
			]
		},
		{
			heading: '5. AI summary',
			body: [
				'Paid plans can ask for a summary of the results, written by Bulle. A summary can be wrong or incomplete. It is not professional, legal, medical or financial advice; always check the sources.'
			]
		},
		{
			heading: '6. Suspension',
			body: [
				'Facile Studio monitors use of the service to enforce these terms and its providers’ terms. On a breach it may restrict or suspend your access without notice.'
			]
		},
		{
			heading: '7. Liability',
			body: [
				'Sonar and its results are provided “as is”. Facile Studio does not guarantee that results, which come from third parties, are accurate, complete or available.'
			]
		},
		{
			heading: '8. Governing law and contact',
			body: ['These terms are governed by French law. Questions: contact@facile.studio.']
		}
	]
};

/** The terms for a locale: French or English, with English for es and de. */
export function termsFor(locale: string): Terms {
	return locale === 'fr' ? fr : en;
}

package bangs

func internal() []Bang {
	return []Bang{
		{Trigger: "i", Name: "Images", Kind: "images"},
		{Trigger: "n", Name: "News", Kind: "news"},
		{Trigger: "web", Name: "Web", Kind: "web"},
		{Trigger: "brave", Name: "Brave", Engine: "brave"},
		{Trigger: "tavily", Name: "Tavily", Engine: "tavily"},
		{Trigger: "exa", Name: "Exa", Engine: "exa"},
		{Trigger: "parallel", Name: "Parallel", Engine: "parallel"},
	}
}

func external() []Bang {
	return []Bang{
		{Trigger: "g", Name: "Google", URL: "https://www.google.com/search?q=%s"},
		{Trigger: "gi", Name: "Google Images", URL: "https://www.google.com/search?tbm=isch&q=%s"},
		{Trigger: "gm", Name: "Google Maps", URL: "https://www.google.com/maps/search/%s"},
		{Trigger: "ddg", Name: "DuckDuckGo", URL: "https://duckduckgo.com/?q=%s"},
		{Trigger: "b", Name: "Bing", URL: "https://www.bing.com/search?q=%s"},
		{Trigger: "qw", Name: "Qwant", URL: "https://www.qwant.com/?q=%s"},
		{Trigger: "w", Name: "Wikipedia", URL: "https://en.wikipedia.org/w/index.php?search=%s"},
		{Trigger: "wfr", Name: "Wikipédia (fr)", URL: "https://fr.wikipedia.org/w/index.php?search=%s"},
		{Trigger: "yt", Name: "YouTube", URL: "https://www.youtube.com/results?search_query=%s"},
		{Trigger: "gh", Name: "GitHub", URL: "https://github.com/search?q=%s"},
		{Trigger: "so", Name: "Stack Overflow", URL: "https://stackoverflow.com/search?q=%s"},
		{Trigger: "mdn", Name: "MDN", URL: "https://developer.mozilla.org/search?q=%s"},
		{Trigger: "npm", Name: "npm", URL: "https://www.npmjs.com/search?q=%s"},
		{Trigger: "go", Name: "Go packages", URL: "https://pkg.go.dev/search?q=%s"},
		{Trigger: "r", Name: "Reddit", URL: "https://www.reddit.com/search/?q=%s"},
		{Trigger: "hn", Name: "Hacker News", URL: "https://hn.algolia.com/?q=%s"},
		{Trigger: "osm", Name: "OpenStreetMap", URL: "https://www.openstreetmap.org/search?query=%s"},
		{Trigger: "wr", Name: "WordReference", URL: "https://www.wordreference.com/fren/%s"},
		{Trigger: "dl", Name: "DeepL", URL: "https://www.deepl.com/translator#auto/en/%s"},
	}
}

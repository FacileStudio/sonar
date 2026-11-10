package mcp

// searchInput is the input of sonar_search.
type searchInput struct {
	Query   string   `json:"query" jsonschema:"what to search for, at most 400 characters. Quotes and site:example.com narrow a web search"`
	Kind    string   `json:"kind,omitempty" jsonschema:"web (default), images or news"`
	Count   int      `json:"count,omitempty" jsonschema:"results to return, 1 to 20. Default 10"`
	Engines []string `json:"engines,omitempty" jsonschema:"web only, and only when the user saved their own engine keys in Sonar: run just these engines, such as exa or tavily. Omit otherwise"`
}

// webHit is one web result.
type webHit struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// imageHit is one image result. URL is the page that shows the image.
type imageHit struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	ImageURL string `json:"image_url,omitempty"`
	Source   string `json:"source,omitempty"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
}

// newsHit is one news result.
type newsHit struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Snippet   string `json:"snippet"`
	Source    string `json:"source,omitempty"`
	Age       string `json:"age,omitempty"`
	Published string `json:"published,omitempty"`
}

// searchOutput is the output of sonar_search: results for web, images or
// news for those kinds.
type searchOutput struct {
	Query   string     `json:"query"`
	Kind    string     `json:"kind"`
	Results []webHit   `json:"results,omitempty"`
	Images  []imageHit `json:"images,omitempty"`
	News    []newsHit  `json:"news,omitempty"`
	OwnKeys bool       `json:"own_keys" jsonschema:"true when the search ran on the user's own keys and spent no quota"`
	Note    string     `json:"note,omitempty"`
}

// fetchInput is the input of sonar_fetch.
type fetchInput struct {
	URL       string `json:"url" jsonschema:"the public http or https page to read, such as a url from sonar_search"`
	Offset    int    `json:"offset,omitempty" jsonschema:"where to start in the markdown, in bytes: next_offset from the previous call. Default 0"`
	MaxLength int    `json:"max_length,omitempty" jsonschema:"bytes of markdown to return, 1000 to 60000. Default 30000"`
}

// fetchOutput is the output of sonar_fetch.
type fetchOutput struct {
	URL             string `json:"url"`
	FinalURL        string `json:"final_url"`
	Title           string `json:"title"`
	Markdown        string `json:"markdown"`
	Offset          int    `json:"offset"`
	NextOffset      int    `json:"next_offset,omitempty" jsonschema:"set when more markdown follows: pass it as offset to read on"`
	TotalLength     int    `json:"total_length" jsonschema:"bytes of markdown in the whole page"`
	SourceTruncated bool   `json:"source_truncated,omitempty" jsonschema:"the page was longer than sonar reads, so its end is missing"`
	Note            string `json:"note,omitempty"`
}

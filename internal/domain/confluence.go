package domain

// DefaultBodyFormat is markdown for Confluence create/update/get.
const DefaultBodyFormat = "markdown"

// Page is one Confluence page on a single cloud.
// Site is the hostname. URL is the wiki page link.
type Page struct {
	ID            string `json:"id"`
	Site          string `json:"site"`
	Space         string `json:"space"`
	Title         string `json:"title"`
	Body          string `json:"body,omitempty"`
	ContentFormat string `json:"content_format,omitempty"`
	Status        string `json:"status,omitempty"`
	URL           string `json:"url,omitempty"`
	Version       int    `json:"version,omitempty"`
}

// PageSearchResult is one-site CQL output.
type PageSearchResult struct {
	CQL   string `json:"cql"`
	Site  string `json:"site"`
	Count int    `json:"count"`
	Items []Page `json:"items"`
}

// CreatePage is the REST field set we own for atlas confluence create.
// Body format is markdown. Space accepts a space key (CCAB) or numeric id.
type CreatePage struct {
	Space string
	Title string
	Body  string
}

// WikiPageURL is https://<hostname>/wiki/spaces/<SPACE>/pages/<id>.
func WikiPageURL(hostname, space, id string) string {
	return "https://" + hostname + "/wiki/spaces/" + space + "/pages/" + id
}

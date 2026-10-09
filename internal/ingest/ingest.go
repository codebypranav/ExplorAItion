package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	// Open-Meteo's geocoder is free and needs no key. The project already
	// depends on Open-Meteo for weather.
	geocodeURL = "https://geocoding-api.open-meteo.com/v1/search"
	// Wikipedia GeoSearch is the fallback POI source when no Overpass
	// instance can be reached.
	wikiGeoRadiusMax = 10000
	// Wikipedia supplies the descriptions that OSM tags usually lack.
	wikipediaAPIFmt = "https://%s.wikipedia.org/w/api.php"
	// Titles per Wikipedia extracts request (the API caps this at 50).
	wikiBatchSize = 20
	// Longest description kept, in characters, before truncating at a sentence.
	maxDescription = 600
	// Wikimedia's user agent policy asks for an identifying contact URL.
	userAgent = "ExplorAItion/1.0 (https://github.com/codebypranav/ExplorAItion)"
	// Pause between successive Wikipedia batches, to stay a polite client.
	wikiPause = 300 * time.Millisecond
)

// Public Overpass instances, tried in order. Set OVERPASS_URL to override the
// list with a single endpoint (for a self-hosted instance, say).
var overpassMirrors = []string{
	"https://overpass-api.de/api/interpreter",
	"https://overpass.kumi.systems/api/interpreter",
	"https://overpass.private.coffee/api/interpreter",
}

func overpassEndpoints() []string {
	if v := strings.TrimSpace(os.Getenv("OVERPASS_URL")); v != "" {
		return []string{v}
	}
	return overpassMirrors
}

// POI contains fields which we will map to Pinecone metadata, and embed.
// POI describes a Point-of-Interest sourced from OpenStreetMap.
type POI struct {
	XID         string   `json:"xid"`
	Name        string   `json:"name"`
	Kinds       string   `json:"kinds"`
	Lat         float64  `json:"lat"`
	Lon         float64  `json:"lon"`
	Country     string   `json:"country"`
	Description string   `json:"description"`
	Rate        float64  `json:"rate"`
	Image       string   `json:"image"`
	Tags        []string `json:"tags"`
}

// City is a resolved place name with the country it sits in.
type City struct {
	Name        string
	Lat         float64
	Lon         float64
	CountryCode string
}

func httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return do(req)
}

func httpPostForm(ctx context.Context, endpoint string, form url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return do(req)
}

func do(req *http.Request) ([]byte, error) {
	// Overpass queries over a wide radius routinely take tens of seconds.
	client := &http.Client{
		Timeout: 90 * time.Second,
		// Unreachable mirrors should be abandoned quickly rather than burning
		// the full request timeout on a dial that will never connect.
		Transport: &http.Transport{
			DialContext:         (&net.Dialer{Timeout: 8 * time.Second}).DialContext,
			TLSHandshakeTimeout: 8 * time.Second,
		},
	}

	// Both Overpass and Wikipedia shed load with 429/503 under pressure; back
	// off and retry rather than dropping a whole city's worth of POIs. A
	// transport error means the host is unreachable, so return at once and let
	// the caller try the next mirror instead of waiting out more timeouts.
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			wait := time.Duration(1<<attempt) * time.Second
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(wait):
			}
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		b, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		switch {
		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable:
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
			continue
		case resp.StatusCode >= 300:
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
		}
		return b, nil
	}
	return nil, lastErr
}

// LookupCity resolves a place name to coordinates and a country code.
func LookupCity(ctx context.Context, city string) (City, error) {
	u := fmt.Sprintf("%s?name=%s&count=1&language=en&format=json", geocodeURL, url.QueryEscape(city))
	b, err := httpGet(ctx, u)
	if err != nil {
		return City{}, err
	}
	var out struct {
		Results []struct {
			Name        string  `json:"name"`
			Latitude    float64 `json:"latitude"`
			Longitude   float64 `json:"longitude"`
			CountryCode string  `json:"country_code"`
		} `json:"results"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return City{}, err
	}
	if len(out.Results) == 0 {
		return City{}, fmt.Errorf("no match for city %q", city)
	}
	r := out.Results[0]
	return City{
		Name:        r.Name,
		Lat:         r.Latitude,
		Lon:         r.Longitude,
		CountryCode: strings.ToUpper(r.CountryCode),
	}, nil
}

// GetCityCoords finds the latitude/longitude for a city name.
func GetCityCoords(ctx context.Context, city string) (lat, lon float64, err error) {
	c, err := LookupCity(ctx, city)
	if err != nil {
		return 0, 0, err
	}
	return c.Lat, c.Lon, nil
}

// overpassElement is one OSM node, way or relation as Overpass returns it.
type overpassElement struct {
	Type   string  `json:"type"`
	ID     int64   `json:"id"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Center *struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"center"`
	Tags map[string]string `json:"tags"`
}

// SearchPOIs returns travel-relevant, named OSM features within radiusMeters of
// the given point. It asks Overpass for more than requested, then keeps the
// highest-scoring limit of them.
func SearchPOIs(ctx context.Context, lat, lon float64, radiusMeters int, limit int) ([]overpassElement, error) {
	if limit <= 0 {
		limit = 100
	}
	// Ask for a surplus so the ranking below has something to choose from.
	fetch := limit * 4
	if fetch > 2000 {
		fetch = 2000
	}

	area := fmt.Sprintf("(around:%d,%f,%f)", radiusMeters, lat, lon)
	query := fmt.Sprintf(`[out:json][timeout:90];
(
  nwr["tourism"~"^(museum|attraction|gallery|artwork|viewpoint|zoo|aquarium|theme_park)$"]["name"]%[1]s;
  nwr["historic"~"^(castle|monument|memorial|ruins|archaeological_site|church|fort|city_gate|building)$"]["name"]%[1]s;
  nwr["leisure"~"^(park|garden|nature_reserve)$"]["name"]%[1]s;
  nwr["amenity"~"^(theatre|place_of_worship|marketplace)$"]["name"]%[1]s;
);
out center %[2]d;`, area, fetch)

	// Public Overpass instances are frequently overloaded or unreachable, so
	// fall through the mirror list rather than failing on the first one.
	var (
		out struct {
			Elements []overpassElement `json:"elements"`
		}
		lastErr error
		ok      bool
	)
	for _, endpoint := range overpassEndpoints() {
		b, err := httpPostForm(ctx, endpoint, url.Values{"data": {query}})
		if err != nil {
			lastErr = fmt.Errorf("%s: %w", endpoint, err)
			continue
		}
		if err := json.Unmarshal(b, &out); err != nil {
			lastErr = fmt.Errorf("%s: %w", endpoint, err)
			continue
		}
		ok = true
		break
	}
	if !ok {
		return nil, fmt.Errorf("all Overpass endpoints failed, last error: %w", lastErr)
	}

	// Keep the most travel-worthy features first: a Wikipedia article is the
	// strongest signal that a place is worth visiting.
	sort.SliceStable(out.Elements, func(i, j int) bool {
		return scoreTags(out.Elements[i].Tags) > scoreTags(out.Elements[j].Tags)
	})
	if len(out.Elements) > limit {
		out.Elements = out.Elements[:limit]
	}
	return out.Elements, nil
}

// scoreTags approximates OpenTripMap's 1-7 "rate": how notable a place looks
// from its tags alone.
func scoreTags(tags map[string]string) float64 {
	score := 1.0
	if tags["wikipedia"] != "" {
		score += 3
	}
	if tags["wikidata"] != "" {
		score += 1
	}
	switch tags["tourism"] {
	case "museum", "gallery", "attraction", "zoo", "aquarium", "theme_park":
		score += 2
	case "viewpoint", "artwork":
		score += 1
	}
	if tags["historic"] != "" {
		score += 1
	}
	if tags["heritage"] != "" {
		score += 1
	}
	if score > 7 {
		score = 7
	}
	return score
}

// kindsOf builds an OpenTripMap-style comma separated category string.
func kindsOf(tags map[string]string) string {
	kinds := make([]string, 0, 4)
	for _, key := range []string{"tourism", "historic", "leisure", "amenity"} {
		if v := tags[key]; v != "" {
			kinds = append(kinds, v)
		}
	}
	if tags["heritage"] != "" {
		kinds = append(kinds, "heritage")
	}
	return strings.Join(kinds, ",")
}

func nameOf(tags map[string]string) string {
	if v := tags["name:en"]; v != "" {
		return v
	}
	return tags["name"]
}

// toPOI converts an Overpass element, using countryCode from the city lookup
// since OSM features do not carry a country tag.
func toPOI(el overpassElement, countryCode string) (POI, bool) {
	name := nameOf(el.Tags)
	if name == "" {
		return POI{}, false
	}
	lat, lon := el.Lat, el.Lon
	if el.Center != nil {
		// Ways and relations carry their position in `center`.
		lat, lon = el.Center.Lat, el.Center.Lon
	}
	if lat == 0 && lon == 0 {
		return POI{}, false
	}

	p := POI{
		// OSM ids are only unique per element type, so keep the type prefix.
		XID:     fmt.Sprintf("osm-%s%d", el.Type[:1], el.ID),
		Name:    name,
		Kinds:   kindsOf(el.Tags),
		Lat:     lat,
		Lon:     lon,
		Country: countryCode,
		Rate:    scoreTags(el.Tags),
	}
	if d := el.Tags["description:en"]; d != "" {
		p.Description = d
	} else if d := el.Tags["description"]; d != "" {
		p.Description = d
	}
	if img := el.Tags["image"]; strings.HasPrefix(img, "http") {
		p.Image = img
	}
	if p.Kinds != "" {
		p.Tags = strings.Split(p.Kinds, ",")
	}
	return p, true
}

// FetchPOIsForCity returns POIs around a city, described where possible.
func FetchPOIsForCity(ctx context.Context, city string, radiusMeters int, limit int) ([]POI, error) {
	c, err := LookupCity(ctx, city)
	if err != nil {
		return nil, err
	}
	elements, err := SearchPOIs(ctx, c.Lat, c.Lon, radiusMeters, limit)
	if err != nil {
		// Overpass is the richer source, but it is not always reachable.
		// Wikipedia GeoSearch covers the notable places either way.
		log.Printf("Overpass unavailable (%v); falling back to Wikipedia GeoSearch", err)
		return fetchPOIsFromWikipedia(ctx, c, radiusMeters, limit)
	}
	if len(elements) == 0 {
		log.Printf("Overpass returned nothing for %s; falling back to Wikipedia GeoSearch", city)
		return fetchPOIsFromWikipedia(ctx, c, radiusMeters, limit)
	}

	pois := make([]POI, 0, len(elements))
	// wikiTitles maps a POI index to the article that describes it.
	wikiTitles := map[int]string{}
	for _, el := range elements {
		p, ok := toPOI(el, c.CountryCode)
		if !ok {
			continue
		}
		if lang, title, ok := parseWikipediaTag(el.Tags["wikipedia"]); ok && lang == "en" {
			wikiTitles[len(pois)] = title
		}
		pois = append(pois, p)
	}

	// Descriptions drive embedding quality, so fill the gaps from Wikipedia.
	// A failure here is not fatal: the POIs are still usable without extracts.
	if extracts, err := fetchWikipediaExtracts(ctx, "en", values(wikiTitles)); err == nil {
		for i, title := range wikiTitles {
			if text := extracts[normaliseTitle(title)]; text != "" && pois[i].Description == "" {
				pois[i].Description = text
			}
		}
	}
	return pois, nil
}

// parseWikipediaTag splits an OSM wikipedia tag such as "en:Louvre".
func parseWikipediaTag(tag string) (lang, title string, ok bool) {
	if tag == "" {
		return "", "", false
	}
	lang, title, found := strings.Cut(tag, ":")
	if !found || lang == "" || title == "" {
		return "", "", false
	}
	return lang, title, true
}

func values(m map[int]string) []string {
	out := make([]string, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func normaliseTitle(t string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(t), "_", " "))
}

// fetchWikipediaExtracts returns intro paragraphs keyed by normalised title.
func fetchWikipediaExtracts(ctx context.Context, lang string, titles []string) (map[string]string, error) {
	out := map[string]string{}
	if len(titles) == 0 {
		return out, nil
	}
	endpoint := fmt.Sprintf(wikipediaAPIFmt, lang)

	for start := 0; start < len(titles); start += wikiBatchSize {
		end := start + wikiBatchSize
		if end > len(titles) {
			end = len(titles)
		}
		q := url.Values{
			"action":      {"query"},
			"format":      {"json"},
			"prop":        {"extracts"},
			"exintro":     {"1"},
			"explaintext": {"1"},
			"redirects":   {"1"},
			"titles":      {strings.Join(titles[start:end], "|")},
		}
		if start > 0 {
			time.Sleep(wikiPause)
		}
		b, err := httpGet(ctx, endpoint+"?"+q.Encode())
		if err != nil {
			return out, err
		}
		var resp struct {
			Query struct {
				// The API rewrites titles, so follow both mappings back to
				// whatever we asked for.
				Normalized []struct {
					From string `json:"from"`
					To   string `json:"to"`
				} `json:"normalized"`
				Redirects []struct {
					From string `json:"from"`
					To   string `json:"to"`
				} `json:"redirects"`
				Pages map[string]struct {
					Title   string `json:"title"`
					Extract string `json:"extract"`
				} `json:"pages"`
			} `json:"query"`
		}
		if err := json.Unmarshal(b, &resp); err != nil {
			return out, err
		}

		byTitle := map[string]string{}
		for _, p := range resp.Query.Pages {
			if p.Extract != "" {
				byTitle[normaliseTitle(p.Title)] = truncateAtSentence(p.Extract, maxDescription)
			}
		}
		// Walk requested title -> normalized -> redirected -> page.
		alias := map[string]string{}
		for _, n := range resp.Query.Normalized {
			alias[normaliseTitle(n.From)] = normaliseTitle(n.To)
		}
		for _, r := range resp.Query.Redirects {
			alias[normaliseTitle(r.From)] = normaliseTitle(r.To)
		}
		for _, requested := range titles[start:end] {
			key := normaliseTitle(requested)
			resolved := key
			for i := 0; i < 3; i++ {
				if next, ok := alias[resolved]; ok {
					resolved = next
					continue
				}
				break
			}
			if text, ok := byTitle[resolved]; ok {
				out[key] = text
			}
		}
	}
	return out, nil
}

// truncateAtSentence trims text to max characters, preferring a sentence end.
func truncateAtSentence(text string, max int) string {
	text = strings.TrimSpace(strings.ReplaceAll(text, "\n", " "))
	if len(text) <= max {
		return text
	}
	cut := text[:max]
	if i := strings.LastIndex(cut, ". "); i > max/3 {
		return cut[:i+1]
	}
	if i := strings.LastIndex(cut, " "); i > 0 {
		return cut[:i] + "…"
	}
	return cut
}

// --- Wikipedia GeoSearch fallback -------------------------------------------

// categoryKinds maps a keyword found in a Wikipedia category name to the kind
// of place it implies, so Wikipedia-sourced POIs carry categories comparable to
// the OSM-sourced ones.
var categoryKinds = []struct{ keyword, kind string }{
	{"museum", "museum"},
	{"art galler", "gallery"},
	{"galler", "gallery"},
	{"park", "park"},
	{"garden", "garden"},
	{"church", "church"},
	{"cathedral", "church"},
	{"basilica", "church"},
	{"mosque", "place_of_worship"},
	{"synagogue", "place_of_worship"},
	{"temple", "place_of_worship"},
	{"castle", "castle"},
	{"palace", "castle"},
	{"monument", "monument"},
	{"memorial", "memorial"},
	{"theatre", "theatre"},
	{"theater", "theatre"},
	{"opera", "theatre"},
	{"bridge", "bridge"},
	{"tower", "attraction"},
	{"square", "attraction"},
	{"market", "marketplace"},
	{"zoo", "zoo"},
	{"aquarium", "aquarium"},
	{"cemeter", "historic"},
	{"library", "attraction"},
	{"historic", "historic"},
	{"tourist attraction", "attraction"},
}

// nonPlaceCategories mark articles that have coordinates but are not places you
// can visit: events, organisations, people and the like.
var nonPlaceCategories = []string{
	"fires", "disasters", "explosions", "accidents", "attacks", "bombings",
	"massacres", "battles", "sieges", "riots", "protests", "demonstrations",
	"events", "festivals", "deaths", "births", "people", "writers", "painters",
	"politicians", "dioceses", "archdioceses", "organizations", "organisations",
	"companies", "political parties", "schools of", "universities and colleges",
	"populated places established", "defunct", "demolished",
}

// placeCategories confirm an article describes somewhere physical.
var placeCategories = []string{
	"building", "structure", "landmark", "visitor attraction", "tourist attraction",
	"museum", "gallery", "park", "garden", "church", "cathedral", "basilica",
	"mosque", "synagogue", "temple", "castle", "palace", "monument", "memorial",
	"theatre", "theater", "bridge", "square", "tower", "market", "zoo",
	"aquarium", "cemeter", "library", "station", "street", "fountain", "hotel",
}

func looksLikePlace(categories []string, kinds string) bool {
	for _, c := range categories {
		lc := strings.ToLower(c)
		for _, bad := range nonPlaceCategories {
			if strings.Contains(lc, bad) {
				return false
			}
		}
	}
	if kinds != "" {
		return true
	}
	for _, c := range categories {
		lc := strings.ToLower(c)
		for _, good := range placeCategories {
			if strings.Contains(lc, good) {
				return true
			}
		}
	}
	return false
}

// fetchPOIsFromWikipedia finds articles with coordinates near the city and
// turns them into POIs. Every result is by definition notable enough to have an
// article, which suits a travel index.
func fetchPOIsFromWikipedia(ctx context.Context, c City, radiusMeters, limit int) ([]POI, error) {
	// GeoSearch caps the radius it will consider.
	if radiusMeters > wikiGeoRadiusMax {
		radiusMeters = wikiGeoRadiusMax
	}
	if limit <= 0 {
		limit = 100
	}
	// GeoSearch returns nearest-first, so pull a wider pool than requested and
	// rank it by notability rather than keeping only the closest articles.
	candidates := limit * 5
	if candidates > 300 {
		candidates = 300
	}

	q := url.Values{
		"action":      {"query"},
		"format":      {"json"},
		"list":        {"geosearch"},
		"gscoord":     {fmt.Sprintf("%f|%f", c.Lat, c.Lon)},
		"gsradius":    {fmt.Sprintf("%d", radiusMeters)},
		"gslimit":     {fmt.Sprintf("%d", candidates)},
		"gsnamespace": {"0"},
	}
	b, err := httpGet(ctx, fmt.Sprintf(wikipediaAPIFmt, "en")+"?"+q.Encode())
	if err != nil {
		return nil, err
	}
	var resp struct {
		Query struct {
			GeoSearch []struct {
				PageID int64   `json:"pageid"`
				Title  string  `json:"title"`
				Lat    float64 `json:"lat"`
				Lon    float64 `json:"lon"`
			} `json:"geosearch"`
		} `json:"query"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	pois := make([]POI, 0, len(resp.Query.GeoSearch))
	titles := make([]string, 0, len(resp.Query.GeoSearch))
	for _, g := range resp.Query.GeoSearch {
		pois = append(pois, POI{
			XID:     fmt.Sprintf("wiki-%d", g.PageID),
			Name:    g.Title,
			Lat:     g.Lat,
			Lon:     g.Lon,
			Country: c.CountryCode,
			Rate:    4, // refined below once the article is known
		})
		titles = append(titles, g.Title)
	}

	details, err := fetchWikipediaDetails(ctx, "en", titles)
	if err != nil {
		return nil, err
	}
	kept := pois[:0]
	for i := range pois {
		d, ok := details[normaliseTitle(pois[i].Name)]
		if !ok || d.extract == "" {
			// Without a description there is little to embed.
			continue
		}
		kinds := kindsFromCategories(d.categories)
		if !looksLikePlace(d.categories, kinds) {
			continue
		}
		pois[i].Description = d.extract
		pois[i].Kinds = kinds
		if pois[i].Kinds != "" {
			pois[i].Tags = strings.Split(pois[i].Kinds, ",")
		}
		pois[i].Rate = scoreWikipedia(pois[i])
		kept = append(kept, pois[i])
	}

	sort.SliceStable(kept, func(i, j int) bool { return kept[i].Rate > kept[j].Rate })
	if len(kept) > limit {
		kept = kept[:limit]
	}
	return kept, nil
}

func kindsFromCategories(categories []string) string {
	seen := map[string]bool{}
	kinds := []string{}
	for _, c := range categories {
		lc := strings.ToLower(c)
		for _, m := range categoryKinds {
			if strings.Contains(lc, m.keyword) && !seen[m.kind] {
				seen[m.kind] = true
				kinds = append(kinds, m.kind)
			}
		}
	}
	if len(kinds) > 4 {
		kinds = kinds[:4]
	}
	return strings.Join(kinds, ",")
}

// scoreWikipedia mirrors scoreTags' 1-7 scale for Wikipedia-sourced places.
func scoreWikipedia(p POI) float64 {
	score := 4.0 // having an article at all is already a notability signal
	if p.Kinds != "" {
		score += 1
	}
	if len(p.Description) > 300 {
		score += 1
	}
	if len(p.Description) > 500 {
		score += 1
	}
	if score > 7 {
		score = 7
	}
	return score
}

type wikiDetail struct {
	extract    string
	categories []string
}

// fetchWikipediaDetails pulls intro extracts and categories for many titles.
func fetchWikipediaDetails(ctx context.Context, lang string, titles []string) (map[string]wikiDetail, error) {
	out := map[string]wikiDetail{}
	if len(titles) == 0 {
		return out, nil
	}
	endpoint := fmt.Sprintf(wikipediaAPIFmt, lang)

	for start := 0; start < len(titles); start += wikiBatchSize {
		end := start + wikiBatchSize
		if end > len(titles) {
			end = len(titles)
		}
		q := url.Values{
			"action":      {"query"},
			"format":      {"json"},
			"prop":        {"extracts|categories"},
			"exintro":     {"1"},
			"explaintext": {"1"},
			"exlimit":     {"max"},
			"cllimit":     {"max"},
			"clshow":      {"!hidden"},
			"redirects":   {"1"},
			"titles":      {strings.Join(titles[start:end], "|")},
		}
		if start > 0 {
			time.Sleep(wikiPause)
		}
		b, err := httpGet(ctx, endpoint+"?"+q.Encode())
		if err != nil {
			return out, err
		}
		var resp struct {
			Query struct {
				Pages map[string]struct {
					Title      string `json:"title"`
					Extract    string `json:"extract"`
					Categories []struct {
						Title string `json:"title"`
					} `json:"categories"`
				} `json:"pages"`
			} `json:"query"`
		}
		if err := json.Unmarshal(b, &resp); err != nil {
			return out, err
		}
		for _, p := range resp.Query.Pages {
			cats := make([]string, 0, len(p.Categories))
			for _, c := range p.Categories {
				cats = append(cats, strings.TrimPrefix(c.Title, "Category:"))
			}
			out[normaliseTitle(p.Title)] = wikiDetail{
				extract:    truncateAtSentence(p.Extract, maxDescription),
				categories: cats,
			}
		}
	}
	return out, nil
}

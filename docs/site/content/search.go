package content

import (
	"sort"
	"strings"
)

// SearchEntry represents an indexable document chunk.
type SearchEntry struct {
	DocID     string
	DocTitle  string
	Heading   string
	HeadingID string
	Text      string
	Weight    int
}

// SearchHit is a scored match result.
type SearchHit struct {
	Entry SearchEntry
	Score int
}

// SearchIndex is a lightweight inverted index.
type SearchIndex struct {
	Entries []SearchEntry
	tokens  map[string][]int
}

// NewSearchIndex creates an empty search index.
func NewSearchIndex() *SearchIndex {
	return &SearchIndex{tokens: map[string][]int{}}
}

// AddEntries adds multiple entries to the index.
func (s *SearchIndex) AddEntries(entries []SearchEntry) {
	for _, entry := range entries {
		s.Add(entry)
	}
}

// Add adds a single entry to the index.
func (s *SearchIndex) Add(entry SearchEntry) {
	if s == nil {
		return
	}
	id := len(s.Entries)
	s.Entries = append(s.Entries, entry)
	tokens := uniqueTokens(tokenize(entry.DocTitle + " " + entry.Heading + " " + entry.Text))
	for _, token := range tokens {
		s.tokens[token] = append(s.tokens[token], id)
	}
}

// Search finds matching entries for a query.
func (s *SearchIndex) Search(query string, limit int) []SearchHit {
	if s == nil {
		return nil
	}
	tokens := tokenize(query)
	if len(tokens) == 0 {
		return nil
	}
	scores := map[int]int{}
	for _, token := range tokens {
		for _, id := range s.tokens[token] {
			weight := s.Entries[id].Weight
			if weight <= 0 {
				weight = 1
			}
			scores[id] += weight
		}
	}
	if len(scores) == 0 {
		return nil
	}
	hits := make([]SearchHit, 0, len(scores))
	for id, score := range scores {
		hits = append(hits, SearchHit{Entry: s.Entries[id], Score: score})
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		if hits[i].Entry.DocTitle != hits[j].Entry.DocTitle {
			return hits[i].Entry.DocTitle < hits[j].Entry.DocTitle
		}
		return hits[i].Entry.Heading < hits[j].Entry.Heading
	})
	if limit <= 0 || limit >= len(hits) {
		return hits
	}
	return hits[:limit]
}

func entriesForDoc(doc *Doc) []SearchEntry {
	if doc == nil {
		return nil
	}
	body := doc.Summary
	if body == "" {
		body = trimToLength(plainTextFromLines(doc.Lines), 500)
	}
	entries := []SearchEntry{
		{
			DocID:    doc.ID,
			DocTitle: doc.Title,
			Text:     body,
			Weight:   1,
		},
	}
	for _, heading := range doc.Headings {
		entries = append(entries, SearchEntry{
			DocID:     doc.ID,
			DocTitle:  doc.Title,
			Heading:   heading.Text,
			HeadingID: heading.ID,
			Text:      heading.Text,
			Weight:    2,
		})
	}
	return entries
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	var current strings.Builder
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			current.WriteRune(r)
			continue
		}
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

func uniqueTokens(tokens []string) []string {
	seen := make(map[string]struct{}, len(tokens))
	unique := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token == "" {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		unique = append(unique, token)
	}
	return unique
}

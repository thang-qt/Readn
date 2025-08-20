package hackernews

import (
	"github.com/nkanaev/yarr/src/content/discussion"
)

// NewHackerNewsProvider creates a new HackerNews provider
func NewHackerNewsProvider() discussion.Provider {
	return &discussion.HackerNewsProvider{}
}

// IsHackerNewsItem checks if an RSS item is from HackerNews (legacy function)
func IsHackerNewsItem(itemURL string) bool {
	provider := NewHackerNewsProvider()
	return provider.IsProviderItem(itemURL)
}

// ExtractHNItemIDFromContent tries to extract HN item ID from item description (legacy function)
func ExtractHNItemIDFromContent(content string) (int, error) {
	provider := NewHackerNewsProvider()
	return provider.ExtractItemIDFromContent(content)
}

// ExtractHNItemID extracts the HN item ID from a HN URL (legacy function)
func ExtractHNItemID(url string) (int, error) {
	provider := NewHackerNewsProvider()
	return provider.ExtractItemID(url)
}

// GetHNThread scrapes a HackerNews discussion thread (legacy function)
func GetHNThread(itemID int) (*discussion.Thread, error) {
	provider := NewHackerNewsProvider()
	return provider.GetThread(itemID)
}

// GetHNThreadAsHTML converts a HN thread to HTML for display (legacy function)
func GetHNThreadAsHTML(thread *discussion.Thread) string {
	return discussion.GetThreadAsHTML(thread)
}

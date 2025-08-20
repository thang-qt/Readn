package discussion

import (
	"fmt"
	"strings"
)

// Comment represents a generic discussion comment
type Comment struct {
	ID       int       `json:"id"`
	Author   string    `json:"author"`
	Content  string    `json:"content"`
	Time     string    `json:"time"`
	Level    int       `json:"level"`
	Children []Comment `json:"children,omitempty"`
}

// Thread represents a generic discussion thread
type Thread struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	URL      string    `json:"url"`
	Content  string    `json:"content"`
	Author   string    `json:"author"`
	Points   int       `json:"points"`
	Time     string    `json:"time"`
	Comments []Comment `json:"comments"`
	Provider string    `json:"provider"` // "hn", "lobsters", etc.
}

// Provider defines the interface for discussion providers
type Provider interface {
	Name() string
	ThemeClass() string
	IsProviderItem(itemURL string) bool
	ExtractItemID(url string) (int, error)
	ExtractItemIDFromContent(content string) (int, error)
	GetThread(itemID int) (*Thread, error)
}

// GetThreadAsHTML converts a discussion thread to HTML for display
func GetThreadAsHTML(thread *Thread) string {
	var html strings.Builder
	
	// Thread container with provider-specific theme class
	providerClass := ""
	switch thread.Provider {
	case "hn":
		providerClass = "hn"
	case "lobsters":
		providerClass = "lobsters"
	default:
		providerClass = "generic"
	}
	
	html.WriteString(`<div class="discussion-thread discussion-` + providerClass + `">`)
	
	// Title with consistent styling
	html.WriteString(`<h1 class="discussion-title mb-2">`)
	html.WriteString(thread.Title)
	html.WriteString(`</h1>`)
	
	// URL link (similar to yarr's link styling)
	if thread.URL != "" && !strings.HasPrefix(thread.URL, "item?id=") {
		html.WriteString(`<div class="discussion-url mb-2">`)
		html.WriteString(`<a href="` + thread.URL + `" target="_blank" rel="noopener noreferrer" class="text-muted">` + thread.URL + `</a>`)
		html.WriteString(`</div>`)
	}
	
	// Metadata with consistent styling (similar to yarr's item metadata)
	html.WriteString(`<div class="discussion-meta text-muted mb-3">`)
	if thread.Author != "" {
		html.WriteString(`by <span class="discussion-author">` + thread.Author + `</span>`)
	}
	if thread.Time != "" {
		html.WriteString(` <span class="discussion-time">` + thread.Time + `</span>`)
	}
	html.WriteString(`</div>`)
	
	// Main post content (for text posts)
	if thread.Content != "" {
		html.WriteString(`<div class="discussion-content mb-4">`)
		html.WriteString(thread.Content)
		html.WriteString(`</div>`)
	}
	
	// Comments section with consistent header
	if len(thread.Comments) > 0 {
		html.WriteString(`<hr>`)
		html.WriteString(`<div class="discussion-comments">`)
		html.WriteString(`<h3 class="mb-3">`)
		html.WriteString(pluralize(len(thread.Comments), "comment"))
		html.WriteString(`</h3>`)
		
		html.WriteString(renderNestedComments(thread.Comments, 0))
		
		html.WriteString(`</div>`)
	}
	
	html.WriteString(`</div>`)
	
	return html.String()
}

func renderNestedComments(comments []Comment, level int) string {
	var html strings.Builder
	
	for i := 0; i < len(comments); i++ {
		comment := comments[i]
		if comment.Level != level {
			continue
		}
		
		html.WriteString(`<div class="discussion-comment" data-comment-id="`)
		html.WriteString(stringFromInt(comment.ID))
		html.WriteString(`" data-level="`)
		html.WriteString(stringFromInt(comment.Level))
		html.WriteString(`">`)
		
		// Comment header
		html.WriteString(`<div class="discussion-comment-header">`)
		html.WriteString(`<button class="discussion-comment-toggle" onclick="discussionToggleComment(this)" title="Toggle this comment and its replies" data-expanded="true">`)
		html.WriteString(`<span class="discussion-toggle-icon-expanded"><span class="icon"><svg width="1rem" height="1rem" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m6 9 6 6 6-6"/></svg></span></span>`)
		html.WriteString(`<span class="discussion-toggle-icon-collapsed" style="display: none;"><span class="icon"><svg width="1rem" height="1rem" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m9 18 6-6-6-6"/></svg></span></span>`)
		html.WriteString(`</button>`)
		html.WriteString(` <span class="discussion-comment-author">` + comment.Author + `</span>`)
		if comment.Time != "" {
			html.WriteString(` <span class="discussion-comment-time">` + comment.Time + `</span>`)
		}
		html.WriteString(` <span class="discussion-comment-separator">|</span> <button class="discussion-nav-btn" onclick="discussionPrevComment(this)" title="Previous comment">prev</button>`)
		html.WriteString(` <span class="discussion-comment-separator">|</span> <button class="discussion-nav-btn" onclick="discussionNextComment(this)" title="Next comment">next</button>`)
		html.WriteString(`</div>`)
		
		// Comment content
		html.WriteString(`<div class="discussion-comment-body">`)
		html.WriteString(`<div class="discussion-comment-content">`)
		html.WriteString(comment.Content)
		html.WriteString(`</div>`)
		html.WriteString(`</div>`)
		
		// Find and render replies
		var replies []Comment
		for j := i + 1; j < len(comments) && comments[j].Level > comment.Level; j++ {
			replies = append(replies, comments[j])
		}
		
		if len(replies) > 0 {
			html.WriteString(`<div class="discussion-comment-replies">`)
			html.WriteString(renderNestedComments(replies, level+1))
			html.WriteString(`</div>`)
		}
		
		html.WriteString(`</div>`)
	}
	
	return html.String()
}

func pluralize(count int, word string) string {
	countStr := stringFromInt(count)
	if count == 1 {
		return countStr + " " + word
	}
	return countStr + " " + word + "s"
}

func stringFromInt(i int) string {
	return strings.Trim(strings.Replace(fmt.Sprintf("%d", i), "\x00", "", -1), " ")
}

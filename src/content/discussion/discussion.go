package discussion

import (
	"fmt"
	"strings"
)

// Comment represents a comment in a discussion thread
type Comment struct {
	ID       int       `json:"id"`
	Author   string    `json:"author"`
	Time     string    `json:"time"`
	Content  string    `json:"content"`
	Level    int       `json:"level"`    // For flat representation
	Children []Comment `json:"children"` // For hierarchical representation
}

// Thread represents a discussion thread (story + comments)
type Thread struct {
	Title    string    `json:"title"`
	Author   string    `json:"author"`
	Time     string    `json:"time"`
	Content  string    `json:"content"`
	URL      string    `json:"url"`
	Provider string    `json:"provider"`
	Comments []Comment `json:"comments"`
}

// CountAllComments counts total comments including nested ones
func CountAllComments(comments []Comment) int {
	total := len(comments)
	for _, comment := range comments {
		total += CountAllComments(comment.Children)
	}
	return total
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
		html.WriteString(pluralize(CountAllComments(thread.Comments), "comment"))
		html.WriteString(`</h3>`)
		
		// Check if comments use hierarchical (Children) or flat (Level) structure
		if len(thread.Comments) > 0 && len(thread.Comments[0].Children) > 0 {
			// Hierarchical structure (Lobsters)
			html.WriteString(renderHierarchicalComments(thread.Comments, 0))
		} else {
			// Flat structure (HN)
			html.WriteString(renderFlatComments(thread.Comments, 0))
		}
		
		html.WriteString(`</div>`)
	}
	
	html.WriteString(`</div>`)
	
	return html.String()
}

func renderHierarchicalComments(comments []Comment, level int) string {
	var html strings.Builder
	
	for _, comment := range comments {
		html.WriteString(`<div class="discussion-comment" data-comment-id="`)
		html.WriteString(stringFromInt(comment.ID))
		html.WriteString(`" data-level="`)
		html.WriteString(stringFromInt(level))
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
		
		// Render nested replies directly from Children array
		if len(comment.Children) > 0 {
			html.WriteString(`<div class="discussion-comment-replies">`)
			html.WriteString(renderHierarchicalComments(comment.Children, level+1))
			html.WriteString(`</div>`)
		}
		
		html.WriteString(`</div>`)
	}
	
	return html.String()
}

func renderFlatComments(comments []Comment, targetLevel int) string {
	var html strings.Builder
	
	for i := 0; i < len(comments); i++ {
		comment := comments[i]
		if comment.Level != targetLevel {
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
		
		// Find and render replies from flat array
		var replies []Comment
		for j := i + 1; j < len(comments) && comments[j].Level > comment.Level; j++ {
			replies = append(replies, comments[j])
		}
		
		if len(replies) > 0 {
			html.WriteString(`<div class="discussion-comment-replies">`)
			html.WriteString(renderFlatComments(replies, targetLevel+1))
			html.WriteString(`</div>`)
		}
		
		html.WriteString(`</div>`)
		
		// Skip the replies we just processed
		for j := i + 1; j < len(comments) && comments[j].Level > comment.Level; j++ {
			i = j
		}
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

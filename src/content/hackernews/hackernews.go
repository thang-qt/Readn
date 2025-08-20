package hackernews

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// HNComment represents a HackerNews comment
type HNComment struct {
	ID       int        `json:"id"`
	Author   string     `json:"author"`
	Content  string     `json:"content"`
	Time     string     `json:"time"`
	Level    int        `json:"level"`
	Children []HNComment `json:"children,omitempty"`
}

// HNThread represents a HackerNews thread
type HNThread struct {
	ID       int         `json:"id"`
	Title    string      `json:"title"`
	URL      string      `json:"url"`
	Content  string      `json:"content"`
	Author   string      `json:"author"`
	Points   int         `json:"points"`
	Time     string      `json:"time"`
	Comments []HNComment `json:"comments"`
}

// IsHackerNewsItem checks if an RSS item is from HackerNews
func IsHackerNewsItem(itemURL string) bool {
	return strings.Contains(itemURL, "news.ycombinator.com") || 
		   strings.Contains(itemURL, "ycombinator.com")
}

// ExtractHNItemIDFromContent tries to extract HN item ID from item description
func ExtractHNItemIDFromContent(content string) (int, error) {
	// Look for HN discussion link in content
	re := regexp.MustCompile(`https://news\.ycombinator\.com/item\?id=(\d+)`)
	matches := re.FindStringSubmatch(content)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no HN item ID found in content")
	}
	
	id, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid item ID: %s", matches[1])
	}
	
	return id, nil
}

// ExtractHNItemID extracts the HN item ID from a HN URL
func ExtractHNItemID(url string) (int, error) {
	re := regexp.MustCompile(`id=(\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no item ID found in URL: %s", url)
	}
	
	id, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid item ID: %s", matches[1])
	}
	
	return id, nil
}

// GetHNThread scrapes a HackerNews discussion thread
func GetHNThread(itemID int) (*HNThread, error) {
	url := fmt.Sprintf("https://news.ycombinator.com/item?id=%d", itemID)
	
	// Fetch the page
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HN page: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}
	
	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}
	
	thread := &HNThread{
		ID:       itemID,
		Comments: []HNComment{},
	}
	
	// Extract thread details from the main story
	storyCell := doc.Find(".athing").First()
	if storyCell.Length() > 0 {
		thread.Title = storyCell.Find(".storylink").Text()
		thread.URL = storyCell.Find(".storylink").AttrOr("href", "")
	}
	
	// Extract metadata (points, author, time)
	subtext := doc.Find(".subtext").First()
	if subtext.Length() > 0 {
		pointsText := subtext.Find(".score").Text()
		if pointsText != "" {
			// Extract number from "123 points"
			re := regexp.MustCompile(`(\d+)\s+point`)
			matches := re.FindStringSubmatch(pointsText)
			if len(matches) >= 2 {
				thread.Points, _ = strconv.Atoi(matches[1])
			}
		}
		
		thread.Author = subtext.Find(".hnuser").Text()
		thread.Time = subtext.Find(".age").Text()
	}
	
	// Extract main post content (for text posts like Ask HN, Show HN, etc.)
	toptext := doc.Find(".toptext")
	if toptext.Length() > 0 {
		content, _ := toptext.Html()
		thread.Content = strings.TrimSpace(content)
	}
	
	// Extract comments
	comments := extractComments(doc)
	thread.Comments = comments
	
	return thread, nil
}

// extractComments recursively extracts comments from the HTML document
func extractComments(doc *goquery.Document) []HNComment {
	var comments []HNComment
	
	// Find all comment rows
	doc.Find("tr.athing").Each(func(i int, s *goquery.Selection) {
		// Skip if this is the main story (it has .athing but no .comtr class)
		if !s.HasClass("comtr") {
			return
		}
		
		comment := extractSingleComment(s)
		if comment != nil {
			comments = append(comments, *comment)
		}
	})
	
	return comments
}

// extractSingleComment extracts a single comment from a table row
func extractSingleComment(s *goquery.Selection) *HNComment {
	// Get comment ID
	idStr, exists := s.Attr("id")
	if !exists {
		return nil
	}
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil
	}
	
	// Extract indentation level (30px per level)
	indent := s.Find(".ind img").AttrOr("width", "0")
	level := 0
	if indentInt, err := strconv.Atoi(indent); err == nil {
		level = indentInt / 40 // HN uses 40px indentation per level
	}
	
	// Extract author
	author := s.Find(".hnuser").Text()
	
	// Extract time
	timeStr := s.Find(".age").Text()
	
	// Extract comment content
	commentSpan := s.Find(".commtext")
	content, _ := commentSpan.Html()
	content = strings.TrimSpace(content)
	
	// Skip deleted/dead comments
	if content == "" || author == "" {
		return nil
	}
	
	return &HNComment{
		ID:      id,
		Author:  author,
		Content: content,
		Time:    timeStr,
		Level:   level,
	}
}

// GetHNThreadAsHTML converts a HN thread to HTML for display
func GetHNThreadAsHTML(thread *HNThread) string {
	var html strings.Builder
	
	// Thread header with yarr-style spacing and typography
	html.WriteString(`<div class="hn-thread">`)
	

	
	// Title with consistent styling
	html.WriteString(fmt.Sprintf(`<h1 class="hn-title mb-2">%s</h1>`, thread.Title))
	
	// URL link (similar to yarr's link styling)
	if thread.URL != "" && !strings.HasPrefix(thread.URL, "item?id=") {
		html.WriteString(`<div class="hn-url mb-2">`)
		html.WriteString(fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer" class="text-muted">%s</a>`, thread.URL, thread.URL))
		html.WriteString(`</div>`)
	}
	
	// Metadata with consistent styling (similar to yarr's item metadata)
	html.WriteString(`<div class="hn-meta text-muted mb-3">`)
	if thread.Author != "" {
		html.WriteString(fmt.Sprintf(`by <span class="hn-author">%s</span>`, thread.Author))
	}
	if thread.Time != "" {
		html.WriteString(fmt.Sprintf(` <span class="hn-time">%s</span>`, thread.Time))
	}
	html.WriteString(`</div>`)
	
	// Main post content (for text posts)
	if thread.Content != "" {
		html.WriteString(`<div class="hn-content mb-4">`)
		html.WriteString(thread.Content)
		html.WriteString(`</div>`)
	}
	
	// Comments section with consistent header
	if len(thread.Comments) > 0 {
		html.WriteString(`<hr>`)
		html.WriteString(`<div class="hn-comments">`)
		html.WriteString(fmt.Sprintf(`<h3 class="mb-3">%d comment%s</h3>`, len(thread.Comments), pluralize(len(thread.Comments))))
		
		html.WriteString(renderNestedComments(thread.Comments, 0))
		
		html.WriteString(`</div>`)
	}
	
	html.WriteString(`</div>`)
	
	return html.String()
}

func renderNestedComments(comments []HNComment, level int) string {
	var html strings.Builder
	
	for i := 0; i < len(comments); i++ {
		comment := comments[i]
		if comment.Level != level {
			continue
		}
		
		html.WriteString(fmt.Sprintf(`<div class="hn-comment" data-comment-id="%d" data-level="%d">`, comment.ID, comment.Level))
		
		// Comment header
		html.WriteString(`<div class="hn-comment-header">`)
		html.WriteString(`<button class="hn-comment-toggle" onclick="hnToggleComment(this)" title="Toggle this comment and its replies" data-expanded="true">`)
		html.WriteString(`<span class="hn-toggle-icon-expanded"><span class="icon"><svg width="1rem" height="1rem" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m6 9 6 6 6-6"/></svg></span></span>`)
		html.WriteString(`<span class="hn-toggle-icon-collapsed" style="display: none;"><span class="icon"><svg width="1rem" height="1rem" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m9 18 6-6-6-6"/></svg></span></span>`)
		html.WriteString(`</button>`)
		html.WriteString(fmt.Sprintf(` <span class="hn-comment-author">%s</span>`, comment.Author))
		if comment.Time != "" {
			html.WriteString(fmt.Sprintf(` <span class="hn-comment-time">%s</span>`, comment.Time))
		}
		html.WriteString(` <span class="hn-comment-separator">|</span> <button class="hn-nav-btn" onclick="hnPrevComment(this)" title="Previous comment">prev</button>`)
		html.WriteString(` <span class="hn-comment-separator">|</span> <button class="hn-nav-btn" onclick="hnNextComment(this)" title="Next comment">next</button>`)
		html.WriteString(`</div>`)
		
		// Comment content
		html.WriteString(`<div class="hn-comment-body">`)
		html.WriteString(`<div class="hn-comment-content">`)
		html.WriteString(comment.Content)
		html.WriteString(`</div>`)
		html.WriteString(`</div>`)
		
		// Find and render replies
		var replies []HNComment
		for j := i + 1; j < len(comments) && comments[j].Level > comment.Level; j++ {
			replies = append(replies, comments[j])
		}
		
		if len(replies) > 0 {
			html.WriteString(`<div class="hn-comment-replies">`)
			html.WriteString(renderNestedComments(replies, level+1))
			html.WriteString(`</div>`)
		}
		
		html.WriteString(`</div>`)
	}
	
	return html.String()
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

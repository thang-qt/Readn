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
	
	// Thread header
	html.WriteString(`<div class="hn-thread">`)
	html.WriteString(fmt.Sprintf(`<h1 class="hn-title">%s</h1>`, thread.Title))
	
	if thread.URL != "" && !strings.HasPrefix(thread.URL, "item?id=") {
		html.WriteString(fmt.Sprintf(`<div class="hn-url"><a href="%s" target="_blank" rel="noopener noreferrer">%s</a></div>`, thread.URL, thread.URL))
	}
	
	html.WriteString(`<div class="hn-meta">`)
	if thread.Author != "" {
		html.WriteString(fmt.Sprintf(`by <span class="hn-author">%s</span>`, thread.Author))
	}
	if thread.Time != "" {
		html.WriteString(fmt.Sprintf(` <span class="hn-time">%s</span>`, thread.Time))
	}
	html.WriteString(`</div>`)
	
	// Main post content (for text posts)
	if thread.Content != "" {
		html.WriteString(`<div class="hn-content">`)
		html.WriteString(thread.Content)
		html.WriteString(`</div>`)
	}
	
	// Comments section
	if len(thread.Comments) > 0 {
		html.WriteString(`<div class="hn-comments">`)
		html.WriteString(fmt.Sprintf(`<h3>%d comment%s</h3>`, len(thread.Comments), pluralize(len(thread.Comments))))
		
		for _, comment := range thread.Comments {
			html.WriteString(formatCommentAsHTML(comment))
		}
		
		html.WriteString(`</div>`)
	}
	
	html.WriteString(`</div>`)
	
	return html.String()
}

func formatCommentAsHTML(comment HNComment) string {
	var html strings.Builder
	
	html.WriteString(fmt.Sprintf(`<div class="hn-comment" style="margin-left: %dpx;">`, comment.Level*20))
	html.WriteString(`<div class="hn-comment-meta">`)
	html.WriteString(fmt.Sprintf(`<span class="hn-comment-author">%s</span>`, comment.Author))
	if comment.Time != "" {
		html.WriteString(fmt.Sprintf(` <span class="hn-comment-time">%s</span>`, comment.Time))
	}
	html.WriteString(`</div>`)
	html.WriteString(fmt.Sprintf(`<div class="hn-comment-content">%s</div>`, comment.Content))
	html.WriteString(`</div>`)
	
	return html.String()
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

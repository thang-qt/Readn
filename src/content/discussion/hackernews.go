package discussion

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// HackerNewsProvider implements the Provider interface for HackerNews
type HackerNewsProvider struct{}

func (p *HackerNewsProvider) Name() string {
	return "HackerNews"
}

func (p *HackerNewsProvider) ThemeClass() string {
	return "hn"
}

// IsProviderItem checks if an RSS item is from HackerNews
func (p *HackerNewsProvider) IsProviderItem(itemURL string) bool {
	return strings.Contains(itemURL, "news.ycombinator.com") || 
		   strings.Contains(itemURL, "ycombinator.com")
}

// ExtractItemIDFromContent tries to extract HN item ID from item description
func (p *HackerNewsProvider) ExtractItemIDFromContent(content string) (int, error) {
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

// ExtractItemID extracts the HN item ID from a HN URL
func (p *HackerNewsProvider) ExtractItemID(url string) (int, error) {
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

// GetThread scrapes a HackerNews discussion thread
func (p *HackerNewsProvider) GetThread(itemID int) (*Thread, error) {
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
	
	thread := &Thread{
		ID:       itemID,
		Provider: "hn",
		Comments: []Comment{},
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
	comments := p.extractComments(doc)
	thread.Comments = comments
	
	return thread, nil
}

// extractComments recursively extracts comments from the HTML document
func (p *HackerNewsProvider) extractComments(doc *goquery.Document) []Comment {
	var comments []Comment
	
	// Find all comment rows
	doc.Find("tr.athing").Each(func(i int, s *goquery.Selection) {
		// Skip if this is the main story (it has .athing but no .comtr class)
		if !s.HasClass("comtr") {
			return
		}
		
		comment := p.extractSingleComment(s)
		if comment != nil {
			comments = append(comments, *comment)
		}
	})
	
	return comments
}

// extractSingleComment extracts a single comment from a table row
func (p *HackerNewsProvider) extractSingleComment(s *goquery.Selection) *Comment {
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
	
	return &Comment{
		ID:      id,
		Author:  author,
		Content: content,
		Time:    timeStr,
		Level:   level,
	}
}

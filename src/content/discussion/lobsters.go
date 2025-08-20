package discussion

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// LobstersProvider parses comments from Lobste.rs discussion pages.
type LobstersProvider struct{}

// Name of the provider.
func (p *LobstersProvider) Name() string { return "Lobste.rs" }

// Hostname returned by the provider.
func (p *LobstersProvider) Hostname() string { return "lobste.rs" }

// Theme name for rendering.
func (p *LobstersProvider) Theme() string { return "lobsters" }

// IsValidURL checks whether the URL looks like a Lobste.rs story page.
func (p *LobstersProvider) IsValidURL(url string) bool {
	re := regexp.MustCompile(`^https?://lobste\.rs/s/[a-z0-9]+`)
	return re.MatchString(url)
}

// FetchThread fetches a discussion thread (story + comments).
func (p *LobstersProvider) FetchThread(url string) (*Thread, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	thread := &Thread{
		URL:      url,
		Provider: "lobsters",
	}

	// Extract story metadata
	story := doc.Find("li.story").First()
	if story.Length() > 0 {
		if titleLink := story.Find(".link a").First(); titleLink.Length() > 0 {
			thread.Title = strings.TrimSpace(titleLink.Text())
		}

		if authorLink := story.Find(".byline .u-author").First(); authorLink.Length() > 0 {
			thread.Author = strings.TrimSpace(authorLink.Text())
		}

		if timeElem := story.Find(".byline time").First(); timeElem.Length() > 0 {
			if title, ok := timeElem.Attr("title"); ok && title != "" {
				thread.Time = title
			} else if datetime, ok := timeElem.Attr("datetime"); ok && datetime != "" {
				thread.Time = datetime
			} else {
				thread.Time = strings.TrimSpace(timeElem.Text())
			}
		}

		if contentDiv := doc.Find(".story_content .story_text").First(); contentDiv.Length() > 0 {
			if htmlContent, err := contentDiv.Html(); err == nil {
				thread.Content = strings.TrimSpace(htmlContent)
			} else {
				thread.Content = strings.TrimSpace(contentDiv.Text())
			}
		}
	}

	comments, err := p.parseCommentsFromDoc(doc)
	if err != nil {
		return nil, err
	}
	thread.Comments = comments

	return thread, nil
}

// FetchComments fetches only the comments of a discussion.
func (p *LobstersProvider) FetchComments(url string) ([]Comment, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return p.parseCommentsFromDoc(doc)
}

// parseCommentsFromDoc extracts comments from the document.
func (p *LobstersProvider) parseCommentsFromDoc(doc *goquery.Document) ([]Comment, error) {
	var comments []Comment

	// Find comments container
	var commentsContainer *goquery.Selection
	if li := doc.Find("li#story_comments").First(); li.Length() > 0 {
		if ol := li.Find("ol.comments").First(); ol.Length() > 0 {
			commentsContainer = ol
		}
	}
	if commentsContainer == nil {
		commentsContainer = doc.Find("ol.comments").First()
		if commentsContainer.Length() == 0 {
			return comments, nil
		}
	}

	// Parse top-level comments
	topLevelItems := commentsContainer.Children().Filter("li.comments_subtree")
	topLevelItems.Each(func(i int, s *goquery.Selection) {
		if c := p.parseComment(s, 0); c != nil {
			comments = append(comments, *c)
		}
	})

	return comments, nil
}

// parseComment parses a single comment subtree recursively.
func (p *LobstersProvider) parseComment(s *goquery.Selection, depth int) *Comment {
	commentDiv := s.Find("div[data-shortid]").First()
	if commentDiv.Length() == 0 || commentDiv.HasClass("comment_form_container") {
		return nil
	}

	shortID, exists := commentDiv.Attr("data-shortid")
	if !exists || shortID == "" {
		return nil
	}

	author := ""
	commentDiv.Find("div.details div.byline a[href^='/~']").Each(func(i int, link *goquery.Selection) {
		if link.Find("img").Length() == 0 {
			author = strings.TrimSpace(link.Text())
		}
	})

	timeStr := ""
	if timeElem := commentDiv.Find("div.details div.byline time").First(); timeElem.Length() > 0 {
		if title, ok := timeElem.Attr("title"); ok && title != "" {
			timeStr = title
		} else if datetime, ok := timeElem.Attr("datetime"); ok && datetime != "" {
			timeStr = datetime
		} else {
			timeStr = strings.TrimSpace(timeElem.Text())
		}
	}

	text := ""
	if textDiv := commentDiv.Find("div.details div.comment_text").First(); textDiv.Length() > 0 {
		if htmlContent, err := textDiv.Html(); err == nil {
			text = strings.TrimSpace(htmlContent)
		} else {
			text = strings.TrimSpace(textDiv.Text())
		}
	}

	// Deterministic numeric ID
	var id int
	for i, r := range shortID {
		id = id*31 + int(r) + i
	}
	if id < 0 {
		id = -id
	}

	comment := &Comment{
		ID:      id,
		Author:  author,
		Time:    timeStr,
		Content: text,
		Level:   depth,
	}

	// Parse nested replies
	if nestedList := s.Children().Filter("ol.comments"); nestedList.Length() > 0 {
		nestedItems := nestedList.Children().Filter("li.comments_subtree")
		nestedItems.Each(func(i int, nestedSel *goquery.Selection) {
			if nested := p.parseComment(nestedSel, depth+1); nested != nil && nested.ID != comment.ID {
				comment.Children = append(comment.Children, *nested)
			}
		})
	}

	return comment
}

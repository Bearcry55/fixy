package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

func newClient() *http.Client {
	return &http.Client{Timeout: 8 * time.Second}
}

func debugLog(msg string) {
	if os.Getenv("FIXY_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "%s[DEBUG] %s%s\n", dim, msg, reset)
	}
}

func fetchStackOverflow(wg *sync.WaitGroup, query, tag string, ch chan<- Result) {
	defer wg.Done()

	apiURL := fmt.Sprintf(
		"https://api.stackexchange.com/2.3/search/advanced?order=desc&sort=relevance&q=%s&site=stackoverflow&filter=default",
		url.QueryEscape(query),
	)
	if tag != "" {
		apiURL += "&tagged=" + tag
	}

	res, err := newClient().Get(apiURL)
	if err != nil {
		debugLog("StackOverflow: " + err.Error())
		return
	}
	defer res.Body.Close()

	var data struct {
		Items []struct {
			Title       string `json:"title"`
			Link        string `json:"link"`
			Score       int    `json:"score"`
			IsAnswered  bool   `json:"is_answered"`
			AnswerCount int    `json:"answer_count"`
		} `json:"items"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		debugLog("StackOverflow decode: " + err.Error())
		return
	}

	for i, item := range data.Items {
		if i >= 3 {
			break
		}
		answered := "Unanswered"
		if item.IsAnswered {
			answered = "✓ Answered"
		}
		ch <- Result{
			Source: "StackOverflow",
			Title:  item.Title,
			Link:   item.Link,
			Info:   fmt.Sprintf("Votes: %d | Answers: %d | %s", item.Score, item.AnswerCount, answered),
		}
	}
}

func fetchGitHubIssues(wg *sync.WaitGroup, query, tag string, ch chan<- Result) {
	defer wg.Done()

	fullQuery := query
	if tag != "" {
		fullQuery = fmt.Sprintf("language:%s %s", tag, query)
	}

	req, _ := http.NewRequest("GET",
		fmt.Sprintf("https://api.github.com/search/issues?q=%s&type=issue", url.QueryEscape(fullQuery)),
		nil,
	)
	req.Header.Set("User-Agent", "Fixy-CLI")
	req.Header.Set("Accept", "application/vnd.github+json")

	res, err := newClient().Do(req)
	if err != nil {
		debugLog("GitHub: " + err.Error())
		return
	}
	defer res.Body.Close()

	if res.StatusCode == 403 || res.StatusCode == 429 {
		debugLog("GitHub: rate limited")
		return
	}

	var data struct {
		Items []struct {
			Title    string `json:"title"`
			HTMLUrl  string `json:"html_url"`
			State    string `json:"state"`
			Comments int    `json:"comments"`
		} `json:"items"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		debugLog("GitHub decode: " + err.Error())
		return
	}

	for i, item := range data.Items {
		if i >= 3 {
			break
		}
		ch <- Result{
			Source: "GitHub Issues",
			Title:  item.Title,
			Link:   item.HTMLUrl,
			Info:   fmt.Sprintf("State: %s | Comments: %d", item.State, item.Comments),
		}
	}
}

func fetchReddit(wg *sync.WaitGroup, query, tag string, ch chan<- Result) {
	defer wg.Done()

	subreddit := "programming+learnprogramming+webdev+devops"
	switch tag {
	case "go":
		subreddit = "golang"
	case "python":
		subreddit = "learnpython+python"
	case "javascript":
		subreddit = "javascript+learnjavascript+node"
	case "rust":
		subreddit = "rust"
	case "java":
		subreddit = "java+learnprogramming"
	case "cpp":
		subreddit = "cpp+learnprogramming"
	}

	req, _ := http.NewRequest("GET",
		fmt.Sprintf("https://www.reddit.com/r/%s/search.json?q=%s&restrict_sr=1&sort=relevance&limit=5",
			subreddit, url.QueryEscape(query)),
		nil,
	)
	req.Header.Set("User-Agent", "fixy-cli/1.0")

	res, err := newClient().Do(req)
	if err != nil {
		debugLog("Reddit: " + err.Error())
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		debugLog(fmt.Sprintf("Reddit status: %d", res.StatusCode))
		return
	}

	var data struct {
		Data struct {
			Children []struct {
				Data struct {
					Title       string `json:"title"`
					URL         string `json:"url"`
					Score       int    `json:"score"`
					NumComments int    `json:"num_comments"`
					Subreddit   string `json:"subreddit"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		debugLog("Reddit decode: " + err.Error())
		return
	}

	for i, child := range data.Data.Children {
		if i >= 3 {
			break
		}
		item := child.Data
		ch <- Result{
			Source: "Reddit",
			Title:  item.Title,
			Link:   item.URL,
			Info:   fmt.Sprintf("r/%s | ⬆ %d | 💬 %d comments", item.Subreddit, item.Score, item.NumComments),
		}
	}
}
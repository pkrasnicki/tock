package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-faster/errors"
)

type Client struct {
	baseURL    string
	username   string
	apiToken   string
	httpClient *http.Client
}

type WorklogRequest struct {
	Comment          *Comment `json:"comment,omitempty"`
	Started          string   `json:"started"` // Format: 2026-02-01T14:15:00.000+0000
	TimeSpentSeconds int      `json:"timeSpentSeconds"`
}

type Comment struct {
	Type    string    `json:"type"`
	Version int       `json:"version"`
	Content []Content `json:"content"`
}

type Content struct {
	Type    string      `json:"type"`
	Content []TextBlock `json:"content,omitempty"`
	Text    string      `json:"text,omitempty"`
}

type TextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type WorklogResponse struct {
	ID               string   `json:"id"`
	IssueID          string   `json:"issueId"`
	Comment          *Comment `json:"comment,omitempty"`
	Started          string   `json:"started"`
	TimeSpentSeconds int      `json:"timeSpentSeconds"`
}

func NewClient(baseURL, username, apiToken string) *Client {
	return &Client{
		baseURL:    baseURL,
		username:   username,
		apiToken:   apiToken,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// AddWorklog adds a worklog to a Jira issue
func (c *Client) AddWorklog(issueKey string, req WorklogRequest) (*WorklogResponse, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/worklog", c.baseURL, issueKey)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "marshal request")
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrap(err, "create request")
	}

	httpReq.SetBasicAuth(c.username, c.apiToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, errors.Errorf("jira API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var worklog WorklogResponse
	if err := json.NewDecoder(resp.Body).Decode(&worklog); err != nil {
		return nil, errors.Wrap(err, "decode response")
	}

	return &worklog, nil
}

// UpdateWorklog updates an existing worklog
func (c *Client) UpdateWorklog(issueKey, worklogID string, req WorklogRequest) error {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/worklog/%s", c.baseURL, issueKey, worklogID)

	body, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "marshal request")
	}

	httpReq, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return errors.Wrap(err, "create request")
	}

	httpReq.SetBasicAuth(c.username, c.apiToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return errors.Wrap(err, "execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return errors.Errorf("jira API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// DeleteWorklog deletes a worklog from a Jira issue
func (c *Client) DeleteWorklog(issueKey, worklogID string) error {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/worklog/%s", c.baseURL, issueKey, worklogID)

	httpReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return errors.Wrap(err, "create request")
	}

	httpReq.SetBasicAuth(c.username, c.apiToken)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return errors.Wrap(err, "execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return errors.Errorf("jira API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// NewComment creates a comment in Atlassian Document Format (ADF)
func NewComment(text string) *Comment {
	if text == "" {
		return nil
	}
	return &Comment{
		Type:    "doc",
		Version: 1,
		Content: []Content{
			{
				Type: "paragraph",
				Content: []TextBlock{
					{
						Type: "text",
						Text: text,
					},
				},
			},
		},
	}
}

// IssueSearchResult represents the response from Jira issue search
type IssueSearchResult struct {
	Issues []Issue `json:"issues"`
	Total  int     `json:"total"`
}

// Issue represents a simplified Jira issue
type Issue struct {
	Key    string      `json:"key"`
	Fields IssueFields `json:"fields"`
}

// IssueFields represents the fields of a Jira issue
type IssueFields struct {
	Summary string `json:"summary"`
}

// SearchRequest represents the request body for Jira search API
type SearchRequest struct {
	JQL        string   `json:"jql"`
	MaxResults int      `json:"maxResults"`
	Fields     []string `json:"fields"`
}

// SearchIssues searches for Jira issues using JQL (Jira Query Language)
// Returns issues matching the query, limited to maxResults
func (c *Client) SearchIssues(query string, maxResults int) ([]Issue, error) {
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 50 {
		maxResults = 50 // Cap at 50 to prevent large responses
	}

	// Build JQL query to search in summary and key
	jql := fmt.Sprintf("text ~ \"%s*\" OR key ~ \"%s*\" ORDER BY updated DESC", query, query)

	searchReq := SearchRequest{
		JQL:        jql,
		MaxResults: maxResults,
		Fields:     []string{"summary", "key"},
	}

	body, err := json.Marshal(searchReq)
	if err != nil {
		return nil, errors.Wrap(err, "marshal request")
	}

	url := fmt.Sprintf("%s/rest/api/3/search/jql", c.baseURL)

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrap(err, "create request")
	}

	httpReq.SetBasicAuth(c.username, c.apiToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, errors.Errorf("jira API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var result IssueSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decode response")
	}

	return result.Issues, nil
}

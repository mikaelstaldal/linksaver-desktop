package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Item struct {
	ID          int64     `json:"ID,omitempty"`
	URL         string    `json:"URL"`
	Title       string    `json:"Title"`
	Description string    `json:"Description,omitempty"`
	AddedAt     time.Time `json:"AddedAt,omitempty"`
}

func (i Item) IsNote() bool {
	return strings.HasPrefix(i.URL, "note:")
}

type APIClient struct {
	baseURL string
	client  *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				// Set reasonable timeouts
				IdleConnTimeout:       30 * time.Second,
				TLSHandshakeTimeout:   5 * time.Second,
				ResponseHeaderTimeout: 5 * time.Second,
			},
		},
	}
}

func (c *APIClient) GetItems(searchTerm string) ([]Item, error) {
	u := c.baseURL + "/"
	if searchTerm != "" {
		u += "?s=" + searchTerm
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get items: %s", resp.Status)
	}

	var links []Item
	if err := json.NewDecoder(resp.Body).Decode(&links); err != nil {
		return nil, err
	}
	return links, nil
}

func (c *APIClient) GetItem(id string) (*Item, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", c.baseURL, id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get item: %s", resp.Status)
	}

	var link Item
	if err := json.NewDecoder(resp.Body).Decode(&link); err != nil {
		return nil, err
	}
	return &link, nil
}

func (c *APIClient) AddLink(newUrl string) error {
	form := url.Values{}
	form.Set("url", newUrl)
	data := form.Encode()

	req, err := http.NewRequest("POST", c.baseURL+"/", bytes.NewBufferString(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add link: %s", resp.Status)
	}
	return nil
}

func (c *APIClient) AddNote(title string, text string) error {
	form := url.Values{}
	form.Set("note-title", title)
	form.Set("note-text", text)
	data := form.Encode()

	req, err := http.NewRequest("POST", c.baseURL+"/", bytes.NewBufferString(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add note: %s", resp.Status)
	}
	return nil
}

func (c *APIClient) UpdateItem(id string, title string, description string) error {
	form := url.Values{}
	form.Set("title", title)
	form.Set("description", description)
	data := form.Encode()

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/%s", c.baseURL, id), bytes.NewBufferString(data))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update link: %s", resp.Status)
	}
	return nil
}

func (c *APIClient) DeleteItem(id string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", c.baseURL, id), nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete link: %s", resp.Status)
	}
	return nil
}

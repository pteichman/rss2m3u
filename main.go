package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func main() {
	var client http.Client

	for _, rawURL := range os.Args[1:] {
		u, err := url.Parse(rawURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: parsing %s: %s", rawURL, err)
			continue
		}

		feed, err := fetch(client, u)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: fetching %s: %s", rawURL, err)
			continue
		}

		episodes, err := parseRss(feed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: parsing %s: %s", rawURL, err)
			continue
		}

		for i := len(episodes) - 1; i >= 0; i-- {
			ep := episodes[i]
			fmt.Println(ep.rawURL)
		}
	}
}

func fetch(client http.Client, u *url.URL) ([]byte, error) {
	req := &http.Request{
		Method: http.MethodGet,
		URL:    u,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-OK: %s", resp.Status)
	}

	body := bytes.NewBuffer(nil)
	if _, err := io.Copy(body, resp.Body); err != nil {
		return nil, err
	}

	return body.Bytes(), nil
}

type episode struct {
	title  string
	rawURL string
}

func parseRss(doc []byte) ([]episode, error) {
	var f feed
	if err := xml.Unmarshal(doc, &f); err != nil {
		return nil, fmt.Errorf("unmarshaling: %s", err)
	}

	var episodes []episode

items:
	for _, item := range f.Items {
		for _, en := range item.Enclosures {
			if en.Type == "audio/mpeg" {
				episodes = append(episodes, episode{title: item.Title, rawURL: en.URL})
				continue items
			}
		}
	}

	return episodes, nil
}

type feed struct {
	XMLName xml.Name `xml:"rss"`

	Title       string `xml:"channel>title"`
	Description string `xml:"channel>description"`

	Items []item `xml:"channel>item"`
}

type item struct {
	Title      string      `xml:"title"`
	Enclosures []enclosure `xml:"enclosure"`
}

type enclosure struct {
	URL  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}

package taninari

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"time"
)

const blogPostEndpoint = "https://api.amebaowndme.com/v2/public/blogPosts?siteId=18381&searchType=recent&limit=15"

var tagRegexp = regexp.MustCompile(`<[\S\s]+?>`)

type BlogPost struct {
	Meta struct {
		Code       int `json:"code"`
		Pagination struct {
			Total   int `json:"total"`
			Offset  int `json:"offset"`
			Limit   int `json:"limit"`
			Cursors struct {
				After  string `json:"after"`
				Before string `json:"before"`
			} `json:"cursors"`
		} `json:"pagination"`
	} `json:"meta"`
	Body []struct {
		ID       string `json:"id"`
		Contents []struct {
			Type   string `json:"type"`
			Format string `json:"format"`
			Value  string `json:"value"`
			Url    string `json:"url"`
		} `json:"contents"`
		PublishedURL string `json:"publishedUrl"`
		PublishedAt  string `json:"publishedAt"`
	} `json:"body"`
}

type Goroku struct {
	Text         string
	ImageURL     string
	PublishedURL string
	PublishedAt  string
}

func getBlogPosts(api string) (string, error) {
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Error: status code is", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func parseJson(jsonStr string) (*BlogPost, error) {
	jsonBytes := []byte(jsonStr)
	blogPost := new(BlogPost)

	if err := json.Unmarshal(jsonBytes, blogPost); err != nil {
		return nil, err
	}

	return blogPost, nil
}

func GetAllGorokus() ([]*Goroku, error) {
	gorokus := []*Goroku{}

	url := blogPostEndpoint
	for {
		blogPostsJson, err := getBlogPosts(url)
		if err != nil {
			return nil, err
		}

		blogPost, err := parseJson(blogPostsJson)
		if err != nil {
			return nil, err
		}

		for _, b := range blogPost.Body {
			goroku := &Goroku{
				PublishedURL: b.PublishedURL,
				PublishedAt:  b.PublishedAt,
			}

			for _, c := range b.Contents {
				if c.Type == "text" {
					t := tagRegexp.ReplaceAllString(c.Value, "")
					goroku.Text = t
				} else if c.Type == "image" {
					goroku.ImageURL = c.Url
				}
			}

			gorokus = append(gorokus, goroku)
		}

		if len(gorokus) >= blogPost.Meta.Pagination.Total {
			break
		}

		url = blogPostEndpoint + "&cursor=" + blogPost.Meta.Pagination.Cursors.After
		time.Sleep(5 * time.Millisecond)
	}

	return gorokus, nil
}

func GetGoroku() (*Goroku, error) {
	gorokus, err := GetAllGorokus()
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(gorokus))

	return gorokus[index], nil
}

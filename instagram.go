package social

import (
	"appengine"
	"appengine/urlfetch"

	"encoding/json"

	"fmt"
	"log"
	"net/http"
	"errors"
)

type InstagramMediaImage struct {
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
	URL    string `json:"url"`
}

type InstagramMediaImages struct {
	LowResolution      *InstagramMediaImage `json:"low_resolution"`
	StandardResolution *InstagramMediaImage `json:"standard_resolution"`
	Thumbnail          *InstagramMediaImage `json:"thumbnail"`
}

type InstagramMedia struct {
	Link              string                `json:"link"`
	Location          json.RawMessage       `json:"location"`
	CanDeleteComments bool                  `json:"can_delete_comments"`
	AltMediaURL       string                `json:"alt_media_url"`
	Caption           json.RawMessage       `json:"caption"`
	Type              string                `json:"type"`
	Comments          json.RawMessage       `json:"comments"`
	Code              string                `json:"code"`
	User              json.RawMessage       `json:"user"`
	CanViewComments   bool                  `json:"can_view_comments"`
	Images            *InstagramMediaImages `json:"images"`
	CreatedTime       string                `json:"created_time"`
	ID                string                `json:"id"`
	Likes             json.RawMessage       `json:"likes"`
}

type InstagramData struct {
	Items         []*InstagramMedia `json:"items"`
	MoreAvailable bool              `json:"more_available"`
	Status        string            `json:"status"`
}

func GetDataFromInstagram(ctx *appengine.Context, username string) (*InstagramData, error) {
	if username == "" {
		return nil, errors.New("Username can't be empty")
	}

	url := fmt.Sprintf("https://instagram.com/%s/media/", username)

	c := urlfetch.Client(*ctx)
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Instagram request error. Status code: %d", resp.StatusCode)
	}

	d := json.NewDecoder(resp.Body)
	data := &InstagramData{}

	err = d.Decode(data)
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}

	return data, nil
}

package social

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
)

type CreateUserPayload struct {
	InstagramUsername string `json:"instagram_username"`
}

func GetUser(w http.ResponseWriter, r *http.Request) error {
	setResponseHeader(w)

	seg := strings.Split(r.URL.Path, "/")
	if len(seg) != 3 {
		return errors.New("Missformed URL")
	}

	err := sanitizeUsername(seg[2])
	if err != nil {
		return err
	}

	ctx := appengine.NewContext(r)
	media, err := RetrieveUserData(&ctx, seg[2])
	if err != nil {
		return err
	}

	json.NewEncoder(w).Encode(*media)
	return nil
}

func CreateUser(w http.ResponseWriter, r *http.Request) error {
	setResponseHeader(w)

	p := &CreateUserPayload{}
	err := json.NewDecoder(r.Body).Decode(p)
	if err != nil {
		return err
	}

	if p.InstagramUsername == "" {
		return errors.New("Instagram username is empty")
	}

	err = sanitizeUsername(p.InstagramUsername)
	if err != nil {
		return err
	}

	ctx := appengine.NewContext(r)
	err = AddTaskInQueue(&ctx, FetchNewUser, Payload{Username: p.InstagramUsername, Page: 0})
	if err != nil {
		return err
	}

	return nil
}

func CreateRequestQueue(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	query := datastore.NewQuery(UserKind)
	iterator := query.Run(ctx)

	for {
		var user User
		_, err := iterator.Next(&user)
		if err == datastore.Done {
			break
		}

		if err != nil {
			return err
		}

		err = AddTaskInQueue(&ctx, FetchOldUser, Payload{Username: user.InstagramUsername, Page: 0})
		if err != nil {
			return err
		}
	}

	return nil
}

func sanitizeUsername(username string) error {
	m, err := regexp.MatchString("^[a-zA-Z0-9._]+$", username)
	if err != nil || !m {
		return errors.New("Passed parameter is not Instagram Username")
	}

	return nil
}

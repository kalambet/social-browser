package social

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"log"
	"strconv"
	"time"
)

const (
	UserKind  = "User"
	MediaKind = "Media"
)

type User struct {
	InstagramUsername string
	Added             int64
	LastUpdated       int64
}

type Media struct {
	InstagramThumbnailURL string
	InstagramUsername     string
	InstagramStandardURL  string
	InstagramLowURL       string
	Created               int64
}

func SaveNewUser(ctx *appengine.Context, username string, data *InstagramData) error {
	if data == nil || len(data.Items) == 0 {
		return fmt.Errorf("Instagram data for user [%s] was not provided", username)
	}

	user, err := getUserFormStore(ctx, username)
	if err != nil {
		return err
	}

	if user != nil {
		return fmt.Errorf("User [%s] already exist\n", username)
	}

	now := time.Now().Unix()
	err = saveUserToStore(ctx, User{
		InstagramUsername: username,
		Added:             now,
		LastUpdated:       now,
	})

	if err != nil {
		return err
	}

	for _, item := range data.Items {
		err = saveInstagramMedia(ctx, username, item)
		if err != nil {
			return err
		}
	}

	return nil
}

func SaveUserMissingPhotos(ctx *appengine.Context, username string, data *InstagramData) (int, error) {
	if data == nil || len(data.Items) == 0 {
		return -1, fmt.Errorf("Instagram data for user [%s] was not provided", username)
	}

	user, err := getUserFormStore(ctx, username)
	if err != nil {
		return -1, err
	}

	if user == nil {
		return -1, SaveNewUser(ctx, username, data)
	} else {
		// Update Last Updated Time
		user.LastUpdated = time.Now().Unix()
		saveUserToStore(ctx, *user)
	}

	q := datastore.NewQuery(MediaKind).
		Filter("InstagramUsername =", username).
		Order("-Created").
		Limit(1)

	var lastPhotos []Media
	_, err = q.GetAll(*ctx, &lastPhotos)
	if err != nil {
		return -1, err
	}

	var lastUpdtaedIdx = 0
	for idx, item := range data.Items {
		t, err := strconv.ParseInt(item.CreatedTime, 10, 64)
		if err != nil {
			log.Printf("Problem pasrsing created time %s for [%s]: %s",
				item.CreatedTime,
				username,
				err.Error())
			continue
		}

		if t > lastPhotos[0].Created {
			saveInstagramMedia(ctx, username, item)
		} else {
			break
		}
		lastUpdtaedIdx = idx
	}

	return lastUpdtaedIdx, nil
}

func saveInstagramMedia(ctx *appengine.Context, username string, media *InstagramMedia) error {
	mkey := datastore.NewKey(*ctx, MediaKind, media.Code, 0, nil)

	t, err := strconv.ParseInt(media.CreatedTime, 10, 64)
	if err != nil {
		t = 0
	}
	_, err = datastore.Put(*ctx, mkey, &Media{
		InstagramUsername:     username,
		InstagramThumbnailURL: media.Images.Thumbnail.URL,
		InstagramLowURL:       media.Images.LowResolution.URL,
		InstagramStandardURL:  media.Images.StandardResolution.URL,
		Created:               t,
	})

	if err != nil {
		return err
	}

	return nil
}

func getUserFormStore(ctx *appengine.Context, username string) (*User, error) {
	ukey := datastore.NewKey(*ctx, UserKind, username, 0, nil)

	var user = &User{}
	err := datastore.Get(*ctx, ukey, user)
	if err == datastore.ErrNoSuchEntity {
		return nil, nil
	} else {
		return nil, err
	}

	return user, nil
}

func saveUserToStore(ctx *appengine.Context, user User) error {
	ukey := datastore.NewKey(*ctx, UserKind, user.InstagramUsername, 0, nil)

	_, err := datastore.Put(*ctx, ukey, &user)
	if err != nil {
		return err
	}

	return nil
}

func RetrieveUserData(ctx *appengine.Context, username string) (media *[]Media, err error) {
	q := datastore.NewQuery(MediaKind).
		Filter("InstagramUsername =", username).
		Order("-Created").
		Limit(20)

	media = &[]Media{}
	_, err = q.GetAll(*ctx, media)
	if err != nil {
		return nil, fmt.Errorf("Can't retrieve media for user [%s]: %s", username, err.Error())
	}

	return media, nil
}

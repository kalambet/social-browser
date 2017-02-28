package social_test

import (
  "testing"
  "appengine"
  "appengine/aetest"
  "github.com/kalambet/social-browser"
  "appengine/datastore"
  "time"
  "strings"
)

func TestSaveNewUser(t *testing.T) {
  var ctx appengine.Context
  ctx, err := aetest.NewContext(&aetest.Options{})
  if err != nil {
    t.Fatal(err)
  }

  data, err := social.GetDataFromInstagram(&ctx, TestInstagramUsername)
  if err != nil {
    t.Fatal(err)
  }

  if data == nil {
    t.Error("Data from istagram is empty")
  }

  now := time.Now().Unix()
  err = social.SaveNewUser(&ctx, TestInstagramUsername, data)
  if err != nil {
    t.Error(err)
  }

  u := &social.User{}
  ukey := datastore.NewKey(ctx, social.UserKind, TestInstagramUsername, 0, nil)
  err = datastore.Get(ctx, ukey, u)
  if err != nil {
    t.Error(err)
  }

  if strings.Compare(u.InstagramUsername, TestInstagramUsername) != 0 {
    t.Errorf("Stored username is not that we saved %s != %s", u.InstagramUsername, TestInstagramUsername)
  }

  if u.Added < now {
    t.Error("User added time sets incorrectly")
  }

  for _, item := range data.Items {
    mkey := datastore.NewKey(ctx, social.MediaKind, item.Code, 0, nil)

    media := &social.Media{}
    datastore.Get(ctx, mkey, media)

    if strings.Compare(media.InstagramUsername, u.InstagramUsername) != 0 {
      t.Errorf("Instagram username for %s are not the same: %s and %s",
        item.Code,
        media.InstagramUsername,
        u.InstagramUsername)
    }

    if strings.Compare(media.InstagramLowURL, item.Images.LowResolution.URL) != 0 {
      t.Errorf("Low resolution images for %s are not the same: %s and %s",
        item.Code,
        media.InstagramLowURL,
        item.Images.LowResolution.URL)
    }

    if strings.Compare(media.InstagramThumbnailURL, item.Images.Thumbnail.URL) != 0 {
      t.Errorf("Thumbnail images for %s are not the same: %s and %s",
        item.Code,
        media.InstagramThumbnailURL,
        item.Images.Thumbnail.URL)
    }

    if strings.Compare(media.InstagramStandardURL, item.Images.StandardResolution.URL) != 0 {
      t.Errorf("Standard resolution images for %s are not the same: %s and %s",
        item.Code,
        media.InstagramStandardURL,
        item.Images.StandardResolution.URL)
    }
  }
}

package social_test

import (
  "testing"
  "github.com/kalambet/social-browser"
  "appengine/aetest"
  "appengine"
  "strings"
)

const TestInstagramUsername = "vindiesel"

func TestGetDataFromInstagram(t *testing.T) {
  var ctx appengine.Context
  ctx, err := aetest.NewContext(&aetest.Options{})
  if err != nil {
    t.Fatal(err)
  }

  data, err := social.GetDataFromInstagram(&ctx, "")
  if err == nil {
    t.Error("Empty username must throw an error")
  }

  if data != nil {
    t.Error("Empty username data must be nil")
  }

  data, err = social.GetDataFromInstagram(&ctx, TestInstagramUsername)
  if err != nil {
    t.Error(err)
  }

  if data == nil {
    t.Error("Instagram data must not be empty")
  }

  if strings.Compare(data.Status, "ok") != 0 {
    t.Error("Instagram request failed")
  }

  if len(data.Items) > 20 {
    t.Error("The first batch must be less then 20 items")
  }
}

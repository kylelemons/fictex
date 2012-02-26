package ui

import (
	"fmt"

	"appengine"
	"appengine/user"
	"appengine/datastore"
)

func UserKey(c appengine.Context) (*user.User, *datastore.Key) {
	user := user.Current(c)

	uid := user.ID
	if uid == "" {
		uid = fmt.Sprintf("%s@%s",
			user.FederatedIdentity,
			user.FederatedProvider)
	}

	key := datastore.NewKey(c, "User", uid, 0, nil)

	return user, key
}

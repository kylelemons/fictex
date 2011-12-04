package ui

import (
	"os"

	"appengine"
	"appengine/datastore"
)

type Story struct {
	ctx appengine.Context
	key *datastore.Key

	Title  string
	Source []byte
}

func NewStory(c appengine.Context, id string, owner *datastore.Key) *Story {
	return &Story{
		ctx: c,
		key: datastore.NewKey(c, "Story", id, 0, owner),
	}
}

func (s *Story) Put() os.Error {
	key, err := datastore.Put(s.ctx, s.key, s)
	if err != nil {
		return err
	}

	s.key = key
	return nil
}

func (s *Story) Get() os.Error {
	return datastore.Get(s.ctx, s.key, s)
}

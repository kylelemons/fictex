package ui

import (
	"crypto/sha1"
	"fmt"
	"json"
	"os"
	"time"

	"appengine"
	"appengine/datastore"
)

func GenID(seed string) string {
	h := sha1.New()
	fmt.Fprintf(h, "%s:%s", seed, time.LocalTime())
	return fmt.Sprintf("%x", h.Sum())
}

type Story struct {
	key  *datastore.Key

	ID     string
	Title  string
	Source []byte
	Meta   map[string]*Property `datastore:"-"`
}

func NewStory(c appengine.Context, id string, owner *datastore.Key) *Story {
	return &Story{
		key:  datastore.NewKey(c, "Story", id, 0, owner),
		ID:   id,
	}
}

func (s *Story) Put(c appengine.Context) os.Error {
	return datastore.RunInTransaction(c, func(tx appengine.Context) os.Error {
		key, err := datastore.Put(tx, s.key, s)
		if err != nil {
			return err
		}

		s.key = key

		for _, prop := range s.Meta {
			if err := prop.Put(tx); err != nil {
				return err
			}
		}

		return nil
	}, nil)
}

func (s *Story) Get(c appengine.Context) os.Error {
	// Construct the query once
	q := datastore.NewQuery("Property")
	q.Ancestor(s.key)

	return datastore.RunInTransaction(c, func(tx appengine.Context) os.Error {
		if err := datastore.Get(tx, s.key, s); err != nil {
			return err
		}

		props := []*Property{}
		keys, err := q.GetAll(tx, &props)
		if err != nil {
			return err
		}

		s.Meta = make(map[string]*Property)
		for i, prop := range props {
			prop.key = keys[i]
			s.Meta[prop.Name] = prop
		}

		return nil
	}, nil)
}

func GetStory(c appengine.Context, id string) (*Story, os.Error) {
	q := datastore.NewQuery("Story")
	q.Filter("ID =", id)
	q.KeysOnly()

	keys, err := q.GetAll(c, nil)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, os.NewError(id + ": no such story")
	}

	s := &Story{ key: keys[0]}
	return s, s.Get(c)
}

func JSONStoryList(c appengine.Context, user *datastore.Key) ([]byte, os.Error) {
	type storydata struct{
		Id string `json:"id"`
		Name string `json:"name"`
	}
	var stories []storydata

	q := datastore.NewQuery("Story")
	q.Ancestor(user)
	iter := q.Run(c)

	for {
		s := new(Story)
		key, err := iter.Next(s)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		if s.Title == "" {
			continue
		}
		stories = append(stories, storydata{
			Id: key.StringID(),
			Name: s.Title,
		})
	}

	return json.MarshalIndent(stories, "", "  ")
}

type Property struct {
	key *datastore.Key

	Name  string
	Value string
}

func (s *Story) NewProperty(c appengine.Context, name, value string) *Property {
	p := &Property{
		key: datastore.NewKey(c, "Property", name, 0, s.key),
		Name: name,
		Value: value,
	}
	s.Meta[name] = p
	return p
}

func (p *Property) Put(c appengine.Context) os.Error {
	key, err := datastore.Put(c, p.key, p)
	if err != nil {
		return err
	}

	p.key = key

	return nil
}

func (p *Property) Get(c appengine.Context) os.Error {
	return datastore.Get(c, p.key, p)
}

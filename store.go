package main

import (
	"errors"

	"go.etcd.io/bbolt"
	"golang.org/x/exp/slog"
)

func Open(s string) *Store {
	db, err := bbolt.Open(s, 0600, bbolt.DefaultOptions)
	if err != nil {
		panic(err)
	}
	return &Store{
		db: db,
	}
}

type Store struct {
	db *bbolt.DB
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) Get(user, key string) (string, error) {
	var out []byte
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(user))
		if b == nil {
			return errors.New("not found")
		}

		out = b.Get([]byte(key))
		return nil
	})

	return string(out), err
}

func (s *Store) Set(user, key, value string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(user))
		if err != nil {
			return err
		}

		return b.Put([]byte(key), []byte(value))
	})
}

func (s *Store) Delete(user, key string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(user))
		if err != nil {
			return err
		}

		return b.Delete([]byte(key))
	})
}

func (s *Store) List(user string) (map[string]string, error) {
	out := map[string]string{}
	return out, s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(user))
		if b == nil {
			return errors.New("not found")
		}

		return b.ForEach(func(k, v []byte) error {
			slog.Debug("list", "key", string(k), "value", string(v))
			out[string(k)] = string(v)
			return nil
		})
	})

}

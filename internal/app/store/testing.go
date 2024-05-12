package store

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestStore(t *testing.T, databaseURL string) (*Store, func(...string)) {
	t.Helper()

	config := NewConfig()
	config.DatabaseURL = databaseURL
	s := New(config)
	if err := s.Open(); err != nil {
		t.Fatal(err)
	}

	return s, func(docs ...string) {
		if len(docs) > 0 {
			if _, err := s.db.Collection("users").Find(context.Background(), bson.M{"name": bson.M{"$in": docs}}); err != nil {
				t.Fatal(err)
			}
		}

		s.Close()
	}
}

func newMongoClient() *mongo.Client {
	return nil
}

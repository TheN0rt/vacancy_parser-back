package store

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Store ...
type Store struct {
	config            *Config
	client            *mongo.Client
	db                *mongo.Database
	UserRepository    *UserRepository
	VacancyRepository *VacancyRepository
}

// New ...
func New(config *Config) *Store {
	return &Store{
		config: config,
	}
}

// Open ...
func (s *Store) Open() error {
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(s.config.DatabaseURL))
	if err != nil {
		return err
	}

	err = mongoClient.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal("Ping error:", err)
	}

	db := mongoClient.Database("vacancy_parser")
	s.db = db
	fmt.Println("Connected to MongoDB!")

	return nil
}

// Close ...
func (s *Store) Close() {
	s.client.Disconnect(context.Background())
}

func (s *Store) User() *UserRepository {
	if s.UserRepository != nil {
		return s.UserRepository
	}

	s.UserRepository = &UserRepository{
		store: s,
	}

	return s.UserRepository
}

func (s *Store) Vacancy() *VacancyRepository {
	if s.VacancyRepository != nil {
		return s.VacancyRepository
	}

	s.VacancyRepository = &VacancyRepository{
		store: s,
	}

	return s.VacancyRepository
}

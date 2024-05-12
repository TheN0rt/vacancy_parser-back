package store

import (
	"context"
	"fmt"
	"vacancy-parser/internal/app/model"

	"go.mongodb.org/mongo-driver/bson"
)

type UserRepository struct {
	store *Store
}

func (r *UserRepository) CreateUser(user *model.User) (interface{}, error) {
	result, err := r.store.db.Collection("test").InsertOne(context.Background(), user)
	if err != nil {
		return nil, fmt.Errorf("cannot create user: %w", err)
	}

	return result.InsertedID, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.store.db.Collection("test").FindOne(context.Background(), bson.D{{Key: "email", Value: email}}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("cannot find user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) FindAll() ([]model.User, error) {
	var users []model.User
	cursor, err := r.store.db.Collection("test").Find(context.Background(), bson.D{})
	if err != nil {
		return nil, fmt.Errorf("cannot find users: %w", err)
	}

	if err := cursor.All(context.Background(), &users); err != nil {
		return nil, fmt.Errorf("cannot decode users: %w", err)
	}
	return users, nil
}

func (r *UserRepository) UpdateUserByEmail(userEmail string, updUser *model.User) (int64, error) {

	result, err := r.store.db.Collection("test").UpdateOne(context.Background(), bson.D{{Key: "email", Value: userEmail}}, bson.D{{Key: "$set", Value: updUser}})

	if err != nil {
		return 0, fmt.Errorf("cannot update user: %w", err)
	}

	return result.ModifiedCount, nil
}

func (r *UserRepository) DeleteUserByEmail(userEmail string) (int64, error) {

	result, err := r.store.db.Collection("test").DeleteOne(context.Background(), bson.D{{Key: "email", Value: userEmail}})

	if err != nil {
		return 0, fmt.Errorf("cannot delete user: %w", err)
	}

	return result.DeletedCount, nil
}

func (r *UserRepository) DeleteAll() (int64, error) {

	result, err := r.store.db.Collection("test").DeleteMany(context.Background(), bson.D{})

	if err != nil {
		return 0, fmt.Errorf("cannot delete users: %w", err)
	}

	return result.DeletedCount, nil
}

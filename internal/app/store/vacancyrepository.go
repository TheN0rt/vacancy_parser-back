package store

import (
	"context"
	"fmt"
	"vacancy-parser/internal/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VacancyRepository struct {
	store *Store
}

func (r *VacancyRepository) InsertVacancy(vacancy *model.Vacancy) (interface{}, error) {
	result, err := r.store.db.Collection("vacancies").InsertOne(context.Background(), vacancy)
	if err != nil {
		return nil, fmt.Errorf("cannot insert vacancy: %w", err)
	}

	return result.InsertedID, nil
}

func (r *VacancyRepository) FindAllVacancy() ([]model.Vacancy, error) {
	var vacancies []model.Vacancy
	cursor, err := r.store.db.Collection("vacancies").Find(context.Background(), bson.D{})
	if err != nil {
		return nil, fmt.Errorf("cannot find vacancies: %w", err)
	}

	if err := cursor.All(context.Background(), &vacancies); err != nil {
		return nil, fmt.Errorf("cannot decode vacancies: %w", err)
	}

	return vacancies, nil
}

func (r *VacancyRepository) FindVacancyByTitle(title string) (*model.Vacancy, error) {
	var vacancy model.Vacancy
	err := r.store.db.Collection("vacancies").FindOne(context.Background(), bson.D{{Key: "title", Value: title}}).Decode(&vacancy)
	if err != nil {
		return nil, fmt.Errorf("cannot find vacancy: %w", err)
	}
	return &vacancy, nil
}

func (r *VacancyRepository) DeleteAllVacancy() (int64, error) {
	result, err := r.store.db.Collection("vacancies").DeleteMany(context.Background(), bson.D{})
	if err != nil {
		return 0, fmt.Errorf("cannot delete vacancies: %w", err)
	}

	return result.DeletedCount, nil
}

func (r *VacancyRepository) GetVacancies(page, limit int64) ([]model.Vacancy, error) {
	var vacancies []model.Vacancy
	opts := options.Find().SetLimit(limit).SetSkip((page - 1) * limit)
	cursor, err := r.store.db.Collection("vacancies").Find(context.Background(), bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("cannot find vacancies: %w", err)
	}

	if err := cursor.All(context.Background(), &vacancies); err != nil {
		return nil, fmt.Errorf("cannot decode vacancies: %w", err)
	}

	return vacancies, nil
}

func (r *VacancyRepository) GetAllVacanciesCount() (int64, error) {
	var count int64
	count, err := r.store.db.Collection("vacancies").CountDocuments(context.Background(), bson.D{})
	if err != nil {
		return 0, fmt.Errorf("cannot count vacancies: %w", err)
	}
	return count, nil
}

func (r *VacancyRepository) GetAllHardSkills() ([]model.Vacancy, error) {
	var skills []model.Vacancy // исправить потом на { hardskills: string[] }

	cursor, err := r.store.db.Collection("vacancies").Find(context.Background(), bson.D{{Key: "hardskills", Value: bson.D{{Key: "$exists", Value: true}}}}, options.Find().SetProjection(bson.D{{Key: "hardskills", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("cannot find vacancies: %w", err)
	}
	if err := cursor.All(context.Background(), &skills); err != nil {
		return nil, fmt.Errorf("cannot decode vacancies: %w", err)
	}
	return skills, nil
}

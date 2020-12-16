package stored

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type DBEvaluation evaluate.Evaluation

type Evaluation struct {
	ID         int          `db:"id" json:"id"`
	Created    time.Time    `db:"created" json:"created"`
	Evaluation DBEvaluation `db:"val" json:"val"`
}


type Model struct {
	ID              int    `db:"id" json:"id"`
	Name            string `db:"model_name" json:"model_name"`
	IntervalsPassed int    `db:"intervals_passed" json:"intervals_passed"`
}

type Storer interface {
	GetEvaluation(model string) ([]*Evaluation, error)
	AddEvaluation(model string, evaluation *evaluate.Evaluation) error
	RemoveEvaluation(id int) error
	GetModel(model string) (*Model, error)
	UpdateModel(model string, intervalsPassed int) error
}

type LocalStore struct {
	DB *sql.DB
}

func (s *LocalStore) GetEvaluation(model string) ([]*Evaluation, error) {
	rows, err := s.DB.Query("SELECT evaluation.id, evaluation.created, evaluation.val FROM evaluation, model WHERE evaluation.model_id = model.id AND model.model_name = ?;", model)
	if err != nil {
		return nil, err
	}

	var saved []*Evaluation
	for rows.Next() {
		evaluation := Evaluation{}
		err = rows.Scan(&evaluation.ID, &evaluation.Created, &evaluation.Evaluation)
		if err != nil {
			return nil, err
		}
		saved = append(saved, &evaluation)
	}

	return saved, nil
}

func (s *LocalStore) AddEvaluation(model string, evaluation *evaluate.Evaluation) error {
	modelObj, err := s.GetModel(model)
	if err != nil {
		return err
	}

	evaluationJSON, err := json.Marshal(evaluation)
	if err != nil {
		log.Panic(err)
	}


	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO evaluation(model_id, val, created) VALUES(&evaluation.ID, &evaluation.Created, &evaluation.Evaluation);")
	if err != nil {
		return err
	}
	defer stmt.Close()
	stmt.Exec(modelObj.ID, string(evaluationJSON), time.Now().UTC().Unix())
	return tx.Commit()
}
}

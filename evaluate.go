package evaluate

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
)

type PredictiveEvaluate struct {
	HPAEvaluator hpaevaluate.Evaluater
	Store        stored.Storer
	Predicters   []prediction.Predicter
}

func (p *PredictiveEvaluate) GetEvaluation(predictiveConfig *config.Config, metrics []*metric.Metric, ) (*cpaevaluate.Evaluation, error) {
	evaluation, err := p.HPAEvaluator.GetEvaluation(metrics)
	if err != nil {
		return nil, err
	}

	predictions := []int32{evaluation.TargetReplicas}


	predicter := prediction.ModelPredict{
		Predicters: p.Predicters,
	}

	for _, model := range predictiveConfig.Models {

		dbModel, err := p.Store.GetModel(model.Name)
		if err == sql.ErrNoRows {
			err = p.Store.UpdateModel(model.Name, 1)
			if err != nil {
				return nil, err
			}
			dbModel, err = p.Store.GetModel(model.Name)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
			err = p.Store.AddEvaluation(model.Name, evaluation)
			if err != nil {
				return nil, err
			}
		}

		saved, err := p.Store.GetEvaluation(model.Name)
		if err != nil {
			return nil, err
		}
		prediction, err := predicter.GetPrediction(model, saved)
		if err != nil {
			return nil, err
		}
		predictions = append(predictions, prediction)

	targetPrediction := evaluation.TargetReplicas
	switch predictiveConfig.DecisionType {
	case config.DecisionMaximum:
		max := int32(0)
		for i, prediction := range predictions {
			if i == 0 || prediction > max {
				max = prediction
			}
		}
		targetPrediction = max
		break
	case config.DecisionMinimum:
		min := int32(0)
		for i, prediction := range predictions {
			if i == 0 || prediction < min {
				min = prediction
			}
		}
		targetPrediction = min
		break
	case config.DecisionMean:
		total := int32(0)
		for _, prediction := range predictions {
			total += prediction
		}
		targetPrediction = int32(math.Ceil(float64(int(total) / len(predictions))))
		break
	case config.DecisionMedian:
		halfIndex := len(predictions) / 2
		if len(predictions)%2 == 0 {
			// Even
			targetPrediction = (predictions[halfIndex-1] + predictions[halfIndex]) / 2
		} else {
			// Odd
			targetPrediction = predictions[halfIndex]
		}
		break
	default:
		return nil, fmt.Errorf("Unknown decision type '%s'", predictiveConfig.DecisionType)
	}

	return &cpaevaluate.Evaluation{
		TargetReplicas: targetPrediction,
	}, nil
}

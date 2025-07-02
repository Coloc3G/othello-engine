package learning

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func (t *Trainer) createModelDirectory() error {
	// Create a directory for models if it doesn't exist
	if _, err := os.Stat("training"); os.IsNotExist(err) {
		return os.Mkdir("training", 0755)
	}

	// Create a subdirectory for the specific model
	subdir := fmt.Sprintf("training/%s", t.Name)
	if _, err := os.Stat(subdir); os.IsNotExist(err) {
		return os.Mkdir(subdir, 0755)
	}

	return nil
}

// SaveModel saves a model to a JSON file
func (t *Trainer) SaveModel(filename string, model EvaluationModel) error {
	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return err
	}
	filePath := fmt.Sprintf("training/%s/%s", t.Name, filename)
	return os.WriteFile(filePath, data, 0644)
}

// LoadModel loads a model from a JSON file
func (t *Trainer) LoadModel(filename string) (EvaluationModel, error) {
	var model EvaluationModel
	data, err := os.ReadFile(filename)
	if err != nil {
		return model, err
	}
	err = json.Unmarshal(data, &model)
	return model, err
}

// SaveModelToFile is a generic helper method to save structs to JSON files
func (t *Trainer) SaveModelToFile(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	filePath := fmt.Sprintf("training/%s/%s", t.Name, filename)
	return os.WriteFile(filePath, jsonData, 0644)
}

// SaveGenerationStats saves statistics about the current generation
func (t *Trainer) SaveGenerationStats(gen int) error {
	stats := struct {
		Generation  int             `json:"generation"`
		BestFitness float64         `json:"best_fitness"`
		AvgFitness  float64         `json:"avg_fitness"`
		BestModel   EvaluationModel `json:"best_model"`
		Timestamp   string          `json:"timestamp"`
	}{
		Generation:  gen,
		BestFitness: t.Models[0].Fitness,
		BestModel:   t.Models[0],
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	// Calculate average fitness
	var sum float64
	for _, model := range t.Models {
		sum += model.Fitness
	}
	stats.AvgFitness = sum / float64(len(t.Models))

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("training/%s/stats_gen_%d.json", t.Name, gen)
	return os.WriteFile(filename, data, 0644)
}

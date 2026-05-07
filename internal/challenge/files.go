package challenge

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadNormalization() (Normalization, error) {
	var normalization Normalization

	if err := loadJSONFile(NormalizationJSONPath, &normalization); err != nil {
		return Normalization{}, fmt.Errorf("load normalization: %w", err)
	}

	return normalization, nil
}

func LoadMCCRisk() (map[string]float64, error) {
	var riskByMCC map[string]float64

	if err := loadJSONFile(MCCRiskJSONPath, &riskByMCC); err != nil {
		return nil, fmt.Errorf("load mcc risk: %w", err)
	}

	return riskByMCC, nil
}

func loadJSONFile(path string, out any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(out)
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"rinha-backend-2026-go/internal/challenge"
)

// application guarda somente os dados necessários para vetorizar a transação.
// A API não carrega references.json.gz e não faz busca vetorial local,
// para evitar duplicar o dataset nas duas instâncias obrigatórias.
type application struct {
	normalization NormalizationData
	mccRisk       map[string]float64
	client        *http.Client
}

type NormalizationData = challenge.Normalization

func main() {
	normalization, err := challenge.LoadNormalization()
	if err != nil {
		fail(err)
	}

	mccRisk, err := challenge.LoadMCCRisk()
	if err != nil {
		fail(err)
	}

	app := &application{
		normalization: normalization,
		mccRisk:       mccRisk,
		client:        http.DefaultClient,
	}

	mux := http.NewServeMux()
	mux.HandleFunc(challenge.ReadyPath, app.handleReady)
	mux.HandleFunc(challenge.FraudScorePath, app.handleFraudScore)

	if err := http.ListenAndServe(challenge.APIServerAddress, mux); err != nil {
		fail(err)
	}
}

func (app *application) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	response, err := app.client.Get(challenge.VectorEngineInternalURL + challenge.ReadyPath)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) handleFraudScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request challenge.FraudScoreRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vector := challenge.Vectorize(request, app.normalization, app.mccRisk)
	fraudCount, err := app.searchFraudCount(vector)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fraudScore := float64(fraudCount) / float64(challenge.NeighborsK)

	response := challenge.FraudScoreResponse{
		Approved:   fraudScore < challenge.FraudThreshold,
		FraudScore: fraudScore,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (app *application) searchFraudCount(vector [challenge.VectorDimensions]float32) (int, error) {
	rawVector := challenge.EncodeVector(vector)

	response, err := app.client.Post(
		challenge.VectorEngineInternalURL+challenge.VectorSearchPath,
		"application/octet-stream",
		bytes.NewReader(rawVector[:]),
	)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("vector engine status: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}

	if len(body) != 1 {
		return 0, fmt.Errorf("invalid vector engine response size")
	}

	return int(body[0]), nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"unsafe"

	"rinha-backend-2026-go/internal/challenge"
)

// vectorStore mantém o dataset oficial já pré-processado em memória.
//
// A busca é brute force exata sobre os 14 campos:
// - percorre todos os vetores;
// - calcula distância euclidiana quadrática;
// - mantém os 5 menores resultados;
// - retorna apenas quantos dos 5 são fraude.
//
// Não existe regra de fraude aqui. O serviço só calcula vizinhança.
// A decisão final fica na API, seguindo o enunciado.
type vectorStore struct {
	rawVectors []byte
	vectors    []float32
	labels     []byte
	count      int
}

func main() {
	store, err := loadVectorStore()
	if err != nil {
		fail(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc(challenge.ReadyPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc(challenge.VectorSearchPath, store.handleSearch)

	if err := http.ListenAndServe(challenge.VectorEngineServerAdress, mux); err != nil {
		fail(err)
	}
}

func loadVectorStore() (*vectorStore, error) {
	rawVectors, err := os.ReadFile(challenge.ProcessedVectorsPath)
	if err != nil {
		return nil, fmt.Errorf("read processed vectors: %w", err)
	}

	labels, err := os.ReadFile(challenge.ProcessedLabelsPath)
	if err != nil {
		return nil, fmt.Errorf("read processed labels: %w", err)
	}

	if len(rawVectors)%challenge.VectorByteSize != 0 {
		return nil, fmt.Errorf("invalid vector file size")
	}

	count := len(rawVectors) / challenge.VectorByteSize
	if count != len(labels) {
		return nil, fmt.Errorf("vectors count and labels count mismatch")
	}

	vectors := unsafe.Slice((*float32)(unsafe.Pointer(&rawVectors[0])), len(rawVectors)/4)

	return &vectorStore{
		rawVectors: rawVectors,
		vectors:    vectors,
		labels:     labels,
		count:      count,
	}, nil
}

func (s *vectorStore) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var rawQuery [challenge.VectorByteSize]byte

	if _, err := io.ReadFull(r.Body, rawQuery[:]); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	query := challenge.DecodeVector(rawQuery[:])
	fraudCount := s.searchFraudCount(query)

	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write([]byte{byte(fraudCount)})
}

func (s *vectorStore) searchFraudCount(query [challenge.VectorDimensions]float32) int {
	var bestDistances [challenge.NeighborsK]float32
	var bestLabels [challenge.NeighborsK]byte

	for i := 0; i < challenge.NeighborsK; i++ {
		bestDistances[i] = float32(math.MaxFloat32)
	}

	for recordIndex := 0; recordIndex < s.count; recordIndex++ {
		base := recordIndex * challenge.VectorDimensions

		var distance float32

		for dimension := 0; dimension < challenge.VectorDimensions; dimension++ {
			diff := s.vectors[base+dimension] - query[dimension]
			distance += diff * diff

			// Como só precisamos dos 5 menores, se a distância parcial já passou
			// do pior candidato atual, este registro não pode entrar no top 5.
			if distance >= bestDistances[challenge.NeighborsK-1] {
				break
			}
		}

		if distance < bestDistances[challenge.NeighborsK-1] {
			insertCandidate(distance, s.labels[recordIndex], &bestDistances, &bestLabels)
		}
	}

	fraudCount := 0

	for _, label := range bestLabels {
		if label == challenge.LabelFraud {
			fraudCount++
		}
	}

	return fraudCount
}

func insertCandidate(
	distance float32,
	label byte,
	bestDistances *[challenge.NeighborsK]float32,
	bestLabels *[challenge.NeighborsK]byte,
) {
	position := challenge.NeighborsK - 1

	for position > 0 && distance < bestDistances[position-1] {
		bestDistances[position] = bestDistances[position-1]
		bestLabels[position] = bestLabels[position-1]
		position--
	}

	bestDistances[position] = distance
	bestLabels[position] = label
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

package main

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"

	"rinha-backend-2026-go/internal/challenge"
)

// O desafio permite pré-processar os arquivos porque eles não mudam durante o teste.
// Este binário converte references.json.gz para dois arquivos simples:
//
// /app/processed/references.f32
//   - todos os vetores em float32 little-endian
//   - 14 float32 por registro
//
// /app/processed/labels.bin
//   - 1 byte por registro
//   - 0 = legit
//   - 1 = fraud
//
// Isso evita carregar/parsing de JSON no runtime do vector-engine.
func main() {
	if err := os.MkdirAll(challenge.ProcessedDir, 0o755); err != nil {
		fail(err)
	}

	source, err := os.Open(challenge.ReferencesJSONGZPath)
	if err != nil {
		fail(fmt.Errorf("open references json gz: %w", err))
	}
	defer source.Close()

	gzipReader, err := gzip.NewReader(source)
	if err != nil {
		fail(fmt.Errorf("create gzip reader: %w", err))
	}
	defer gzipReader.Close()

	vectorFile, err := os.Create(challenge.ProcessedVectorsPath)
	if err != nil {
		fail(fmt.Errorf("create processed vectors file: %w", err))
	}
	defer vectorFile.Close()

	labelFile, err := os.Create(challenge.ProcessedLabelsPath)
	if err != nil {
		fail(fmt.Errorf("create processed labels file: %w", err))
	}
	defer labelFile.Close()

	vectorWriter := bufio.NewWriterSize(vectorFile, 1024*1024)
	defer vectorWriter.Flush()

	labelWriter := bufio.NewWriterSize(labelFile, 1024*1024)
	defer labelWriter.Flush()

	decoder := json.NewDecoder(gzipReader)

	token, err := decoder.Token()
	if err != nil {
		fail(fmt.Errorf("read json start token: %w", err))
	}

	if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
		fail(fmt.Errorf("references json must start with array"))
	}

	var count int

	for decoder.More() {
		var record challenge.ReferenceRecord

		if err := decoder.Decode(&record); err != nil {
			fail(fmt.Errorf("decode reference record %d: %w", count, err))
		}

		writeVector(vectorWriter, record.Vector)
		writeLabel(labelWriter, record.Label)

		count++
	}

	token, err = decoder.Token()
	if err != nil {
		fail(fmt.Errorf("read json end token: %w", err))
	}

	if delimiter, ok := token.(json.Delim); !ok || delimiter != ']' {
		fail(fmt.Errorf("references json must end with array"))
	}

	fmt.Printf("preprocessed references: %d\n", count)
}

func writeVector(writer *bufio.Writer, vector [challenge.VectorDimensions]float64) {
	var raw [challenge.VectorByteSize]byte

	for i, value := range vector {
		offset := i * 4
		binary.LittleEndian.PutUint32(raw[offset:offset+4], math.Float32bits(float32(value)))
	}

	if _, err := writer.Write(raw[:]); err != nil {
		fail(fmt.Errorf("write processed vector: %w", err))
	}
}

func writeLabel(writer *bufio.Writer, label string) {
	value := challenge.LabelLegit

	if label == "fraud" {
		value = challenge.LabelFraud
	}

	if err := writer.WriteByte(value); err != nil {
		fail(fmt.Errorf("write processed label: %w", err))
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

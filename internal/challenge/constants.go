package challenge

const (
	// O enunciado define exatamente 14 dimensões para o vetor.
	VectorDimensions = 14

	// O enunciado define busca dos 5 vizinhos mais próximos.
	NeighborsK = 5

	// O enunciado define approved = fraud_score < 0.6.
	FraudThreshold = 0.6

	// O enunciado define -1 como sentinela quando last_transaction é null.
	MissingLastTransactionValue = -1.0

	// O enunciado define 0.5 como risco padrão para MCC ausente no mcc_risk.json.
	DefaultMCCRisk = 0.5

	// Arquivos oficiais fornecidos pelo desafio.
	ResourceDir              = "/app/resources"
	ProcessedDir             = "/app/processed"
	ReferencesJSONGZPath     = ResourceDir + "/references.json.gz"
	MCCRiskJSONPath          = ResourceDir + "/mcc_risk.json"
	NormalizationJSONPath    = ResourceDir + "/normalization.json"
	ProcessedVectorsPath     = ProcessedDir + "/references.f32"
	ProcessedLabelsPath      = ProcessedDir + "/labels.bin"
	VectorEngineInternalURL  = "http://vector-engine:8080"
	ReadyPath                = "/ready"
	FraudScorePath           = "/fraud-score"
	VectorSearchPath         = "/search"
	APIServerAddress         = ":8080"
	VectorEngineServerAdress = ":8080"

	// Cada float32 ocupa 4 bytes.
	// O vetor interno enviado da API para o vector-engine possui 14 float32.
	VectorByteSize = VectorDimensions * 4
)

const (
	LabelLegit byte = 0
	LabelFraud byte = 1
)

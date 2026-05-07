package challenge

import "time"

// Normalization representa o conteúdo oficial de normalization.json.
type Normalization struct {
	MaxAmount            float64 `json:"max_amount"`
	MaxInstallments      float64 `json:"max_installments"`
	AmountVsAvgRatio    float64 `json:"amount_vs_avg_ratio"`
	MaxMinutes           float64 `json:"max_minutes"`
	MaxKM                float64 `json:"max_km"`
	MaxTxCount24h        float64 `json:"max_tx_count_24h"`
	MaxMerchantAvgAmount float64 `json:"max_merchant_avg_amount"`
}

// FraudScoreRequest representa exatamente o payload recebido em POST /fraud-score.
type FraudScoreRequest struct {
	ID              string               `json:"id"`
	Transaction     Transaction          `json:"transaction"`
	Customer        Customer             `json:"customer"`
	Merchant        Merchant             `json:"merchant"`
	Terminal        Terminal             `json:"terminal"`
	LastTransaction *LastTransaction      `json:"last_transaction"`
}

type Transaction struct {
	Amount      float64   `json:"amount"`
	Installments int      `json:"installments"`
	RequestedAt time.Time `json:"requested_at"`
}

type Customer struct {
	AvgAmount      float64  `json:"avg_amount"`
	TxCount24h     int      `json:"tx_count_24h"`
	KnownMerchants []string `json:"known_merchants"`
}

type Merchant struct {
	ID        string  `json:"id"`
	MCC       string  `json:"mcc"`
	AvgAmount float64 `json:"avg_amount"`
}

type Terminal struct {
	IsOnline    bool    `json:"is_online"`
	CardPresent bool   `json:"card_present"`
	KMFromHome  float64 `json:"km_from_home"`
}

type LastTransaction struct {
	Timestamp     time.Time `json:"timestamp"`
	KMFromCurrent float64   `json:"km_from_current"`
}

// FraudScoreResponse representa exatamente a resposta esperada pelo desafio.
type FraudScoreResponse struct {
	Approved   bool    `json:"approved"`
	FraudScore float64 `json:"fraud_score"`
}

// ReferenceRecord representa cada item de references.json.gz.
type ReferenceRecord struct {
	Vector [VectorDimensions]float64 `json:"vector"`
	Label  string                    `json:"label"`
}

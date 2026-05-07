package challenge

import (
	"encoding/binary"
	"math"
	"time"
)

// Vectorize transforma o payload recebido no vetor oficial de 14 dimensões.
//
// Ordem exigida pelo enunciado:
// 0  amount
// 1  installments
// 2  amount_vs_avg
// 3  hour_of_day
// 4  day_of_week
// 5  minutes_since_last_tx
// 6  km_from_last_tx
// 7  km_from_home
// 8  tx_count_24h
// 9  is_online
// 10 card_present
// 11 unknown_merchant
// 12 mcc_risk
// 13 merchant_avg_amount
func Vectorize(
	request FraudScoreRequest,
	normalization Normalization,
	mccRisk map[string]float64,
) [VectorDimensions]float32 {
	requestedAt := request.Transaction.RequestedAt.UTC()

	var vector [VectorDimensions]float32

	vector[0] = f32(clamp01(request.Transaction.Amount / normalization.MaxAmount))
	vector[1] = f32(clamp01(float64(request.Transaction.Installments) / normalization.MaxInstallments))
	vector[2] = f32(amountVsAverage(request.Transaction.Amount, request.Customer.AvgAmount, normalization.AmountVsAvgRatio))
	vector[3] = f32(float64(requestedAt.Hour()) / 23.0)
	vector[4] = f32(float64(mondayBasedWeekday(requestedAt)) / 6.0)

	if request.LastTransaction == nil {
		vector[5] = MissingLastTransactionValue
		vector[6] = MissingLastTransactionValue
	} else {
		minutes := request.Transaction.RequestedAt.Sub(request.LastTransaction.Timestamp).Minutes()
		vector[5] = f32(clamp01(minutes / normalization.MaxMinutes))
		vector[6] = f32(clamp01(request.LastTransaction.KMFromCurrent / normalization.MaxKM))
	}

	vector[7] = f32(clamp01(request.Terminal.KMFromHome / normalization.MaxKM))
	vector[8] = f32(clamp01(float64(request.Customer.TxCount24h) / normalization.MaxTxCount24h))
	vector[9] = boolToF32(request.Terminal.IsOnline)
	vector[10] = boolToF32(request.Terminal.CardPresent)
	vector[11] = boolToF32(!contains(request.Customer.KnownMerchants, request.Merchant.ID))
	vector[12] = f32(resolveMCCRisk(request.Merchant.MCC, mccRisk))
	vector[13] = f32(clamp01(request.Merchant.AvgAmount / normalization.MaxMerchantAvgAmount))

	return vector
}

// EncodeVector converte o vetor da API em um payload binário pequeno para o serviço interno de busca.
// O cliente externo nunca vê esse formato; é apenas comunicação interna entre api e vector-engine.
func EncodeVector(vector [VectorDimensions]float32) [VectorByteSize]byte {
	var out [VectorByteSize]byte

	for i, value := range vector {
		offset := i * 4
		binary.LittleEndian.PutUint32(out[offset:offset+4], math.Float32bits(value))
	}

	return out
}

// DecodeVector faz o caminho inverso no vector-engine.
func DecodeVector(raw []byte) [VectorDimensions]float32 {
	var vector [VectorDimensions]float32

	for i := 0; i < VectorDimensions; i++ {
		offset := i * 4
		vector[i] = math.Float32frombits(binary.LittleEndian.Uint32(raw[offset : offset+4]))
	}

	return vector
}

func amountVsAverage(amount, avgAmount, ratio float64) float64 {
	if avgAmount <= 0 || ratio <= 0 {
		return 1.0
	}

	return clamp01((amount / avgAmount) / ratio)
}

func resolveMCCRisk(mcc string, riskByMCC map[string]float64) float64 {
	risk, found := riskByMCC[mcc]
	if !found {
		return DefaultMCCRisk
	}

	return risk
}

func mondayBasedWeekday(t time.Time) int {
	weekday := int(t.UTC().Weekday())

	if weekday == 0 {
		return 6
	}

	return weekday - 1
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}

	if value > 1 {
		return 1
	}

	return value
}

func boolToF32(value bool) float32 {
	if value {
		return 1
	}

	return 0
}

func f32(value float64) float32 {
	return float32(value)
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}

	return false
}

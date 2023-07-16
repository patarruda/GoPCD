package base

import "strings"

type ConversorMedidas struct{}

func (ConversorMedidas) Invoke(reqMedidas RequestConversor) RequestConversor {
	valor := reqMedidas.Valor

	switch strings.ToUpper(reqMedidas.FromUnit + reqMedidas.ToUnit) {
	case "CF":
		reqMedidas.Valor = ConversorMedidas{}.CelsiusToFahrenheit(valor)
	case "FC":
		reqMedidas.Valor = ConversorMedidas{}.FahrenheitToCelsius(valor)
	case "MFT":
		reqMedidas.Valor = ConversorMedidas{}.metersToFeet(valor)
	case "FTM":
		reqMedidas.Valor = ConversorMedidas{}.feetToMeters(valor)
	case "KGLB":
		reqMedidas.Valor = ConversorMedidas{}.kilogramsToPounds(valor)
	case "LBKG":
		reqMedidas.Valor = ConversorMedidas{}.poundsToKilograms(valor)
	default:
		reqMedidas.Valor = 0
		reqMedidas.ToUnit = "Não Suportado"
	}
	return reqMedidas
}

// Temperatura
func (ConversorMedidas) CelsiusToFahrenheit(medida float64) float64 {
	return (medida * 9 / 5) + 32
}

func (ConversorMedidas) FahrenheitToCelsius(medida float64) float64 {
	return (medida - 32) * 5 / 9
}

// Distância
func (ConversorMedidas) metersToFeet(medida float64) float64 {
	return medida * 3.28084
}

func (ConversorMedidas) feetToMeters(medida float64) float64 {
	return medida / 3.28084
}

// Peso
func (ConversorMedidas) kilogramsToPounds(medida float64) float64 {
	return medida * 2.20462
}

func (ConversorMedidas) poundsToKilograms(medida float64) float64 {
	return medida / 2.20462
}

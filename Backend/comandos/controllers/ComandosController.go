package controllers

import (
	"Proyecto/comandos/general"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleCommand(w http.ResponseWriter, r *http.Request) {

	// =========================
	// CORS - Preflight
	// =========================
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	// =========================
	// Validar método
	// =========================
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// =========================
	// Headers
	// =========================
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// =========================
	// Body esperado
	// =========================
	var requestBody struct {
		Comandos *string `json:"Comandos"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&requestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			general.ResultadoSalida("JSON inválido o campos no permitidos", true, nil),
		)
		return
	}

	if requestBody.Comandos == nil || strings.TrimSpace(*requestBody.Comandos) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			general.ResultadoSalida("El campo 'Comandos' es obligatorio", true, nil),
		)
		return
	}

	// =========================
	// Ejecución de comandos
	// =========================
	comandos := strings.Split(*requestBody.Comandos, "\n")
	resultado := general.ExecuteCommandList(comandos)

	// =========================
	// Ejecutar comandos reales y obtener logs
	// =========================
	_, contadorErrores, logs := general.GlobalCom(resultado.Salida.LstComandos)

	// LOG EN CONSOLA
	for _, r := range logs {
		fmt.Println(r)
	}

	// =========================
	// Verificación REAL de errores
	// =========================
	hayError := contadorErrores > 0

	// Fallback: detectar errores por contenido
	if !hayError {
		for _, r := range logs {
			if strings.HasPrefix(strings.TrimSpace(r), "❌") || strings.HasPrefix(strings.TrimSpace(r), "[ERROR]") {
				hayError = true
				break
			}
		}
	}

	// =========================
	// Respuesta HTTP
	// =========================
	status := http.StatusOK
	message := "Comandos ejecutados correctamente"

	if hayError {
		status = http.StatusBadRequest
		message = "Ocurrieron errores al ejecutar los comandos"
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(
		general.ResultadoSalida(
			message,
			hayError,
			logs, // Enviamos logs legibles al frontend
		),
	)
}

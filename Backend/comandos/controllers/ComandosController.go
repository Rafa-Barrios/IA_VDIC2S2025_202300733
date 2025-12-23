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
	// SALIDA CORRECTA (NO Data, NO Respuesta)
	// =========================
	salida := resultado.Salida

	// Ejecutar comandos reales (FDISK, MKDISK, etc.)
	errores, contadorErrores := general.GlobalCom(salida.LstComandos)
	fmt.Println(errores, contadorErrores)

	// =========================
	// Respuesta HTTP
	// =========================
	if contadorErrores > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			general.ResultadoSalida(
				"Ocurrieron errores al ejecutar los comandos",
				true,
				errores,
			),
		)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		general.ResultadoSalida(
			"Comandos ejecutados correctamente",
			false,
			salida.LstComandos,
		),
	)
}

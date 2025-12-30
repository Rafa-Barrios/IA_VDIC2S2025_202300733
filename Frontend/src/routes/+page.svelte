<script lang="ts">
    let inputCommands = "";
    let outputCommands = "";
    let selectedFile: File | null = null;
    let loading = false;

    function handleFileChange(event: Event) {
        const input = event.target as HTMLInputElement;
        if (input.files && input.files.length > 0) {
            selectedFile = input.files[0];
        }
    }

    function loadFile() {
        if (!selectedFile) {
            alert("Seleccione un archivo primero");
            return;
        }

        const reader = new FileReader();
        reader.onload = (e: ProgressEvent<FileReader>) => {
            if (e.target && typeof e.target.result === "string") {
                inputCommands = e.target.result;
            }
        };
        reader.readAsText(selectedFile);
    }

    async function executeCommands() {
        if (!inputCommands.trim()) {
            alert("No hay comandos para ejecutar");
            return;
        }

        loading = true;
        outputCommands = "";

        try {
            const response = await fetch("http://localhost:9700/commands", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ Comandos: inputCommands })
            });

            const data = await response.json();

            if (Array.isArray(data.data)) {
                outputCommands = data.data.join("\n");
            } else {
                outputCommands = data.message;
            }

        } catch {
            outputCommands = "Error al conectar con el servidor";
        } finally {
            loading = false;
        }
    }

    // limpiar todo
    function clearAll() {
        inputCommands = "";
        outputCommands = "";
        selectedFile = null;

        const fileInput = document.getElementById("fileInput") as HTMLInputElement;
        if (fileInput) fileInput.value = "";
    }
</script>

<style>
    .container {
        padding: 30px;
        font-family: Arial, sans-serif;
        max-width: 1400px;
        margin: auto;
    }

    h1 {
        text-align: center;
        font-size: 36px;
        font-weight: bold;
        margin-bottom: 25px;
        color: #1e293b;
    }

    .panels {
        display: flex;
        gap: 40px;
    }

    .panel {
        flex: 1;
        display: flex;
        flex-direction: column;
    }

    .panel label {
        font-weight: bold;
        margin-bottom: 8px;
    }

    textarea {
        width: 100%;
        height: 350px;
        resize: none;
        padding: 12px;
        font-family: monospace;
        font-size: 14px;
        border: 1px solid #cbd5f5;
        border-radius: 8px;
    }

    .controls {
        display: flex;
        gap: 10px;
        align-items: center;
        margin-bottom: 10px;
    }

    button, .file-label {
        padding: 9px 16px;
        border: none;
        border-radius: 8px;
        background-color: #2563eb;
        color: white;
        font-weight: bold;
        cursor: pointer;
        text-align: center;
    }

    button:hover, .file-label:hover {
        opacity: 0.9;
    }

    button:disabled {
        background-color: #94a3b8;
        cursor: not-allowed;
    }

    .file-input {
        display: none;
    }

    .file-name {
        font-size: 13px;
        color: #475569;
    }

</style>

<div class="container">
    <h1>GoDisk</h1>

    <div class="panels">
        <!-- ENTRADA -->
        <div class="panel">
            <label>Entrada de Comandos</label>

            <div class="controls">
                <input
                    id="fileInput"
                    type="file"
                    class="file-input"
                    on:change={handleFileChange}
                />

                <label for="fileInput" class="file-label">
                    Elegir archivo
                </label>

                <button type="button" on:click={loadFile}>
                    Cargar
                </button>

                <span class="file-name">
                    {selectedFile ? selectedFile.name : "No se eligió ningún archivo"}
                </span>
            </div>

            <textarea
                bind:value={inputCommands}
                placeholder="Ingrese comandos aquí o cargue un archivo..."
            ></textarea>
        </div>

        <!-- SALIDA -->
        <div class="panel">
            <label>Salida de Comandos</label>

            <div class="controls">
                <button on:click={executeCommands} disabled={loading}>
                    {loading ? "Ejecutando..." : "Ejecutar"}
                </button>

                <!-- 🔹 NUEVO BOTÓN -->
                <button on:click={clearAll}>
                    Limpiar
                </button>
            </div>

            <textarea
                bind:value={outputCommands}
                readonly
                placeholder="Aquí se mostrará la salida del backend..."
            ></textarea>
        </div>
    </div>
</div>

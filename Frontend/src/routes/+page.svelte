<script lang="ts">
    let inputCommands = "";
    let outputCommands = "";
    let selectedFile: File | null = null;
    let loading = false;

    // Manejar selección de archivo
    function handleFileChange(event: Event) {
        const input = event.target as HTMLInputElement;

        if (input.files && input.files.length > 0) {
            selectedFile = input.files[0];
        }
    }

    // Cargar archivo al textarea de entrada
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

    // Ejecutar comandos
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
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    Comandos: inputCommands
                })
            });

            const data = await response.json();

            // Ajusta esto si tu backend devuelve otro formato
            if (Array.isArray(data)) {
                outputCommands = data.join("\n");
            } else if (data.mensaje) {
                outputCommands = data.mensaje;
            } else {
                outputCommands = JSON.stringify(data, null, 2);
            }

        } catch (error) {
            outputCommands = "Error al conectar con el servidor:\n" + error;
        } finally {
            loading = false;
        }
    }
</script>

<style>
    .container {
        padding: 20px;
        font-family: Arial, sans-serif;
    }

    h1 {
        text-align: center;
        color: #1e40af;
        margin-bottom: 20px;
    }

    .panels {
        display: flex;
        gap: 20px;
    }

    textarea {
        width: 100%;
        height: 350px;
        resize: none;
        padding: 10px;
        font-family: monospace;
        font-size: 14px;
        border: 1px solid #ccc;
        border-radius: 6px;
    }

    .panel {
        flex: 1;
        display: flex;
        flex-direction: column;
    }

    .panel label {
        font-weight: bold;
        margin-bottom: 5px;
    }

    .buttons {
        margin-top: 15px;
        display: flex;
        gap: 10px;
    }

    button {
        padding: 8px 14px;
        border: none;
        border-radius: 6px;
        cursor: pointer;
        font-weight: bold;
    }

    button:hover {
        opacity: 0.9;
    }

    .btn-file {
        background-color: #64748b;
        color: white;
    }

    .btn-load {
        background-color: #0ea5e9;
        color: white;
    }

    .btn-run {
        background-color: #22c55e;
        color: white;
    }

    button:disabled {
        background-color: #9ca3af;
        cursor: not-allowed;
    }
</style>

<div class="container">
    <h1>GoDisk</h1>

    <div class="panels">
        <!-- Entrada -->
        <div class="panel">
            <label>Entrada de Comandos</label>
            <textarea bind:value={inputCommands}
                placeholder="Ingrese comandos aquí o cargue un archivo..."></textarea>
        </div>

        <!-- Salida -->
        <div class="panel">
            <label>Salida de Comandos</label>
            <textarea bind:value={outputCommands} readonly
                placeholder="Aquí se mostrará la salida del backend..."></textarea>
        </div>
    </div>

    <div class="buttons">
        <input type="file" on:change={handleFileChange} />
        <button class="btn-load" on:click={loadFile}>Cargar</button>
        <button class="btn-run" on:click={executeCommands} disabled={loading}>
            {loading ? "Ejecutando..." : "Ejecutar"}
        </button>
    </div>
</div>

package disk

import (
	"fmt"

	"github.com/fatih/color"
)

// mountedExecute muestra todas las particiones montadas
func mountedExecute(_ string, _ map[string]string) (string, bool) {

	color.Green("-----------------------------------------------------------")
	color.Blue("Particiones montadas en el sistema")
	color.Green("-----------------------------------------------------------")

	if len(mountedPartitions) == 0 {
		color.Yellow("No hay particiones montadas actualmente")
		return "No hay particiones montadas", false
	}

	for _, part := range mountedPartitions {
		color.Cyan("• %s", part.Id)
	}

	color.Green("-----------------------------------------------------------")
	return fmt.Sprintf("Total de particiones montadas: %d", len(mountedPartitions)), false
}

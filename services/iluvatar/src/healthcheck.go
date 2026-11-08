package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// runHealthcheck atiende `iluvatar healthcheck`: consulta /health del propio proceso y devuelve el código
// de salida. La imagen de producción no trae shell ni curl, así que el HEALTHCHECK del contenedor usa esto.
func runHealthcheck(port string) int {
	if err := checkHealth("http://127.0.0.1:"+port+"/health", &http.Client{Timeout: 3 * time.Second}); err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck failed:", err)
		return 1
	}
	return 0
}

func checkHealth(url string, client *http.Client) error {
	response, err := client.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", response.StatusCode)
	}
	return nil
}

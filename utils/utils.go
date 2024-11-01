package utils

import (
	"fmt"
	"golang_tui/config"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

type ServersIP []string

// PingServer realiza un ping a una IP específica y envía el resultado a un canal.
func pingServer(ip string, wg *sync.WaitGroup, results chan<- string) {
	defer wg.Done()

	var cmd *exec.Cmd

	// Ajustar el comando según el sistema operativo
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ping", "-n", "1", ip) // Windows usa -n para el número de paquetes
	case "linux", "darwin":
		cmd = exec.Command("ping", "-c", "1", ip) // Linux y macOS usan -c para el número de paquetes
	default:
		results <- fmt.Sprintf("Sistema operativo no soportado para %s", ip)
		return
	}

	// Ejecuta el comando ping en la IP especificada
	out, err := cmd.Output()
	if err != nil {
		results <- fmt.Sprintf("Servidor %s: no está corriendo", ip)
		return
	}

	// Analiza la salida para verificar si el ping fue exitoso
	if strings.Contains(string(out), "1 received") || strings.Contains(string(out), "TTL=") {
		results <- fmt.Sprintf("Servidor %s: corriendo", ip)
	} else {
		results <- fmt.Sprintf("Servidor %s: no está corriendo", ip)
	}
}

// PingServers ejecuta pings a todas las IPs en paralelo y muestra los resultados.
// PingServers ejecuta pings a todas las IPs en paralelo y devuelve el canal de resultados.
func PingServers() <-chan string {
	cfg := config.LoadConfig()
	serversIP := ServersIP{
		cfg.SBS_PUPPET,
		cfg.SBS_INTERFACE,
		cfg.SBS_CORE,
		cfg.SBS_PLATFORM_API,
		cfg.SBS_PLATFORM_CORE,
		cfg.SBS_PLATFORM_DB,
		cfg.SBS_METRICS,
		cfg.SBS_STAGING,
		cfg.SBS_OPC,
		cfg.FLRApp,
		cfg.FLR_DB,
		cfg.FLR_METRICS,
		cfg.FLR_OPC,
		cfg.FLR_FM,
	}

	var wg sync.WaitGroup
	results := make(chan string, len(serversIP))

	// Inicia una goroutine para cada IP en la lista
	for _, ip := range serversIP {
		wg.Add(1)
		go pingServer(ip, &wg, results)
	}

	// Cierra el canal una vez que todas las goroutines hayan terminado
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func ClearConsole() {
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

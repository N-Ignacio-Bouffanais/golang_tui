// utils.go
package utils

import (
	"fmt"
	"golang_tui/config"
	"os"
	"os/exec"
	"runtime"
)

type ServersIP []string

// PingServer realiza un mapeo de la lista de servidores y manda un ping a cada uno, retorna la respuesta de cada uno.
func pingServer(ip string) {

	var cmd *exec.Cmd

	// Ajustar el comando según el sistema operativo
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ping", "-n", "4", ip) // Windows usa -n para el número de paquetes
	case "linux", "darwin":
		cmd = exec.Command("ping", "-c", "4", ip) // Linux y macOS usan -c para el número de paquetes
	default:
		fmt.Printf("Sistema operativo no soportado\n")
		return
	}

	// Ejecuta el comando ping en la IP especificada
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error al hacer ping a %s: %v\n", ip, err)
		return
	}
	fmt.Printf("Respuesta de %s:\n%s\n", ip, string(out))
}

func PingServers() {
	cfg := config.LoadConfig()
	serversIP := ServersIP{
		cfg.SBS_PUPPET,    // SBS_PUPPET
		cfg.SBS_INTERFACE, // SBS_INTERFACE
		cfg.SBS_CORE,      // SBS_CORE
		cfg.SBS_BRIGDE,    // SBS_BRIGDE
		cfg.FLRApp,
		cfg.FLR_DB,
		cfg.FLR_METRICS,
		cfg.FLR_OPC,
	}
	for _, ip := range serversIP {
		pingServer(ip)
	}
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

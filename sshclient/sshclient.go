package sshclient

import (
	"fmt"
	"golang_tui/config"
	"time"

	"golang.org/x/crypto/ssh"
)

func ExecuteRemoteCurl(user, password, ip, ppsNumber, newQueue string) error {
	// Configuración de la conexión SSH
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Para desarrollo, no usar en producción
		Timeout:         5 * time.Second,
	}

	// Conectar al servidor
	client, err := ssh.Dial("tcp", ip+":22", config)
	if err != nil {
		return fmt.Errorf("fallo al conectarse al servidor %s: %w", ip, err)
	}
	defer client.Close()

	fmt.Printf("Conectado al servidor con la IP: %s\n", ip)

	// Crear una sesión SSH
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("fallo al crear una sesión SSH: %w", err)
	}
	defer session.Close()

	// Comando curl que se ejecutará en el servidor remoto
	curlCommand := fmt.Sprintf(
		"curl -X POST http://10.115.43.26:8181/api/mhs/configure_pps_queue_size/%s/%s -H 'Content-Type: application/json' -H 'cache-control: no-cache'",
		ppsNumber, newQueue,
	)

	// Ejecuta el comando curl
	output, err := session.CombinedOutput(curlCommand)
	if err != nil {
		return fmt.Errorf("fallo al ejecutar el comando curl: %w", err)
	}

	fmt.Printf("Resultado del servidor %s: %s\n", ip, output)
	return nil
}

func ConexionSSH(user, password, ip, command string) error {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Para desarrollo, no usar en producción
		Timeout:         5 * time.Second,
	}

	// Conectar al servidor
	client, err := ssh.Dial("tcp", ip+":22", config)
	if err != nil {
		return fmt.Errorf("fallo al conectarse al servidor %s: %w", ip, err)
	}
	defer client.Close()

	fmt.Printf("Conectado al servidor con la IP: %s\n", ip)

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("fallo al crear una sesión SSH: %w", err)
	}
	defer session.Close()

	// Conectar stdout y stderr para la sesión
	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("no se pudo conectar stdout: %w", err)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("no se pudo conectar stderr: %w", err)
	}

	// Ejecutar el comando para limpiar caché
	if err := session.Run(command); err != nil {
		return fmt.Errorf("fallo al ejecutar el comando: %w", err)
	}

	// Leer la salida de stdout
	buf := make([]byte, 1024)
	n, _ := stdout.Read(buf)
	fmt.Printf("Output del servidor %s: %s\n", ip, buf[:n])

	// Leer la salida de stderr
	errBuf := make([]byte, 1024)
	nErr, _ := stderr.Read(errBuf)
	if nErr > 0 {
		fmt.Printf("Error output del servidor %s: %s\n", ip, errBuf[:nErr])
	}

	return nil
}

// ClearCacheOnServers recorre una lista de IPs y ejecuta el comando para limpiar la caché en cada uno.
func ClearCacheOnServersFLR() {
	cfg := config.LoadConfig()
	serversIP := []string{
		cfg.FLRApp,
		cfg.FLR_DB,
		cfg.FLR_METRICS,
		cfg.FLR_OPC,
		cfg.FLR_FM,
	}

	command := "echo '" + cfg.PASSWORD + "' | sudo -S -p '' bash -c 'free -m && sync && echo 3 > /proc/sys/vm/drop_caches && free -m'"

	for _, ip := range serversIP {
		if err := ConexionSSH(cfg.SSHUser, cfg.PASSWORD, ip, command); err != nil {
			fmt.Printf("Error en el servidor %s: %v\n", ip, err)
		}
	}
}

func ClearCacheOnServersSBS() {
	cfg := config.LoadConfig()
	serversIP := []string{
		cfg.SBS_PUPPET,
		cfg.SBS_INTERFACE,
		cfg.SBS_CORE,
		cfg.SBS_PLATFORM_API,
		cfg.SBS_PLATFORM_CORE,
		cfg.SBS_PLATFORM_DB,
		cfg.SBS_METRICS,
	}

	command := "echo '" + cfg.SBS_PASSWORD + "' | sudo -S -p '' bash -c 'free -m && sync && echo 3 > /proc/sys/vm/drop_caches && free -m'"

	for _, ip := range serversIP {
		if err := ConexionSSH(cfg.SSHUser, cfg.SBS_PASSWORD, ip, command); err != nil {
			fmt.Printf("Error en el servidor %s: %v\n", ip, err)
		}
	}
}

func ClearCacheOnStaging() {
	cfg := config.LoadConfig()
	sbsIP := []string{
		cfg.SBS_STAGING,
	}

	commands := `
		free -m && sudo sync && echo 3 | sudo tee /proc/sys/vm/drop_caches && free -m
	`

	for _, ip := range sbsIP {
		if err := ConexionSSH(cfg.SSHUser, cfg.PASSWORD, ip, commands); err != nil {
			fmt.Printf("Error en el servidor %s: %v\n", ip, err)
		} else {
			fmt.Printf("Comando ejecutado exitosamente en el servidor %s\n", ip)
		}
	}
}

func ClearCacheSbs3() {
	cfg := config.LoadConfig()
	sbsIP := []string{
		cfg.SBS_BRIGDE,
		cfg.SBS_OPC,
	}

	command := "echo '" + cfg.PASSWORD + "' | sudo -S -p '' bash -c 'free -m && sync && echo 3 > /proc/sys/vm/drop_caches && free -m'"

	for _, ip := range sbsIP {
		if err := ConexionSSH(cfg.SSHUser, cfg.PASSWORD, ip, command); err != nil {
			fmt.Printf("Error en el servidor %s: %v\n", ip, err)
		} else {
			fmt.Printf("Comando ejecutado exitosamente en el servidor %s\n", ip)
		}
	}
}

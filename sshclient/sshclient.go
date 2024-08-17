package sshclient

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

func Conexion_ssh(user, password, ip string) error {
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

	// Conectar stdin, stdout, stderr para la sesión
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("no se pudo conectar stdin: %w", err)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("no se pudo conectar stdout: %w", err)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("no se pudo conectar stderr: %w", err)
	}

	// Empezar una sesión tipo shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("fallo al iniciar el shell: %w", err)
	}

	// Ejecutar los comandos
	// Ejecutar todos los comandos dentro de sudo con -S y -p para la contraseña
	commands := []string{
		fmt.Sprintf("echo '%s' | sudo -S -p '' bash -c 'free -m && sync && echo 3 > /proc/sys/vm/drop_caches && free -m'", password),
	}

	for _, cmd := range commands {
		fmt.Fprintln(stdin, cmd)
	}

	// Cerrar stdin para indicar el fin de los comandos
	stdin.Close()

	// Esperar a que termine la sesión
	session.Wait()

	// Leer la salida de stdout
	buf := make([]byte, 1024)
	n, _ := stdout.Read(buf)
	fmt.Printf("Output: %s\n", buf[:n])

	// Leer la salida de stderr
	errBuf := make([]byte, 1024)
	nErr, _ := stderr.Read(errBuf)
	if nErr > 0 {
		fmt.Printf("Error output: %s\n", errBuf[:nErr])
	}

	return nil
}

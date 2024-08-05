package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang_tui/config"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/crypto/ssh"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("#00ff00"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4).Foreground(lipgloss.Color("#FF06B7"))
	cfg               config.Config
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list       list.Model
	choice     string
	quitting   bool
	sshMessage string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				m.list.Title = "Opción seleccionada: " + m.choice

				switch m.choice {
				case "Limpiar cashe de los servidores FLR":
					m.sshMessage = sshConnect(cfg.SSHUser, cfg.SSHPassword, cfg.ServerFLR)
				case "Limpiar cashe de los servidores SBS":
					m.sshMessage = sshConnect(cfg.SSHUser, cfg.SSHPassword, cfg.ServerSBS)
				case "Buscar tarea en el server de FLR":
					m.sshMessage = sshConnect(cfg.SSHUser, cfg.SSHPassword, cfg.ServerFLR)
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return quitTextStyle.Render("Que tengas un buen turno!")
	}
	return "\n" + m.list.View() + "\n" + titleStyle.Render(m.sshMessage)
}

func main() {
	cfg = config.LoadConfig()

	items := []list.Item{
		item("Limpiar cashe de los servidores FLR"),
		item("Limpiar cashe de los servidores SBS"),
		item("Buscar tarea en el server de FLR"),
		// item("Cancelar tareas Pick"),
		// item("Cancelar tareas de Put"),
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Bienvenido, que quires hacer?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{list: l}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error corriendo el programa:", err)
		os.Exit(1)
	}
}

func sshConnect(user, password, host string) string {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		return fmt.Sprintf("Error conectando a %s: %v", host, err)
	}
	defer client.Close()

	return fmt.Sprintf("Conexión exitosa a %s", host)
}

package main

import (
	"fmt"
	"golang_tui/config"
	"golang_tui/sshclient"
	"golang_tui/utils"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("#00ff00"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4).Foreground(lipgloss.Color("#FF06B7"))
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
	list          list.Model
	choice        string
	quitting      bool
	step          int
	ppsNumber     string // Almacenará el número de PPS ingresado
	newQueue      string // Almacenará el número de la nueva cola
	inputField    string // Campo de entrada interactiva
	collectingPPS bool   // Indicador de si estamos ingresando el PPS
	collectingQ   bool   // Indicador de si estamos ingresando la nueva cola
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); {
		case keypress == "q" || keypress == "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case keypress == "enter":
			// Verifica el paso actual
			if m.step == 0 {
				// Seleccionando la opción principal
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.choice = string(i)
				}
				switch m.choice {
				case "Limpiar cashe de los servidores FLR":
					fmt.Println("Limpiando caché de los servidores FLR...")
					sshclient.ClearCacheOnServersFLR()
				case "Limpiar cashe de los servidores SBS":
					fmt.Println("Limpiando caché de los servidores SBS...")
					sshclient.ClearCacheOnServersSBS()
					sshclient.ClearCacheOnStaging()
					sshclient.ClearCacheSbs3()
				case "Comprobar que los servidores esten corriendo":
					fmt.Println("Realizando un ping a los servidores...")
					utils.PingServers()
				case "Cambiar colas de pps":
					m.step = 1
					m.inputField = "" // Resetea el campo de entrada
					m.collectingPPS = true
				}
			} else if m.collectingPPS {
				// Confirmamos el número de PPS
				m.ppsNumber = m.inputField
				m.inputField = "" // Resetea el campo de entrada
				m.collectingPPS = false
				m.collectingQ = true
				m.step = 2
			} else if m.collectingQ {
				// Confirmamos el número de la nueva cola
				m.newQueue = m.inputField
				m.collectingQ = false

				// Configuración de conexión y ejecución del comando curl remoto
				cfg := config.LoadConfig()
				err := sshclient.ExecuteRemoteCurl(cfg.SSHUser, cfg.PASSWORD, cfg.SBS_STAGING, m.ppsNumber, m.newQueue)
				if err != nil {
					fmt.Printf("Error al ejecutar el comando curl remoto: %v\n", err)
				}
				return m, tea.Quit
			}

		case keypress == "backspace":
			if len(m.inputField) > 0 {
				m.inputField = m.inputField[:len(m.inputField)-1] // Elimina el último carácter
			}

		default:
			if m.collectingPPS || m.collectingQ {
				m.inputField += keypress // Añade la tecla al campo de entrada
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	if m.quitting {
		return quitTextStyle.Render("Que tengas un buen turno maquina!")
	}
	switch m.step {
	case 1:
		return titleStyle.Render("Que pps necesita cambiar?") + "\n" + m.inputField
	case 2:
		return titleStyle.Render(fmt.Sprintf("Ingrese la nueva cola de la pps %s:", m.ppsNumber)) + "\n" + m.inputField
	default:
		return "\n" + m.list.View()
	}
}

func main() {
	items := []list.Item{
		item("Limpiar cashe de los servidores FLR"),
		item("Limpiar cashe de los servidores SBS"),
		item("Comprobar que los servidores esten corriendo"),
		item("Cambiar colas de pps"),
		//item("Cambiar sector preference"),
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Bienvenido compañero!!!, que quieres hacer?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := &model{list: l}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error corriendo el programa:", err)
		os.Exit(1)
	}
}

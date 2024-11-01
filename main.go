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
					m.step = 5
					go func() {
						sshclient.ClearCacheOnServersFLR()
						m.step = 0
						tea.NewProgram(m).Send(tea.WindowSizeMsg{})
					}()
					return m, nil

				case "Limpiar cashe de los servidores SBS":
					m.step = 6
					go func() {
						sshclient.ClearCacheOnServersSBS()
						sshclient.ClearCacheOnStaging()
						sshclient.ClearCacheSbs3()
						m.step = 0
						tea.NewProgram(m).Send(tea.WindowSizeMsg{})
					}()
					return m, nil

				case "Comprobar que los servidores esten corriendo":
					m.step = 4
					go func() {
						results := utils.PingServers()

						// Procesar resultados conforme se reciben
						for result := range results {
							fmt.Println(result) // Imprime cada resultado en tiempo real
						}

						// Reiniciar estado y regresar al menú inicial
						m.step = 0
						tea.NewProgram(m).Send(tea.WindowSizeMsg{})
					}()
					return m, nil

				case "Cambiar colas de pps":
					m.step = 1
					m.inputField = "" // Resetea el campo de entrada
					m.collectingPPS = true
				case "Largo de colas default":
					fmt.Println("Configurando las colas de las PPS con valores predeterminados y específicos...")

					// Cambiar el paso a 3 para mostrar el mensaje de estado
					m.step = 3

					// Ejecutar la configuración en un goroutine para no bloquear la interfaz
					go func() {
						cfg := config.LoadConfig()
						// Mapa con las colas y sus valores específicos
						specificQueues := map[string]int{
							"3":  7,
							"4":  7,
							"12": 4,
							"15": 7,
							"16": 7,
							"17": 4,
						}

						err := sshclient.ExecuteDefaultQueuesWithExceptions(cfg.SSHUser, cfg.PASSWORD, cfg.SBS_STAGING, specificQueues)
						if err != nil {
							fmt.Printf("Error al configurar las colas de PPS: %v\n", err)
						}

						// Reiniciar estado y regresar al menú inicial
						m.step = 0
						m.ppsNumber = ""
						m.newQueue = ""
						m.inputField = ""
						m.collectingPPS = false
						m.collectingQ = false
						m.choice = ""

						// Actualizar la interfaz de usuario para volver al menú inicial
						tea.NewProgram(m).Send(tea.WindowSizeMsg{})
					}()
					return m, nil

				}
			} else if m.collectingPPS {
				m.ppsNumber = m.inputField
				m.inputField = "" // Resetea el campo de entrada
				m.collectingPPS = false
				m.collectingQ = true
				m.step = 2
			} else if m.collectingQ {
				m.newQueue = m.inputField
				m.collectingQ = false

				// Configuración de conexión y ejecución del comando curl remoto
				cfg := config.LoadConfig()
				err := sshclient.ExecuteRemoteCurl(cfg.SSHUser, cfg.PASSWORD, cfg.SBS_STAGING, m.ppsNumber, m.newQueue)
				if err != nil {
					fmt.Printf("Error al ejecutar el comando curl remoto: %v\n", err)
				}
				// Reiniciar estado y regresar al menú inicial
				m.step = 0
				m.ppsNumber = ""
				m.newQueue = ""
				m.inputField = ""
				m.collectingPPS = false
				m.collectingQ = false
				m.choice = ""

				// Regresa al menú inicial
				return m, nil
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
		return quitTextStyle.Render("Que tengas un buen turno maquina!!!")
	}
	switch m.step {
	case 1:
		return titleStyle.Render("Que pps necesitas cambiar?") + "\n" + m.inputField
	case 2:
		return titleStyle.Render(fmt.Sprintf("Ingrese la nueva cola de la pps %s:", m.ppsNumber)) + "\n" + m.inputField
	case 3:
		return titleStyle.Render("Configurando las pps en modo default...") + "\n"
	case 4:
		return titleStyle.Render("Realizando un ping a cada servidor...") + "\n"
	case 5:
		return titleStyle.Render("Limpiando memorias FLR...") + "\n"
	case 6:
		return titleStyle.Render("Limpiando memorias SBS...") + "\n"
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
		item("Largo de colas default"),
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

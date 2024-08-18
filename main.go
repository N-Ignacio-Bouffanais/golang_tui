package main

import (
	"fmt"
	"golang_tui/config"
	"golang_tui/sshclient"
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
	checkedItemStyle  = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("34"))
)

type item struct {
	title   string
	checked bool
}

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

	checkbox := " "
	if i.checked {
		checkbox = "✓"
	}

	str := fmt.Sprintf("%s %d. %s", checkbox, index+1, i.title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list         list.Model
	choice       string
	quitting     bool
	submenuOpen  bool
	submenuItems []list.Item
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
			if m.submenuOpen {
				selectedItems := []string{}
				for _, listItem := range m.list.Items() {
					i, ok := listItem.(item)
					if ok && i.checked {
						selectedItems = append(selectedItems, i.title)
					}
				}
				fmt.Println("Seleccionaste las siguientes opciones:", selectedItems)
				return m, tea.Quit
			}

			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = i.title
			}
			switch m.choice {
			case "Limpiar cashe de los servidores FLR":
				config := config.LoadConfig() // Carga la configuración
				err := sshclient.Conexion_ssh(config.SSHUser, config.PASSWORD, config.FLR_DB)
				if err != nil {
					fmt.Println("Error:", err)
				}
				fmt.Println("FLR")
			case "Limpiar cashe de los servidores SBS":
				m.submenuOpen = true
				m.submenuItems = []list.Item{
					item{title: "Opción 1: Limpiar SBS A"},
					item{title: "Opción 2: Limpiar SBS B"},
					item{title: "Opción 3: Limpiar SBS C"},
					item{title: "Opción 4: Limpiar SBS D"},
				}
				m.list.SetItems(m.submenuItems)
			case "Buscar tarea en el server de FLR":
				fmt.Println("")
			}

			if m.submenuOpen {
				return m, nil
			}

			return m, tea.Quit

		case " ":
			if m.submenuOpen {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					i.checked = !i.checked
					m.list.SetItem(m.list.Index(), i)
				}
				return m, nil
			}
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

	return "\n" + m.list.View()
}

func main() {
	items := []list.Item{
		item{title: "Limpiar cashe de los servidores FLR"},
		item{title: "Limpiar cashe de los servidores SBS"},
		item{title: "Buscar tarea en el server de FLR"},
		item{title: "Cancelar tareas Pick"},
		item{title: "Cancelar tareas de Put"},
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Bienvenido, que quieres hacer?"
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

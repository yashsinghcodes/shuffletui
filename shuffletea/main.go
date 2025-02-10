package shuffletea

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/yashsinghcodes/shuffle-tui/sqlc/queries"
)


const url = "https://shuffler.io/api/v1/"

type Model struct {
    status      int
    err         error
    Username    string
    SSHkey      []byte
    ApiKey      string
    Workflows   list.Model
}

type ApiModel struct {
    Input       textinput.Model
    SSHkey      []byte
    Username    string
}


type Workflows list.Model
var Queries *shuffletui.Queries
var DB *sql.DB

func GetWorkflow() tea.Msg {
    res, err := http.Get(fmt.Sprintf("%s/workflows", url))
    _ = res
    if err != nil {
        log.Error("Failed to get the workflow for the user")
        return nil
    }
    return Workflows(list.Model{})
}

func (m Model) Init() (tea.Cmd) {
    return GetWorkflow
}

func (m ApiModel) Init() (tea.Cmd) {
    return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model,  tea.Cmd) {
    return m, nil
}

func (m ApiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds tea.Cmd
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.Type == tea.KeyEnter {
            return Model{
                SSHkey: m.SSHkey,
                Username: m.Username,
                ApiKey: m.Input.Value(),
            }, nil
        }
    }
    m.Input, _ = m.Input.Update(msg)

    return m, cmds
}

func (m Model) View() string {
    if m.err != nil {
        return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
    }

    s := fmt.Sprintf("Checking %s ... ", url)
    if m.status > 0 {
        s += fmt.Sprintf("%d %s!", m.status, http.StatusText(m.status))
    }

    s += fmt.Sprintf("\nHello %s\n", m.Username)

    return "\n" + s + "\n\n"
}

func (m ApiModel) View() string {
    return fmt.Sprintf("ApiKey: %s", m.Input.View())
}

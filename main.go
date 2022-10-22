package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harrisoncramer/gitlab-dash/utils"
)

func main() {

	p := tea.NewProgram(model{
		response: make(chan []byte),
		loading:  true,
		spinner:  spinner.NewModel(),
		list:     list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		json:     []MR{},
	}, tea.WithAltScreen())

	if p.Start() != nil {
		fmt.Println("could not start program")
		os.Exit(1)
	}
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type responseMsg []byte
type MR struct {
	Description string `json:"description"`
	Title       string `json:"title"`
}

type model struct {
	response chan []byte
	loading  bool
	spinner  spinner.Model
	list     list.Model
	json     []MR
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func (m model) Init() tea.Cmd {
	return tea.Batch(
		spinner.Tick,
		listenForActivity(m.response),
		waitForActivity(m.response),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		default:
			return m, nil
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case responseMsg:
		m.loading = false
		r := <-m.response
		var jsonData []MR
		err := json.Unmarshal(r, &jsonData)

		utils.Must("Error unmarshalling data: %g", err)

		listOfMrs := []list.Item{}
		for _, mr := range jsonData {
			listOfMrs = append(listOfMrs, item{
				title: mr.Title,
				desc:  mr.Description,
			})
		}

		m.list = list.New(listOfMrs, list.NewDefaultDelegate(), 0, 0)
		m.list.Title = "MRs"
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m model) View() string {
	if m.loading {
		return m.spinner.View()
	} else {
		return docStyle.Render(m.list.View())
	}
}

/* Sends some data to the channel upon completion of the GET */
func listenForActivity(response chan []byte) tea.Cmd {
	return func() tea.Msg {
		for {
			req, err := http.NewRequest("GET", "https://gitlab.com/api/v4/projects/40444811/merge_requests", nil)
			req.Header.Set("PRIVATE-TOKEN", "your-token")
			utils.Must("Error setting up request: %g", err)

			client := &http.Client{}
			resp, err := client.Do(req)

			utils.Must("Error fetching MRs: %g", err)
			defer resp.Body.Close()

			r, err := ioutil.ReadAll(resp.Body)

			utils.Must("Error reading body response: %g", err)

			// Check Status of response

			response <- r
		}
	}
}

// Put the channel's data into a response byte slice. This will be returned back to the Update function
func waitForActivity(response chan []byte) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-response)
	}
}

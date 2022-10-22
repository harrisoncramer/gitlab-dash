package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/harrisoncramer/gitlab-dash/utils"
)

func main() {

	p := tea.NewProgram(model{
		response: make(chan []byte),
		spinner:  spinner.NewModel(),
		loading:  true,
	}, tea.WithAltScreen())

	if p.Start() != nil {
		fmt.Println("could not start program")
		os.Exit(1)
	}
}

type responseMsg []byte
type MR struct {
	Description string `json:"description"`
	Title       string `json:"title"`
}

/* Sends some data to the channel upon completion of the GET */
func listenForActivity(response chan []byte) tea.Cmd {
	return func() tea.Msg {
		for {
			req, err := http.NewRequest("GET", "https://gitlab.com/api/v4/projects/40444811/merge_requests", nil)
			req.Header.Set("PRIVATE-TOKEN", "glpat-WbMsVxMocvYW7U-NB1MS")
			utils.Must("Error setting up request: %g", err)

			client := &http.Client{}
			resp, err := client.Do(req)

			utils.Must("Error fetching MRs: %g", err)
			defer resp.Body.Close()

			r, err := ioutil.ReadAll(resp.Body)

			utils.Must("Error reading body response: %g", err)

			response <- r
		}
	}
}

func handleResponse(response []byte) tea.Cmd {
	return func() tea.Msg {
		return "hi"
	}
}

// Put the channel's data into a response byte slice. This will be returned back to the Update function
func waitForActivity(response chan []byte) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-response)
	}
}

type model struct {
	response chan []byte
	loading  bool
	spinner  spinner.Model
	json     []MR
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		spinner.Tick,
		listenForActivity(m.response), // generate activity
		waitForActivity(m.response),   // wait for activity
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
	case responseMsg:
		m.loading = false
		r := <-m.response
		var jsonData []MR
		err := json.Unmarshal(r, &jsonData)
		utils.Must("Error unmarshalling data: %g", err)
		m.json = jsonData
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
	} else if len(m.json) != 0 {
		return m.json[0].Title
	} else {
		return "Done."
	}
}

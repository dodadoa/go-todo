package main

import (
	"log"
	"os"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/gocarina/gocsv"
)

type TodoCSV struct {
	Name   string `csv:"task"`
	Status string `csv:"status"`
}

type AppState struct {
	isInput bool
	todos   [][]string
}

func draw(p ui.Drawable, l ui.Drawable, tasks ui.Drawable, i ui.Drawable) {
	ui.Render(p, l, tasks, i)
}

func main() {

	COMMANDS := []string{
		"[Up] Move up",
		"[Down] Move down",
		"[q] Quit",
		"[a] Add",
		"[d] Delete",
		"[m] Mark",
		"[u] Unmark",
		"[f] Filter only done",
		"[n] Filter only not done",
		"[r] Remove filter",
		"[s] Save CSV",
	}

	clientsFile, err := os.OpenFile("clients.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to manage file: %v", err)
	}

	todoCSVs := []*TodoCSV{}
	if err := gocsv.UnmarshalFile(clientsFile, &todoCSVs); err != nil {
		if err != gocsv.ErrEmptyCSVFile {
			log.Fatalf("Failed to unmarshal CSV: %v", err)
		}
	}

	defer clientsFile.Close()

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	appState := &AppState{
		isInput: false,
		todos:   [][]string{},
	}

	for _, todoCSV := range todoCSVs {
		appState.todos = append(appState.todos, []string{todoCSV.Name, todoCSV.Status})
	}

	p := widgets.NewParagraph()
	p.Text = "Todo App"
	p.SetRect(0, 0, 50, 25)

	l := widgets.NewList()
	l.Title = "Commands"
	l.Rows = COMMANDS
	l.SetRect(0, 2, 50, 15)

	tasks := widgets.NewTable()
	tasks.TextAlignment = ui.AlignCenter
	tasks.RowSeparator = true
	tasks.Title = "Tasks"
	tasks.Rows = [][]string{
		{"task", "status"},
	}
	if len(appState.todos) > 0 {
		tasks.Rows = append(tasks.Rows, appState.todos...)
	}
	tasks.SetRect(0, 15, 50, 35)

	i := widgets.NewParagraph()
	i.Title = ""
	i.Text = ""
	i.SetRect(0, 0, 0, 0)

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Millisecond).C

	point := 1

	for {
		select {
		case e := <-uiEvents:
			if appState.isInput {
				switch e.ID {
				case "<Escape>":
					tasks.Title = "Tasks"
					tasks.SetRect(0, 15, 50, 35)
					i.SetRect(0, 0, 0, 0)
					i.Title = ""
					i.Text = ""
					l.Rows = COMMANDS
					appState.isInput = false
				case "<Backspace>":
					i.Text = i.Text[:len(i.Text)-1]
				case "<Space>":
					i.Text = i.Text + " "
				case "<Enter>":
					tasks.Rows = append(tasks.Rows, []string{i.Text, "todo"})
					appState.todos = tasks.Rows[1:]
					i.Text = ""
					tasks.Title = "Tasks"
					l.Rows = COMMANDS
					tasks.SetRect(0, 15, 50, 35)
					i.SetRect(0, 0, 0, 0)
					appState.isInput = false
					point = 1
				default:
					i.Text = i.Text + e.ID
				}
			} else {
				switch e.ID {
				case "q", "<C-c>":
					return
				case "<Up>":
					if point > 1 {
						tasks.RowStyles[point] = ui.NewStyle(ui.ColorClear)
						point--
					}
				case "<Down>":
					if len(tasks.Rows)-1 > point {
						tasks.RowStyles[point] = ui.NewStyle(ui.ColorClear)
						point++
					}
				case "d":
					if len(tasks.Rows) == 1 {
						break
					}
					tasks.Rows = append(tasks.Rows[:point], tasks.Rows[point+1:]...)
					appState.todos = tasks.Rows[1:]
					if point != 1 {
						point--
					}
				case "m":
					tasks.Rows[point][1] = "done"
				case "u":
					tasks.Rows[point][1] = "todo"
				case "a":
					tasks.Title = "Add"
					l.Rows = []string{
						"[Escape] Quit",
					}
					tasks.SetRect(0, 0, 0, 0)
					i.Title = "Input task"
					i.SetRect(0, 15, 50, 25)
					appState.isInput = true
				case "f":
					tasks.Rows = [][]string{
						{"task", "status"},
					}

					for _, todo := range appState.todos {
						if todo[1] == "done" {
							tasks.Rows = append(tasks.Rows, todo)
						}
					}
				case "n":
					tasks.Rows = [][]string{
						{"task", "status"},
					}

					for _, todo := range appState.todos {
						if todo[1] == "todo" {
							tasks.Rows = append(tasks.Rows, todo)
						}
					}
				case "r":
					tasks.Rows = appState.todos
				case "s":
					toSave := []*TodoCSV{}
					for _, todo := range appState.todos {
						toSave = append(toSave, &TodoCSV{
							Name:   todo[0],
							Status: todo[1],
						})
					}

					clientsFile.Truncate(0)

					if _, err := clientsFile.Seek(0, 0); err != nil {
						panic(err)
					}

					err := gocsv.MarshalFile(&toSave, clientsFile)
					if err != nil {
						panic(err)
					}
				}

			}

		case <-ticker:
			tasks.RowStyles[point] = ui.NewStyle(ui.ColorBlue, ui.ColorBlack)
			draw(p, l, tasks, i)
		}

	}
}

// Minimal data table displaying a few rows of employee data.
package main

import (
	"log"

	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	table := widgets.NewTable(
		widgets.TableColumn{Title: "Name", Width: 20},
		widgets.TableColumn{Title: "Role", Width: 20},
		widgets.TableColumn{Title: "City", Width: 15},
	)
	table.SetRows([][]string{
		{"Alice Johnson", "Engineer", "Seattle"},
		{"Bob Smith", "Designer", "Portland"},
		{"Carol Lee", "Manager", "Denver"},
		{"Dan Kim", "Analyst", "Austin"},
	})

	if err := fluffy.Run(
		fluffy.VStack(
			fluffy.Label("Team Directory"),
			table,
		),
	); err != nil {
		log.Fatal(err)
	}
}

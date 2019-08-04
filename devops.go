package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"sync"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

var showWork, showPr, verbose bool
var prCount int

func main() {
	// Fetch the access stuff from environment
	acc := os.Getenv("AZUREDEVOPS_ACCOUNT")
	proj := os.Getenv("AZUREDEVOPS_PROJECT")
	token := os.Getenv("AZUREDEVOPS_TOKEN")
	repo := os.Getenv("AZUREDEVOPS_REPO")

	flag.BoolVar(&showWork, "wit", false, "Show workitem stats")
	flag.IntVar(&prCount, "pr", 0, "Number of pull requests to process for count")
	flag.BoolVar(&verbose, "v", false, "Show verbose output")

	flag.Parse()

	if showWork {
		showWorkStats(acc, proj, token)
	}

	// Connect to repo
	if prCount > 0 {
		showPrStats(acc, proj, token, repo, prCount)
	}

	/*rs := []ReviewerStat{
		{
			"Abhi", 20,
		},
		{
			"01234567890123456789012345678943243224", 10,
		},
	}
	savePrStatImage(rs, 40, "img.png")
	*/
}

func showWorkStats(acc, proj, token string) {
	// Get the list of epics from a epic's only query
	epicWitQuery := "0325c50f-3511-4266-a9fe-80b989492c76"
	if verbose {
		fmt.Printf("Fetching epics using query %v\n", epicWitQuery)
	}

	parentEpics, err := getEpics(acc, proj, token, epicWitQuery)
	if err != nil {
		fmt.Println("Error getting list of epics:", err)
		return
	}

	// For each epic start a go-routine and fetch all workitems that are child of it
	var wg sync.WaitGroup
	m := &sync.Mutex{}
	for _, epic := range parentEpics {
		if verbose {
			fmt.Printf("Fetching epic %v: %v\n", epic.Id, epic.Title)
		}

		wg.Add(1)
		go func(epic int, m *sync.Mutex) {
			defer wg.Done()
			printEpicStat(acc, proj, token, epic, m)
		}(epic.Id, m)
	}

	wg.Wait()
}

func getEpics(acc, proj, token, queryID string) ([]WorkItem, error) {
	q := NewWork(acc, proj, token)
	epics, err := q.GetWorkitems(queryID)
	if err != nil {
		return nil, err
	}

	return epics, nil
}

func printEpicStat(acc, proj, token string, parentEpic int, m *sync.Mutex) {
	q := NewWork(acc, proj, token)

	epic, err := q.GetWorkitem(parentEpic)
	if err != nil {
		fmt.Println("Error!!!", err)
		return
	}

	wits, err := q.RefreshWit(parentEpic)
	m.Lock()
	defer m.Unlock()
	fmt.Printf("%v: %v (%v)\n", epic.Id, epic.Title, epic.AssignedTo)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, w := range wits {
		fmt.Println(w)
	}
}

func showPrStats(acc, proj, token, repo string, count int) {
	r := NewRepo(acc, proj, token, repo)
	if r.err != nil {
		fmt.Println(r.err)
		return
	}

	// Fetch PRs
	revStats, max := r.GetPullRequestReviewsByUser(count)
	barmax := float32(80.0)

	// Output!!
	if verbose {
		fmt.Println("\nReviewer Stats\n")
	}
	for _, revStat := range revStats {
		bar := int((barmax / float32(max)) * float32(revStat.Count))
		percentage := float32(revStat.Count) / float32(count) * 100.0
		fmt.Printf("%30s %4d (%4.1f%%) ", revStat.Name, revStat.Count, percentage)
		fmt.Print("[")
		i := 0
		for ; i < bar; i++ {
			fmt.Print("#")
		}

		for ; i < int(barmax); i++ {
			fmt.Print("-")
		}
		fmt.Print("]\n")
	}

	savePrStatImage(revStats, count, "revstat.png")
}

func savePrStatImage(reviewers []ReviewerStat, prCount int, fileName string) {
	if verbose {
		fmt.Println("Generating image ", fileName)
	}

	nReviewers := len(reviewers)
	w := 1000.0
	h := 20.0*float64(nReviewers) + 50.0
	// Initialize the graphic context on an RGBA image
	dest := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	gc := draw2dimg.NewGraphicContext(dest)

	// Font stuff setup
	draw2d.SetFontFolder(".")

	// Draw the border and title
	gc.SetFontSize(14)
	drawRect(gc, 0, 0, w, h, image.Black, image.White)
	str := fmt.Sprintf("Reviewer Stats for %v pull requests", prCount)
	gc.SetFillColor(color.Black)
	gc.FillStringAt(str, 10, 20)

	y := 60.0
	rightX := 300.0 // Right aligning all text to be here
	gap := 10.0     // gap between name and bar chart
	maxNameLen := 30

	maxBarWidth := (w - 10) - (rightX + gap)

	for _, reviewer := range reviewers {
		// trim or use the name if it fits
		str := reviewer.Name
		if len(str) > maxNameLen {
			str = str[:maxNameLen]
		}
		// find the width of the name and right align and print
		gc.SetFontSize(12)
		l, _, r, _ := gc.GetStringBounds(str)
		strW := r - l
		strH := 15.0

		textColor := color.RGBA{50, 50, 50, 0xff}
		x := (rightX - strW)
		gc.SetFillColor(textColor)
		gc.FillStringAt(str, x, y)

		// Draw the bar
		width := (maxBarWidth / float64(prCount)) * float64(reviewer.Count)
		x = rightX + gap
		barFillCol := color.RGBA{100, 100, 100, 0xff}
		drawRect(gc, x, y-strH, width, strH, barFillCol, barFillCol)
		drawRect(gc, x, y-strH, maxBarWidth, strH, color.Black, color.Transparent)

		val := fmt.Sprintf("%v", reviewer.Count)
		gc.SetFillColor(textColor)
		gc.SetFontSize(10)
		gc.FillStringAt(val, x+width+10.0, y-2)

		y += 20
	}
	draw2dimg.SaveToPngFile(fileName, dest)
	fmt.Println("Generated", fileName)
}

func drawRect(gc *draw2dimg.GraphicContext, x, y, w, h float64, lineColor, fillColor color.Color) {
	gc.SetStrokeColor(lineColor)
	gc.SetFillColor(fillColor)
	gc.SetLineWidth(2)
	gc.MoveTo(x, y)
	gc.LineTo(x+w, y)
	gc.LineTo(x+w, y+h)
	gc.LineTo(x, y+h)
	gc.LineTo(x, y)
	gc.Close()
	gc.FillStroke()

}

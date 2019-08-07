package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

var (
	WitNotDoneColor    = color.RGBA{0xff, 0xaa, 0xaa, 0xff} // reddish
	WitInProgressColor = color.RGBA{0xff, 0xff, 0xa0, 0xff} // yellowish
	WitDoneColor       = color.RGBA{0, 0xad, 0, 0xff}       // greenish
	WitUnKnownColor    = color.RGBA{0xaa, 0, 0xff, 0xff}    // purplish
)

// ================================================================================================
// PR related images
func savePrStatImage(reviewers []ReviewerStat, prCount int, fileName string) error {
	if verbose {
		fmt.Println("Generating image ", fileName)
	}

	nReviewers := len(reviewers)
	w := 1000.0

	// dedicate pixel for header, then per row and then footer
	h := 50.0 + 20.0*float64(nReviewers) + 20.0
	// Initialize the graphic context on an RGBA image
	dest := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	gc := draw2dimg.NewGraphicContext(dest)

	// Font stuff setup
	draw2d.SetFontFolder(".")

	// Draw the border and title/header
	str := fmt.Sprintf("Reviewer Stats for %v pull requests", prCount)
	drawHeader(gc, str, w, h)

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

	drawFooter(gc, w, h)

	err := draw2dimg.SaveToPngFile(fileName, dest)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Println("Generated", fileName)
	}

	return nil
}

// ================================================================================================
// Workitem images
func saveWitStatImage(epicStat []EpicStat, fileName string) error {
	if verbose {
		fmt.Println("Generating image ", fileName)
	}

	nEpics := len(epicStat)

	// Find the max count for an Epic
	maxCount := 0
	for _, e := range epicStat {
		count := e.Done + e.InProgress + e.NotDone + e.Unknown
		if count > maxCount {
			maxCount = count
		}
	}

	w := 1000.0

	// dedicate pixel for header, then per row and then footer
	h := 50.0 + 60.0*float64(nEpics) + 20.0
	// Initialize the graphic context on an RGBA image
	dest := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	gc := draw2dimg.NewGraphicContext(dest)

	// Font stuff setup
	draw2d.SetFontFolder(".")

	// Header
	drawHeader(gc, "Epic Status", w, h)

	x, y := 10.0, 50.0
	// Content
	for _, es := range epicStat {
		retY := drawEpicStat(gc, es, maxCount, x, y, w)
		y = retY + 20 // gap between two epic
	}

	// Footer
	drawFooter(gc, w, h)

	err := draw2dimg.SaveToPngFile(fileName, dest)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Println("Generated", fileName)
	}

	return nil
}

// Draw the graph starting at x, y, using width w and return till how many pixel vertically stuff was written
func drawEpicStat(gc *draw2dimg.GraphicContext, es EpicStat, maxCount int, x, y, w float64) float64 {
	gc.SetFontSize(12)
	gc.SetFillColor(color.Black)
	str := fmt.Sprintf("%v : %s (%s)", es.Epic.Id, es.Epic.Title, es.Epic.AssignedTo)
	_, t, _, b := gc.GetStringBounds(str)

	// FillString aligns the lower line of text with y, so t actually is negative
	// So make the adjustment so that the top of text aligns with specified y
	y += -t

	gc.FillStringAt(str, x, y)

	y += b // The text bottom is here now

	y += 5 // Leave some gap between text and the bar
	barH := 15.0

	w -= 20.0 // actual width to use
	barW := (w / float64(maxCount)) * float64(es.Done)
	drawRect(gc, x, y, barW, barH, color.Black, WitDoneColor)
	x += barW

	barW = (w / float64(maxCount)) * float64(es.InProgress)
	drawRect(gc, x, y, barW, barH, color.Black, WitInProgressColor)
	x += barW

	barW = (w / float64(maxCount)) * float64(es.NotDone)
	drawRect(gc, x, y, barW, barH, color.Black, WitNotDoneColor)
	x += barW

	barW = (w / float64(maxCount)) * float64(es.Unknown)
	drawRect(gc, x, y, barW, barH, color.Black, WitUnKnownColor)
	x += barW

	y += barH

	return y
}

// ================================================================================================
// Common utilty

func drawHeader(gc *draw2dimg.GraphicContext, header string, w, h float64) {
	// Draw the border and title/header
	gc.SetFontSize(14)
	drawRect(gc, 0, 0, w, h, image.Black, image.White)
	gc.SetFillColor(color.Black)
	gc.FillStringAt(header, 10, 20)
}

func drawFooter(gc *draw2dimg.GraphicContext, w, h float64) {
	// Write the footer justified to right margin
	str := fmt.Sprintf("Generated at %v", time.Now().Format("01-02-2006 15:04:05"))
	gc.SetFontSize(8)

	gc.SetFillColor(color.Black)
	l, _, r, _ := gc.GetStringBounds(str)
	sw := r - l // width of the text
	gc.FillStringAt(str, w-sw-10, h-10)

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

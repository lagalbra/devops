package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// Fetch the access stuff from environment
	acc := os.Getenv("AZUREDEVOPS_ACCOUNT")
	proj := os.Getenv("AZUREDEVOPS_PROJECT")
	token := os.Getenv("AZUREDEVOPS_TOKEN")
	repo := os.Getenv("AZUREDEVOPS_REPO")

	args := os.Args[1:]

	showPr, showWork := false, false
	if len(args) == 0 {
		showPr, showWork = true, true
	}

	for _, a := range args {
		if strings.EqualFold(a, "pr") {
			showPr = true
		} else if strings.EqualFold(a, "wit") {
			showWork = true
		}
	}

	if showWork {
		q := NewWork(acc, proj, token)
		if q.err != nil {
			fmt.Println(q.err)
			return
		}

		parentEpics := []int{4913314, 4806129, 4884022, 4884072, 4884120, 4669527, 4043454, 3904108,
			3904063, 3904068, 3904071}
		for _, epic := range parentEpics {
			printEpicStat(q, epic)
		}

	}

	// Connect to repo
	if showPr {
		r := NewRepo(acc, proj, token, repo)
		if r.err != nil {
			fmt.Println(r.err)
			return
		}

		// Fetch PRs
		count := 400
		revStats, max := r.GetPullRequestReviewsByUser(count)
		barmax := float32(80.0)
		// Output!!

		fmt.Println("\nREVIEWER STATS\n")
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
	}
}

func printEpicStat(q *AzureDevopsWit, parentEpic int) {
	epic, err := q.GetWorkitem(parentEpic)
	if err != nil {
		fmt.Println("Error!!!", err)
		return
	}

	fmt.Printf("Fetching for %v: %v (%v)\n", epic.Id, epic.Title, epic.AssignedTo)
	wits, err := q.RefreshWit(parentEpic)
	if err != nil {
		fmt.Println(q.err)
		return
	}

	for _, w := range wits {
		fmt.Println(w)
	}
}

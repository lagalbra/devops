package main

import (
	"fmt"
	"os"
	"sync"
)

func main() {
	// Fetch the access stuff from environment
	acc := os.Getenv("AZUREDEVOPS_ACCOUNT")
	proj := os.Getenv("AZUREDEVOPS_PROJECT")
	token := os.Getenv("AZUREDEVOPS_TOKEN")
	repo := os.Getenv("AZUREDEVOPS_REPO")

	args := os.Args[1:]

	var showPr, showWork bool

	for _, a := range args {
		switch a {
		case "pr":
			showPr = true
		case "wit":
			showWork = true
		}
	}

	if showWork {
		showWorkStats(acc, proj, token)
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

func showWorkStats(acc, proj, token string) {
	// Get the list of epics from a epic's only query
	parentEpics, err := getEpics(acc, proj, token, "0325c50f-3511-4266-a9fe-80b989492c76")
	if err != nil {
		fmt.Println("Error getting list of epics:", err)
		return
	}

	// For each epic start a go-routine and fetch all workitems that are child of it
	var wg sync.WaitGroup
	m := &sync.Mutex{}
	for _, epic := range parentEpics {
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

	fmt.Printf("Fetching %v\n", parentEpic)
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

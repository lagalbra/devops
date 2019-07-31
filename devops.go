package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
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
	parentEpics := []int{4913314, 4806129, 4884022, 4884072, 4884120, 4669527, 4043454, 3904108,
		3904063, 3904068, 3904071}
	var wg sync.WaitGroup
	m := &sync.Mutex{}
	for _, epic := range parentEpics {
		wg.Add(1)
		go func(epic int, m *sync.Mutex) {
			defer wg.Done()
			printEpicStat(acc, proj, token, epic, m)
		}(epic, m)
	}

	wg.Wait()
}

func printEpicStat(acc, proj, token string, parentEpic int, m *sync.Mutex) {
	q := NewWork(acc, proj, token)
	if q.err != nil {
		fmt.Println(q.err)
		return
	}

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
		fmt.Println(q.err)
		return
	}

	for _, w := range wits {
		fmt.Println(w)
	}
}

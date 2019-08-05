package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
)

var showWork, showPr, verbose, noUpload bool
var prCount int
var semesterFilter bool

func main() {
	// Fetch the access stuff from environment
	acc := os.Getenv("AZUREDEVOPS_ACCOUNT")
	proj := os.Getenv("AZUREDEVOPS_PROJECT")
	token := os.Getenv("AZUREDEVOPS_TOKEN")
	repo := os.Getenv("AZUREDEVOPS_REPO")
	azStorageAcc := os.Getenv("AZURE_STORAGE_ACCOUNT")
	azStorageKey := os.Getenv("AZURE_STORAGE_ACCESS_KEY")

	// Setup command line parsing
	flag.BoolVar(&showWork, "wit", false, "Show workitem stats")
	flag.IntVar(&prCount, "pr", 0, "Number of pull requests to process for count")
	flag.BoolVar(&verbose, "v", false, "Show verbose output")
	flag.BoolVar(&noUpload, "nu", false, "Do not upload generated data into Azure")
	flag.BoolVar(&semesterFilter, "sem", false, "Filter workitems not finished in this semester")

	flag.Parse()

	if showWork {
		err := showWorkStats(acc, proj, token)
		if err != nil {
			fmt.Println("Error fetching work stats!!", err)
		}
	}

	if prCount > 0 {
		err := showPrStats(acc, proj, token, repo, prCount, azStorageAcc, azStorageKey)
		if err != nil {
			fmt.Println("Error fetching pull-request stats!!", err)
		}
	}
}

func showWorkStats(acc, proj, token string) error {
	// Get the list of epics from a epic's only query
	epicWitQuery := "0325c50f-3511-4266-a9fe-80b989492c76"
	if verbose {
		fmt.Printf("Fetching epics using query %v\n", epicWitQuery)
	}

	parentEpics, err := getEpics(acc, proj, token, epicWitQuery)
	if err != nil {
		return err
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
	return nil
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

	wits, err := q.RefreshWit(parentEpic, semesterFilter)
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

func showPrStats(acc, proj, token, repo string, count int, azStorageAcc, azStorageKey string) error {
	r := NewRepo(acc, proj, token, repo)
	if r.err != nil {
		return r.err
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

	fileName := "revstat.png"
	err := savePrStatImage(revStats, count, fileName)

	if err != nil {
		return err
	}

	if !noUpload {
		url, err := uploadImageToAzure(azStorageAcc, azStorageKey, fileName)
		if err != nil {
			return err
		}

		if verbose {
			fmt.Println("Uploaded to", url)
		}
	}

	return nil
}

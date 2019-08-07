package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
)

var showWork, showPr, verbose, noUpload bool
var prCount int
var semesterFilter bool
var poll bool

const pollDuration time.Duration = 5 * time.Minute

func main() {

	// Setup command line parsing
	flag.BoolVar(&showWork, "wit", false, "Show workitem stats")
	flag.IntVar(&prCount, "pr", 0, "Number of pull requests to process for count")
	flag.BoolVar(&verbose, "v", false, "Show verbose output")
	flag.BoolVar(&noUpload, "nu", false, "Do not upload generated data into Azure")
	flag.BoolVar(&semesterFilter, "sem", false, "Filter workitems not finished in this semester")
	flag.BoolVar(&poll, "poll", false, "Continue polling and updating")

	flag.Parse()

	// Fetch the access stuff from environment
	acc := os.Getenv("AZUREDEVOPS_ACCOUNT")
	proj := os.Getenv("AZUREDEVOPS_PROJECT")
	token := os.Getenv("AZUREDEVOPS_TOKEN")
	repo := os.Getenv("AZUREDEVOPS_REPO")
	azStorageAcc := os.Getenv("AZURE_STORAGE_ACCOUNT")
	azStorageKey := os.Getenv("AZURE_STORAGE_ACCESS_KEY")

	if len(acc) == 0 || len(proj) == 0 || len(token) == 0 || len(repo) == 0 ||
		len(azStorageAcc) == 0 || len(azStorageKey) == 0 {
		fmt.Println("Environment not setup")
		os.Exit(1)
	}

	exitCode := 0

	for true {
		if showWork {
			err := showWorkStats(acc, proj, token, azStorageAcc, azStorageKey)
			if err != nil {
				fmt.Println("Error fetching work stats!!", err)
				exitCode += 2
			}
		}

		if prCount > 0 {
			err := showPrStats(acc, proj, token, repo, prCount, azStorageAcc, azStorageKey)
			if err != nil {
				fmt.Println("Error fetching pull-request stats!!", err)
				exitCode += 4
			}
		}

		if poll {
			fmt.Println("Sleeping for ", pollDuration)
			time.Sleep(pollDuration)
		} else {
			break
		}
	}
	os.Exit(exitCode)
}

// ================================================================================================
// Workitem
func showWorkStats(acc, proj, token string, azStorageAcc, azStorageKey string) error {
	// Get the list of epics from a epic's only query
	epicWitQuery := "0325c50f-3511-4266-a9fe-80b989492c76"
	if verbose {
		fmt.Printf("Fetching epics using query %v\n", epicWitQuery)
	}

	parentEpics, err := getEpics(acc, proj, token, epicWitQuery)
	if err != nil {
		return err
	}

	var epicStats []EpicStat

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
			epicStat, err := getEpicStat(acc, proj, token, epic)
			m.Lock()
			if verbose {
				fmt.Print(".")
			}
			defer m.Unlock()
			if err != nil {
				fmt.Println("Error getting stat for epic", epic)
				return
			}

			epicStats = append(epicStats, epicStat)

		}(epic.Id, m)
	}

	wg.Wait()
	if verbose {
		fmt.Println() // to move to next line after the progress dots shown above
	}

	var maxBars float32 = 120.0
	var maxCount float32
	for _, e := range epicStats {
		count := float32(e.Done + e.InProgress + e.NotDone + e.Unknown)
		if count > maxCount {
			maxCount = count
		}
	}

	for _, e := range epicStats {
		fmt.Printf("%v: %v (%v)\n", e.Epic.Id, e.Epic.Title, e.Epic.AssignedTo)
		fmt.Printf("Done:%v InProgress:%v ToDo:%v Unknown:%v\n", e.Done, e.InProgress, e.NotDone, e.Unknown)
		conv := maxBars / maxCount
		drawBars("#", conv*float32(e.Done))
		drawBars("=", conv*float32(e.InProgress))
		drawBars("-", conv*float32(e.NotDone))
		drawBars(".", conv*float32(e.Unknown))
		fmt.Print("\n\n")
	}

	fileName := "epicstat.png"
	err = saveWitStatImage(epicStats, fileName)
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

func getEpics(acc, proj, token, queryID string) ([]WorkItem, error) {
	q := NewWork(acc, proj, token)
	epics, err := q.GetWorkitems(queryID)
	if err != nil {
		return nil, err
	}

	return epics, nil
}

func getEpicStat(acc, proj, token string, parentEpic int) (EpicStat, error) {
	q := NewWork(acc, proj, token)

	stats, err := q.RefreshWit(parentEpic, semesterFilter)

	return stats, err
}

func drawBars(ch string, count float32) {
	for i := 0; i < int(count); i++ {
		fmt.Print(ch)
	}
}

// ================================================================================================
// PR
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
		fmt.Printf("\nReviewer Stats\n\n")
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

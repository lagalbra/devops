package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Docs
// https://docs.microsoft.com/en-us/rest/api/azure/devops/git/pull%20requests/get%20pull%20requests?view=azure-devops-server-rest-4.1
func main() {
	// Fetch the access stuff from environment
	acc := os.Getenv("AZUREDEVOPS_ACCOUNT")
	proj := os.Getenv("AZUREDEVOPS_PROJECT")
	token := os.Getenv("AZUREDEVOPS_TOKEN")
	repo := os.Getenv("AZUREDEVOPS_REPO")

	// Connect to repo
	r := NewRepo(acc, proj, token, repo)
	if r.err != nil {
		fmt.Println(r.err)
		return
	}

	// Fetch PRs
	count := 200
	fmt.Printf("Processing %v completed PRs.........\n", count)
	r.Refresh(count)
	prs := r.PullRequests

	// Iterate and create a map of reviewers[review-count]
	m := make(map[string]int)
	for _, pr := range prs {
		for _, rv := range pr.Reviewers {
			// filter for specific user and ensure we do not count PR creater approving their own PR
			if !strings.Contains(rv.DisplayName, "AzLinux SAP HANA RP Devs") && rv.Vote != 0 && rv.DisplayName != pr.CreatedBy.DisplayName {
				m[rv.DisplayName]++
			}
		}
	}

	// Sort the PRs by review count, by stuffing into a slice
	type kv struct {
		Key   string
		Value int
	}

	max := 0
	var kvs []kv
	for k, v := range m {
		kvs = append(kvs, kv{k, v})
		if v > max {
			max = v
		}
	}

	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Value > kvs[j].Value
	})

	barmax := float32(100.0)
	// Output!!
	for _, kv := range kvs {
		bar := int((barmax / float32(max)) * float32(kv.Value))
		fmt.Printf("%30s %4d ", kv.Key, kv.Value)
		for i := 0; i < bar; i++ {
			fmt.Print("*")
		}

		fmt.Println()
	}
}

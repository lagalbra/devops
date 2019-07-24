package main

import (
	"fmt"
	"os"
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
	count := 400
	prStats, max := r.GetPullRequestReviewsByUser(count)
	barmax := float32(80.0)
	// Output!!
	for _, ps := range prStats {
		bar := int((barmax / float32(max)) * float32(ps.Count))
		percentage := float32(ps.Count) / float32(count) * 100.0
		fmt.Printf("%30s %4d (%4.1f%%) ", ps.Name, ps.Count, percentage)
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

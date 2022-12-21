package scrapper

import (
	"fmt"
	"strconv"

	"github.com/daichi-m/go18ds/sets/hashset"
	"github.com/gocolly/colly"
	workerpool "github.com/henrique-tavares/tcc_animmendation/dataset_builder/services/workerPool"
)

func GetAllUsersFromClubs(clubList []int) *hashset.Set[string] {
	parallel := 15
	usersSet := hashset.New[string]()

	jobs := make(chan int, parallel)
	results := make(chan *hashset.Set[string], parallel)

	workerPool := workerpool.New(parallel, jobs, results, getUsersFromClub)
	go workerPool.Run()

	go func() {
		for _, clubId := range clubList {
			jobs <- clubId
		}
		close(jobs)
	}()

	for result := range results {
		usersSet.Add(result.Values()...)
	}

	return usersSet
}

func getUsersFromClub(clubId int, log func(string)) *hashset.Set[string] {
	c := colly.NewCollector()
	baseURL := fmt.Sprintf("https://myanimelist.net/clubs.php?action=view&t=members&id=%v", clubId)
	usersSet := hashset.New[string]()
	usersPerPage := 36

	c.OnRequest(func(r *colly.Request) {
		log(fmt.Sprintf("Fetching club: '%v' showing from: '%v'", r.URL.Query().Get("id"), r.URL.Query().Get("show")))
	})

	c.OnHTML("td div:first-child a", func(e *colly.HTMLElement) {
		usersSet.Add(e.Text)
	})

	c.OnScraped(func(r *colly.Response) {
		showParam, _ := strconv.Atoi(r.Request.URL.Query().Get("show"))

		c.Visit(fmt.Sprintf("%v&show=%v", baseURL, showParam+usersPerPage))
	})

	c.Visit(fmt.Sprintf("%v&show=%v", baseURL, 0))

	return usersSet
}

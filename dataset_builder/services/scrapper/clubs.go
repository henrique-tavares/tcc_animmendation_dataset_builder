package scrapper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/barweiss/go-tuple"
	"github.com/daichi-m/go18ds/sets/hashset"
	"github.com/gocolly/colly"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func GetClubsId() *hashset.Set[int] {
	c := colly.NewCollector(colly.Async(true))
	baseURL := "https://myanimelist.net/clubs.php"
	parallel := 5
	clubQueue := make(chan tuple.T2[int, int])

	clubsId := hashset.New[int]()
	possibleUsers := 0

	c.OnRequest(func(r *colly.Request) {
		logrus.Infof("Fetching '%v'\n", r.URL.String())
	})

	c.OnHTML("tr.table-data", func(e *colly.HTMLElement) {
		rawMembers := e.ChildText("td:nth-of-type(2)")
		rawMembers = strings.ReplaceAll(rawMembers, ",", "")
		members, _ := strconv.Atoi(rawMembers)

		clubUrl := e.ChildAttr("a.fw-b", "href")
		rawClubId := strings.Split(clubUrl, "cid=")[1]
		clubId, _ := strconv.Atoi(rawClubId)

		lo.Try0(func() {
			clubQueue <- tuple.New2(members, clubId)
		})
	})

	c.OnScraped(func(r *colly.Response) {
		logrus.Infof("Fetched. Possible users: %v\n", possibleUsers)

		page, _ := strconv.Atoi(r.Request.URL.Query().Get("p"))

		if possibleUsers > 1_000_000 {
			lo.Try0(func() {
				close(clubQueue)
			})
			return
		}

		r.Request.Visit(fmt.Sprintf("%v?p=%d", baseURL, page+parallel))
	})

	for i := 0; i < parallel; i++ {
		c.Visit(fmt.Sprintf("%v?p=%d", baseURL, i+1))
	}

	for pair := range clubQueue {
		members := pair.V1
		clubId := pair.V2

		if members > 30 && !clubsId.Contains(clubId) {
			possibleUsers += pair.V1
			clubsId.Add(pair.V2)
		}
	}

	return clubsId
}

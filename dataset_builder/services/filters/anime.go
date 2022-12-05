package filters

import (
	"sort"
	"strconv"

	csvhandler "github.com/henrique-tavares/tcc_animmendation/dataset_builder/services/csvHandler"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
)

func FilterAnime(filePath string) error {
	errChan := make(chan error)

	animeMapChan := csvhandler.ReadBatched(filePath, errChan)
	filteredAnimes := []map[string]string{}

	for keepRunning := true; keepRunning; {
		select {
		case animeMap, ok := <-animeMapChan:
			if !ok {
				keepRunning = false
				break
			}

			if filter(animeMap) {
				filteredAnimes = append(filteredAnimes, animeMap)
			}
		case err := <-errChan:
			if err != nil {
				return err
			}
		}
	}

	sort.Slice(filteredAnimes, func(i, j int) bool {
		animeIdI, _ := strconv.Atoi(filteredAnimes[i]["id"])
		animeIdJ, _ := strconv.Atoi(filteredAnimes[j]["id"])

		return animeIdI < animeIdJ
	})

	header := []string{"id", "title", "score", "alternative_titles", "picture", "status", "media_type", "start_date", "end_date", "released_season", "genres", "synopsis", "age_classification", "source", "studios", "episodes", "rank", "popularity", "nsfw", "related_anime"}
	filteredAnimesCsv := lo.Map(filteredAnimes, func(animeMap map[string]string, _ int) []string {
		return lo.Map(header, func(field string, _ int) string {
			return animeMap[field]
		})
	})

	err := csvhandler.Write(filePath, header, filteredAnimesCsv)
	if err != nil {
		return err
	}

	return nil
}

func filter(animeMap map[string]string) bool {
	if animeMap["score"] == "0" {
		return false
	}

	if slices.Contains([]string{"music", "unknown"}, animeMap["media_type"]) {
		return false
	}

	if animeMap["status"] == "not_yet_aired" {
		return false
	}

	return true
}

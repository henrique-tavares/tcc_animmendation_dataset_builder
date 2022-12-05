package filters

import (
	"os"

	"github.com/daichi-m/go18ds/sets/hashset"
	csvhandler "github.com/henrique-tavares/tcc_animmendation/dataset_builder/services/csvHandler"
	"github.com/samber/lo"
)

func FilterUserAnimeRatings(filePath string, animePath string) error {
	allAnime, err := csvhandler.Read(animePath)
	if err != nil {
		return err
	}

	animeIdSet := hashset.New[string]()
	for _, anime := range allAnime {
		animeIdSet.Add(anime[0])
	}

	readErrChan := make(chan error)
	userAnimeRatingChan := csvhandler.ReadBatched(filePath, readErrChan)

	filteredUserAnimeRatingChan := make(chan [][]string)
	writeErrChan := make(chan error)
	header := []string{"user_id", "anime_id", "score"}
	go csvhandler.WriteBatched("data/temp.csv", header, filteredUserAnimeRatingChan, writeErrChan)

	go func() {
		for userAnimeRating := range userAnimeRatingChan {
			if !filterUserAnimeRating(userAnimeRating, animeIdSet) {
				continue
			}

			userAnimeRatingCsv := lo.Map(header, func(field string, _ int) string {
				return userAnimeRating[field]
			})

			ok := lo.Try0(func() {
				filteredUserAnimeRatingChan <- [][]string{userAnimeRatingCsv}
			})

			if !ok {
				return
			}
		}
		close(filteredUserAnimeRatingChan)
	}()

	errChans := lo.FanIn(1, readErrChan, writeErrChan)

	for err := range errChans {
		lo.Try0(func() {
			close(filteredUserAnimeRatingChan)
		})
		return err
	}

	err = os.Rename("data/temp.csv", filePath)
	if err != nil {
		return err
	}

	return nil
}

func filterUserAnimeRating(userAnimeRating map[string]string, animesId *hashset.Set[string]) bool {
	return animesId.Contains(userAnimeRating["anime_id"])
}

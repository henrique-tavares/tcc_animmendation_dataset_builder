package usecases

import (
	"fmt"
	"os"
	"strconv"

	"github.com/daichi-m/go18ds/sets/hashset"
	csvhandler "github.com/henrique-tavares/tcc_animmendation/dataset_builder/services/csvHandler"
	"github.com/henrique-tavares/tcc_animmendation/dataset_builder/services/scrapper"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func ScrapeClubs(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {

		logrus.Infof("File '%v' already exists!\n", filePath)
		return nil
	}

	clubsId := scrapper.GetClubsId()

	clubsIdRecords := lo.Map(clubsId.Values(), func(clubId int, _ int) []string {
		return []string{strconv.Itoa(clubId)}
	})

	csvhandler.Write(filePath, []string{"club_id"}, clubsIdRecords)

	return nil
}

func ScrapeUsers(filePath string, clubsPath string) error {
	if _, err := os.Stat(filePath); err == nil {
		logrus.Infof("File '%v' already exists!\n", filePath)
		return nil
	}

	clubsIdRecords, err := csvhandler.Read(clubsPath)
	if err != nil {
		return err
	}

	clubsId := lo.Map(clubsIdRecords, func(clubIdRecord []string, _ int) int {
		clubId, _ := strconv.Atoi(clubIdRecord[0])
		return clubId
	})

	usernames := scrapper.GetAllUsersFromClubs(clubsId)

	usernamesRecords := lo.Map(usernames.Values(), func(username string, _ int) []string {
		return []string{username}
	})

	csvhandler.Write(filePath, []string{"username"}, usernamesRecords)

	return nil
}

func ScrapeUserAnimeRatings(filePath string, usersPath string) error {
	if _, err := os.Stat(filePath); err == nil {
		logrus.Infof("File '%v' already exists!\n", filePath)
		return nil
	}

	usernamesRecords, err := csvhandler.Read(usersPath)
	if err != nil {
		return err
	}

	usernames := lo.Map(usernamesRecords, func(usernameRecord []string, i int) lo.Tuple2[int, string] {
		return lo.T2(i, usernameRecord[0])
	})

	userAnimeRatingChan := scrapper.GetAllUsersAnimeRating(usernames)
	userAnimeRatingRecordsChan := make(chan [][]string)
	errChan := make(chan error)

	go func() {
		for userAnimeRatingList := range userAnimeRatingChan {
			userAnimeRatingRecords := lo.Map(userAnimeRatingList, func(userAnimeRating scrapper.UserAnimeRating, _ int) []string {
				return []string{
					strconv.Itoa(userAnimeRating.UserId),
					strconv.Itoa(userAnimeRating.AnimeId),
					strconv.Itoa(userAnimeRating.Score),
				}
			})

			userAnimeRatingRecordsChan <- userAnimeRatingRecords
		}
		close(userAnimeRatingRecordsChan)
	}()

	go csvhandler.WriteBatched(filePath, []string{"user_id", "anime_id", "score"}, userAnimeRatingRecordsChan, errChan)

	for err := range errChan {
		return err
	}
	return nil
}

func ScrapeAnimes(filePath string, usersListPath string) error {
	if _, err := os.Stat(filePath); err == nil {
		logrus.Infof("File '%v' already exists!\n", filePath)
		return nil
	}

	userListErrChan := make(chan error)
	userRecordMapChan := csvhandler.ReadBatched(usersListPath, userListErrChan)

	animesIdSet := hashset.New[int]()

	for keepRunning := true; keepRunning; {
		select {
		case userRecordMap, ok := <-userRecordMapChan:
			if !ok {
				keepRunning = false
				break
			}

			animeId, err := strconv.Atoi(userRecordMap["anime_id"])
			if err != nil {
				logrus.Infof("userRecordMap: %v\n", userRecordMap)
				return err
			}

			animesIdSet.Add(animeId)

		case err := <-userListErrChan:
			if err != nil {
				return err
			}
		}
	}

	animeCsvChan := scrapper.GetAllAnime(animesIdSet.Values())
	animeRecordsChan := make(chan [][]string)
	animeErrChan := make(chan error)
	header := []string{
		"id", "title", "score", "alternative_titles", "picture", "status", "media_type", "start_date", "end_date", "released_season",
		"genres", "synopsis", "age_classification", "source", "studios", "episodes", "rank", "popularity", "nsfw", "related_anime",
	}

	go csvhandler.WriteBatched(filePath, header, animeRecordsChan, animeErrChan)

	go func() {
		for animeCsv := range animeCsvChan {
			empty := scrapper.AnimeCsv{}
			if animeCsv == empty {
				continue
			}
			animeArr := []string{
				strconv.Itoa(animeCsv.ID),
				animeCsv.Title,
				fmt.Sprintf("%v", animeCsv.Score),
				animeCsv.AlternativeTitles,
				animeCsv.Picture,
				animeCsv.Status,
				animeCsv.MediaType,
				animeCsv.StartDate,
				animeCsv.EndDate,
				animeCsv.ReleasedSeason,
				animeCsv.Genres,
				animeCsv.Synopsis,
				animeCsv.AgeClassification,
				animeCsv.Source,
				animeCsv.Studios,
				strconv.Itoa(animeCsv.Episodes),
				strconv.Itoa(animeCsv.Rank),
				strconv.Itoa(animeCsv.Popularity),
				animeCsv.Nsfw,
				animeCsv.RelatedAnime,
			}

			animeRecordsChan <- [][]string{animeArr}
		}
		close(animeRecordsChan)
	}()

	for err := range animeErrChan {
		return err
	}

	return nil
}

func GenerateAnimeAggregted(destFolder string, userAnimeRatingsPath string) error {
	err := os.MkdirAll(destFolder, 0744)
	if err != nil {
		return err
	}

	errChan := make(chan error)
	dataChan := csvhandler.ReadBatched(userAnimeRatingsPath, errChan)

	header := "user_id,score"
	for entry := range dataChan {
		filePath := fmt.Sprintf("%v/%v.csv", destFolder, entry["anime_id"])
		if _, err := os.Stat(filePath); err != nil {
			f, err := os.Create(filePath)
			if err != nil {
				return err
			}
			f.WriteString(fmt.Sprintln(header))
		}

		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0744)
		if err != nil {
			return err
		}
		f.WriteString(fmt.Sprintf("%v,%v\n", entry["user_id"], entry["score"]))
	}

	return nil
}

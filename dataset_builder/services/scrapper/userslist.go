package scrapper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	workerpool "github.com/henrique-tavares/tcc_animmendation/dataset_builder/services/workerPool"
	"github.com/samber/lo"
)

type UserAnimeListJson struct {
	Status  int `json:"status"`
	Score   int `json:"score"`
	AnimeId int `json:"anime_id"`
}

type UserAnimeRating struct {
	UserId  int
	AnimeId int
	Score   int
}

func GetAllUsersAnimeRating(users []lo.Tuple2[int, string]) <-chan []UserAnimeRating {
	parallel := 100

	jobs := make(chan lo.Tuple2[int, string])
	results := make(chan []UserAnimeRating)

	workerPool := workerpool.New(parallel, jobs, results, getUserAnimeListWrapper)
	go workerPool.Run()

	go func() {
		for _, user := range users {
			jobs <- user
		}
		close(jobs)
	}()

	return results
}

func getUserAnimeListWrapper(user lo.Tuple2[int, string], log func(string)) []UserAnimeRating {

	res, err := getUserAnimeList(user, log)
	if err != nil {
		log(fmt.Sprintf("error: %v", err))
		return []UserAnimeRating{}
	}

	return res
}

func getUserAnimeList(user lo.Tuple2[int, string], log func(string)) ([]UserAnimeRating, error) {
	userId, username := user.Unpack()
	animePerPage := 300

	malClientId := os.Getenv("X_MAL_CLIENT_ID")
	userAnimeRatingList := []UserAnimeRating{}

	for page := 0; ; {
		url := fmt.Sprintf("https://myanimelist.net/animelist/%v/load.json?offset=%v", username, page*animePerPage)
		log(fmt.Sprintf("Fetching user: %v (%v) '%v'", user.B, user.A, url))

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Add("X-MAL-CLIENT-ID", malClientId)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		rawBody, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		if res.StatusCode == 400 {
			return nil, fmt.Errorf("user %v's (%v) anime list is restricted", username, userId)

		} else if res.StatusCode != 200 {
			log(fmt.Sprintf("res.Status: %v | user: %v (%v)", res.Status, username, userId))
			time.Sleep(10 * time.Second)
			continue
		}

		var userAnimeList []UserAnimeListJson

		err = json.Unmarshal(rawBody, &userAnimeList)
		if err != nil {
			return nil, err
		}

		for _, animeRating := range userAnimeList {
			if animeRating.Score != 0 && animeRating.Status != 6 {
				userAnimeRatingList = append(userAnimeRatingList, UserAnimeRating{
					UserId:  userId,
					AnimeId: animeRating.AnimeId,
					Score:   animeRating.Score,
				})
			}
		}

		if len(userAnimeList) == 0 {
			break
		}

		page++
	}

	return userAnimeRatingList, nil
}

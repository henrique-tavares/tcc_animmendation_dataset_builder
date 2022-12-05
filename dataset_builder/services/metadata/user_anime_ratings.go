package metadata

import (
	"strconv"

	csvhandler "github.com/henrique-tavares/tcc_animmendation/dataset_builder/services/csvHandler"
)

func GetMaxUserId(filePath string) (int, error) {

	errChan := make(chan error)
	userAnimeRatingChan := csvhandler.ReadBatched(filePath, errChan)

	maxUserId := -1

	for keepRunning := true; keepRunning; {
		select {
		case userAnimeRating, ok := <-userAnimeRatingChan:
			if !ok {
				keepRunning = false
				break
			}

			userId, err := strconv.Atoi(userAnimeRating["user_id"])
			if err != nil {
				return 0, err
			}
			if userId > maxUserId {
				maxUserId = userId
			}
		case err := <-errChan:
			if err != nil {
				return 0, err
			}
		}
	}

	return maxUserId, nil
}

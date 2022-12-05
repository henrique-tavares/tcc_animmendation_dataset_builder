package usecases

import (
	"strconv"

	"github.com/henrique-tavares/tcc_animmendation/dataset_builder/services/metadata"
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

func GenerateMetadata(filePath string) error {
	metadataIni := ini.Empty()

	metadataAnimes, err := metadataIni.NewSection("animes")
	if err != nil {
		return err
	}

	totalAnimes, err := metadata.GetTotalAnime("data/animes.csv")
	if err != nil {
		return err
	}

	metadataAnimes.NewKey("total_animes", strconv.Itoa(totalAnimes))

	metadataUserAnimeRatings, err := metadataIni.NewSection("user_anime_ratings")
	if err != nil {
		return err
	}

	maxUserId, err := metadata.GetMaxUserId("data/user_anime_ratings.csv")
	if err != nil {
		return err
	}

	metadataUserAnimeRatings.NewKey("max_user_id", strconv.Itoa(maxUserId))

	metadataIni.SaveTo("data/metadata.ini")

	logrus.Infof("Metadata '%v' written", filePath)

	return nil
}

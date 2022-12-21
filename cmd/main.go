package main

import (
	"github.com/henrique-tavares/tcc_animmendation/dataset_builder/services/filters"
	"github.com/henrique-tavares/tcc_animmendation/dataset_builder/services/usecases"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	err := godotenv.Load()
	if err != nil {
		logrus.Fatalln(err)
	}

	err = usecases.ScrapeClubs("data/clubs.csv")
	if err != nil {
		logrus.Fatalln(err)
	}

	err = usecases.ScrapeUsers("data/users.csv", "data/clubs.csv")
	if err != nil {
		logrus.Fatalln(err)
	}

	err = usecases.ScrapeUserAnimeRatings("data/user_anime_ratings.csv", "data/users.csv")
	if err != nil {
		logrus.Fatalln(err)
	}

	err = usecases.ScrapeAnimes("data/animes.csv", "data/user_anime_ratings.csv")
	if err != nil {
		logrus.Fatalln(err)
	}

	err = filters.FilterAnime("data/animes.csv")
	if err != nil {
		logrus.Fatalln(err)
	}

	err = filters.FilterUserAnimeRatings("data/user_anime_ratings.csv", "data/animes.csv")
	if err != nil {
		logrus.Fatalln(err)
	}

	err = usecases.GenerateMetadata("data/metadata.ini")
	if err != nil {
		logrus.Fatalln(err)
	}

	err = usecases.GenerateAnimeAggregted("data/anime_ratings", "data/user_anime_ratings.csv")
	if err != nil {
		logrus.Fatalln(err)
	}
}

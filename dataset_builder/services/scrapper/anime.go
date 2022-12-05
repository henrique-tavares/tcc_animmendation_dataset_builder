package scrapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	workerpool "github.com/henrique-tavares/tcc_animmendation/dataset_builder/services/workerPool"
	"github.com/samber/lo"
)

type AnimeJson struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Picture struct {
		Medium string `json:"medium"`
		Large  string `json:"large"`
	} `json:"main_picture"`
	AlternativeTitles struct {
		Synonyms []string `json:"synonyms"`
		Ja       string   `json:"ja"`
	} `json:"alternative_titles"`
	StartDate   string      `json:"start_date"`
	EndDate     string      `json:"end_date"`
	Synopsis    string      `json:"synopsis"`
	Mean        float64     `json:"mean"`
	Rank        int         `json:"rank"`
	Popularity  int         `json:"popularity"`
	Nsfw        string      `json:"nsfw"`
	MediaType   string      `json:"media_type"`
	Status      string      `json:"status"`
	Genres      []genreJson `json:"genres"`
	NumEpisodes int         `json:"num_episodes"`
	StartSeason struct {
		Year   int    `json:"year"`
		Season string `json:"season"`
	} `json:"start_season"`
	Source       string         `json:"source"`
	Rating       string         `json:"rating"`
	Studios      []studioJson   `json:"studios"`
	RelatedAnime []relationJson `json:"related_anime"`
}

type genreJson struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type studioJson struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type relationJson struct {
	Node struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		MainPicture struct {
			Medium string `json:"medium"`
			Large  string `json:"large"`
		} `json:"main_picture"`
	} `json:"node"`
	RelationType          string `json:"relation_type"`
	RelationTypeFormatted string `json:"relation_type_formatted"`
}

type AnimeCsv struct {
	ID                int     `csv:"id"`
	Title             string  `csv:"title"`
	Picture           string  `csv:"picture"`
	AlternativeTitles string  `csv:"alternative_titles"`
	StartDate         string  `csv:"start_date"`
	EndDate           string  `csv:"end_date"`
	Synopsis          string  `csv:"synopsis"`
	Score             float64 `csv:"score"`
	Rank              int     `csv:"rank"`
	Popularity        int     `csv:"popularity"`
	Nsfw              string  `csv:"nsfw"`
	MediaType         string  `csv:"media_type"`
	Status            string  `csv:"status"`
	Genres            string  `csv:"genres"`
	Episodes          int     `csv:"episodes"`
	ReleasedSeason    string  `csv:"released_season"`
	Source            string  `csv:"source"`
	AgeClassification string  `csv:"age_classification"`
	Studios           string  `csv:"studios"`
	RelatedAnime      string  `csv:"related_anime"`
}

func GetAllAnime(animesId []int) <-chan AnimeCsv {
	parallel := 100

	jobs := make(chan int)
	results := make(chan AnimeCsv)

	workerPool := workerpool.New(parallel, jobs, results, getAnimeWrapper)
	go workerPool.Run()

	go func() {
		for _, animeId := range animesId {
			jobs <- animeId
		}
		close(jobs)
	}()

	return results
}

func getAnimeWrapper(animeId int, log func(string)) AnimeCsv {

	for {
		log(fmt.Sprintf("Fetching anime %v", animeId))
		animeJson, err := getAnime(animeId, log)
		if err != nil {
			delay := 5
			log(fmt.Sprintf("err: %v. Retrying in %v seconds\n", err, delay))
			time.Sleep(time.Duration(delay) * time.Second)
			continue
		}

		animeCsv := parseAnime(animeJson)

		return animeCsv
	}
}

func getAnime(animeId int, log func(string)) (AnimeJson, error) {
	animeUrl := fmt.Sprintf("https://api.myanimelist.net/v2/anime/%v?fields=id,title,main_picture,alternative_titles,start_date,end_date,synopsis,mean,rank,popularity,nsfw,,media_type,status,genres,num_episodes,start_season,source,rating,studios,related_anime", animeId)
	malClientId := os.Getenv("X_MAL_CLIENT_ID")

	proxuUrl, _ := url.Parse(os.Getenv("PROXY_URL"))

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxuUrl),
		},
	}

	var animeJson AnimeJson

	req, err := http.NewRequest(http.MethodGet, animeUrl, nil)
	if err != nil {
		return AnimeJson{}, err
	}

	req.Header.Add("X-MAL-CLIENT-ID", malClientId)

	res, err := client.Do(req)
	if err != nil {
		return AnimeJson{}, err
	}
	if res.StatusCode != 200 {
		return AnimeJson{}, fmt.Errorf("[%v] %v", animeId, res.Status)
	}

	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return AnimeJson{}, err
	}

	err = json.Unmarshal(rawBody, &animeJson)
	if err != nil {
		return AnimeJson{}, err
	}

	return animeJson, nil
}

func parseAnime(animeJson AnimeJson) AnimeCsv {
	parsedStartDate, _ := lo.TryOr(func() (string, error) {
		return parseDate(animeJson.StartDate)
	}, "")

	parsedEndDate, _ := lo.TryOr(func() (string, error) {
		return parseDate(animeJson.EndDate)
	}, "")

	parsedSeason, _ := lo.TryOr(func() (string, error) {
		return parseSeason(animeJson.StartSeason.Season, animeJson.StartSeason.Year)
	}, "")

	return AnimeCsv{
		ID:                animeJson.ID,
		Title:             animeJson.Title,
		Picture:           animeJson.Picture.Large,
		AlternativeTitles: parseArray(append([]string{animeJson.AlternativeTitles.Ja}, animeJson.AlternativeTitles.Synonyms...)),
		StartDate:         parsedStartDate,
		EndDate:           parsedEndDate,
		Synopsis:          handleText(animeJson.Synopsis),
		Score:             animeJson.Mean,
		Rank:              animeJson.Rank,
		Popularity:        animeJson.Popularity,
		Nsfw:              animeJson.Nsfw,
		MediaType:         animeJson.MediaType,
		Status:            animeJson.Status,
		Genres: parseArray(lo.Map(animeJson.Genres, func(genre genreJson, _ int) string {
			return genre.Name
		})),
		Episodes:          animeJson.NumEpisodes,
		ReleasedSeason:    parsedSeason,
		Source:            animeJson.Source,
		AgeClassification: animeJson.Rating,
		Studios: parseArray(lo.Map(animeJson.Studios, func(studio studioJson, _ int) string {
			return studio.Name
		})),
		RelatedAnime: parseArray(lo.Map(animeJson.RelatedAnime, func(relation relationJson, _ int) string {
			return fmt.Sprintf("%v %v", relation.RelationType, relation.Node.ID)
		})),
	}
}

func parseArray(arr []string) string {
	return fmt.Sprintf("{%v}", strings.Join(lo.FilterMap(arr, func(item string, _ int) (string, bool) {
		item = strings.ReplaceAll(item, ",", "\\,")
		item = strings.ReplaceAll(item, "{", "\\{")
		item = strings.ReplaceAll(item, "}", "\\}")
		item = strings.ReplaceAll(item, "\"", "@@")
		return item, item != ""
	}), ","))
}

func parseDate(date string) (string, error) {
	dateSplitted, hasErr := lo.TryOr(func() ([]int, error) {
		return lo.Map(strings.Split(date, "-"), func(datePiece string, _ int) int {
			num, err := strconv.Atoi(datePiece)
			if err != nil {
				panic(err)
			}

			return num
		}), nil
	}, []int{})

	if hasErr {
		return "", errors.New("invalid date")
	}

	switch len(dateSplitted) {
	case 3:
		return fmt.Sprintf("%v-%v-%v", dateSplitted[0], dateSplitted[1], dateSplitted[2]), nil
	case 2:
		return fmt.Sprintf("%v-%v-01", dateSplitted[0], dateSplitted[1]), nil
	case 1:
		return fmt.Sprintf("%v-01-01", dateSplitted[0]), nil
	default:
		return "", errors.New("invalid date")
	}
}

func parseSeason(season string, year int) (string, error) {
	if season == "" || year == 0 {
		return "", errors.New("invalid season or year")
	}

	return fmt.Sprintf("%v, %v", season, year), nil
}

func handleText(text string) string {
	text = strings.ReplaceAll(text, "\n\n", " ")
	text = strings.ReplaceAll(text, "\n", "")
	return text
}

func RewriteQuotesInArrayLiteral(filePath string) {

}

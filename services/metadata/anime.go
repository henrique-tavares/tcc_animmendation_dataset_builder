package metadata

import csvhandler "github.com/henrique-tavares/tcc_animmendation/dataset_builder/services/csvHandler"

func GetTotalAnime(filePath string) (int, error) {
	allAnime, err := csvhandler.Read(filePath)
	if err != nil {
		return 0, err
	}

	return len(allAnime), nil
}

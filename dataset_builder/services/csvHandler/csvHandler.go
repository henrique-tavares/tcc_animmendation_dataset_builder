package csvhandler

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func Write(filePath string, header []string, data [][]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	err = writer.Write(header)
	if err != nil {
		return err
	}

	err = writer.WriteAll(data)
	if err != nil {
		return err
	}

	logrus.Infof("Wrote to '%v'\n", filePath)

	return nil
}

func WriteBatched(filePath string, header []string, chunks <-chan [][]string, errChan chan<- error) {
	file, err := os.Create(filePath)
	if err != nil {
		errChan <- err
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	err = writer.Write(header)
	if err != nil {
		errChan <- err
		return
	}

	for chunk := range chunks {
		err = writer.WriteAll(chunk)
		if err != nil {
			errChan <- err
			return
		}
	}

	logrus.Infof("Wrote to '%v'\n", filePath)
	close(errChan)
}

func Read(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ReadBatched(filePath string, errChan chan<- error) <-chan map[string]string {
	file, err := os.Open(filePath)
	if err != nil {
		errChan <- err
		close(errChan)
		return nil
	}

	reader := csv.NewReader(file)
	recordMapChan := make(chan map[string]string)

	header, err := reader.Read()
	if err != nil {
		errChan <- err
		close(errChan)
		return nil
	}

	go func() {
		defer file.Close()
		for {
			record, err := reader.Read()
			if record == nil && err == io.EOF {
				break
			}
			if err != nil {
				errChan <- err
				break
			}

			userListMap := map[string]string{}
			for i := 0; i < len(header); i++ {
				userListMap[header[i]] = record[i]
			}

			recordMapChan <- userListMap

		}
		close(recordMapChan)
		close(errChan)
	}()

	return recordMapChan
}

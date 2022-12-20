package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

func main() {
	date := "2019-09-18"
	dateSplitted, hasErr := lo.TryOr(func() ([]int, error) {
		return lo.Map(strings.Split(date, "-"), func(datePiece string, _ int) int {
			num, err := strconv.Atoi(datePiece)
			fmt.Println(num, err)
			if err != nil {
				panic(err)
			}

			return num
		}), nil
	}, []int{})

	fmt.Println(date, dateSplitted, hasErr)

}

func parseDate(date string) (string, error) {
	dateSplitted, ok := lo.TryOr(func() ([]int, error) {
		return lo.Map(strings.Split(date, "-"), func(datePiece string, _ int) int {
			num, err := strconv.Atoi(datePiece)
			if err != nil {
				panic(err)
			}

			return num
		}), nil
	}, []int{})

	if !ok {
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

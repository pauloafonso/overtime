package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func openFile(pathFile string) *os.File {
	file, err := os.Open(pathFile)
	check(err)
	return file
}

func breakRows(file *os.File) []string {
	reader := bufio.NewReader(file)
	rows := make([]string, 0)
	for {
		row, err := reader.ReadString('\n')
		rows = append(rows, strings.TrimSpace(row))
		if err == io.EOF {
			break
		}
	}
	file.Close()
	return rows
}

type validatedRow struct {
	content string
}

func validateRows(rows []string) []validatedRow {
	validatedRows := make([]validatedRow, 0)
	for _, element := range rows {
		match, _ := regexp.MatchString("^[0-9]{2}/[0-9]{2}/[0-9]{4} [0-9]{2}:[0-9]{2}-[0-9]{2}:[0-9]{2}$", element)
		if match == false {
			fmt.Println("data/agosto file not validated by regular expressions.")
			os.Exit(-1)
		}
		validatedRows = append(validatedRows, validatedRow{content: element})
	}
	return validatedRows
}

func captureDate(str string) string {
	matchDate, _ := regexp.Compile("[0-9]{2}/[0-9]{2}/[0-9]{4}")
	return matchDate.FindString(str)
}

func captureInitialHour(str string) string {
	initialHour, _ := regexp.Compile(" [0-9]{2}:[0-9]{2}")
	return strings.Replace(initialHour.FindString(str), " ", "", 1)
}

func captureFinalHour(str string) string {
	finalHour, _ := regexp.Compile("-[0-9]{2}:[0-9]{2}")
	return strings.Replace(finalHour.FindString(str), "-", "", 1)
}

func calculateOvertime(validatedRows []validatedRow) {
	for _, element := range validatedRows {
		date := captureDate(element.content)
		initialHour := captureInitialHour(element.content)
		finalHour := captureFinalHour(element.content)
		fmt.Println(date)
		fmt.Println(initialHour)
		fmt.Println(finalHour)
	}
}

func main() {
	file := openFile("data/agosto")
	rows := breakRows(file)
	validatedRows := validateRows(rows)
	calculateOvertime(validatedRows)
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
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

func parseRowToDateTime(element validatedRow) rangeDateTime {
	date := captureDate(element.content)
	initialHour := captureInitialHour(element.content)
	finalHour := captureFinalHour(element.content)
	format := "02/01/2006 15:04"
	initialParsed, _ := time.Parse(format, date+" "+initialHour)
	finalParsed, _ := time.Parse(format, date+" "+finalHour)
	return rangeDateTime{initial: initialParsed, final: finalParsed}
}

func calculate(validatedRows []validatedRow) []rangeDateTime {
	sliceDateTimes := make([]rangeDateTime, 0)
	for _, element := range validatedRows {
		rangeDateTime := parseRowToDateTime(element)
		diffPerDay := calculateDiffPerDay(rangeDateTime)
		time50Time100PerDay := calculateTime50Time100PerDay(diffPerDay)
		fmt.Println(time50Time100PerDay.day, time50Time100PerDay.time50, time50Time100PerDay.time100)
	}
	return sliceDateTimes
}

func calculateTime50Time100PerDay(d diffPerDay) time50Time100PerDay {
	time50 := d.diff
	time100 := 0.00
	if d.diff > 200 {
		time50 = 200.00
		time100 = d.diff - 200.00
	}
	return time50Time100PerDay{day: d.day, time50: time50, time100: time100}
}

type resultDay struct {
	day     time.Time
	initial time.Time
	final   time.Time
	diff    float64
	time50  float64
	time100 float64
}

type rangeDateTime struct {
	initial time.Time
	final   time.Time
}

func calculateDiffPerDay(r rangeDateTime) diffPerDay {
	diff := r.final.Sub(r.initial).Minutes()
	return diffPerDay{day: r.final, diff: diff}
}

type diffPerDay struct {
	day  time.Time
	diff float64
}

type time50Time100PerDay struct {
	day     time.Time
	time50  float64
	time100 float64
}

func main() {
	file := openFile("data/agosto")
	rows := breakRows(file)
	validatedRows := validateRows(rows)
	calculate(validatedRows)
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

const dateTimeFormat = "02/01/2006 15:04"
const dirOverTimeData = "data"

type durationPerDay struct {
	day     string
	minutes float64
}

type validatedRow struct {
	content string
}

type additionalNight struct {
	day     string
	minutes float64
}

type time50Time100PerDay struct {
	day     string
	time50  float64
	time100 float64
}

type dateTimeInterval struct {
	initial time.Time
	final   time.Time
}

func (d *durationPerDay) addMinutes(diff float64) {
	d.minutes = d.minutes + diff
}

func (d durationPerDay) calculateTime50Time100PerDay() time50Time100PerDay {
	time50 := d.minutes
	time100 := 0.00
	if d.minutes > 120 {
		time50 = 120.00
		time100 = d.minutes - 120.00
	}
	return time50Time100PerDay{day: d.day, time50: time50, time100: time100}
}

func (a *additionalNight) addAdditionalMinutes(additional float64) {
	a.minutes = a.minutes + additional
}

func (d dateTimeInterval) getSpecificTimeFromInitialDay(stime string) time.Time {
	t, _ := time.Parse(dateTimeFormat, d.initial.Format("02/01/2006 ")+stime)
	return t
}

func (d dateTimeInterval) getCurrentDay() string {
	// if the initial time is between 06:00 and 00:00, the day of overtime is the same
	if d.initial.Before(d.getSpecificTimeFromInitialDay("00:00")) && d.initial.After(d.getSpecificTimeFromInitialDay("06:00")) {
		return d.initial.AddDate(0, 0, -1).Format("02/01/2006")
	}
	// if the initial time is between 00:00 and 03:00, the day of overtime is one before
	return d.initial.Format("02/01/2006")
}

func main() {
	files := getOverTimeFiles()

	for _, f := range files {
		file := openFile(dirOverTimeData + "/" + f.Name())
		rows := breakRows(file)
		validatedRows := validateRows(rows)
		dateTimeIntervals := parseRowsToDateTimeIntervals(validatedRows)
		durationsPerDay := calculateDurationPerDay(dateTimeIntervals)
		additionsNight := calculateAdditionalNight(dateTimeIntervals)
		calculateResult(durationsPerDay, additionsNight).makeDataResult()
	}
}

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

func validateRows(rows []string) []validatedRow {
	validatedRows := make([]validatedRow, 0)
	for _, element := range rows {
		match, _ := regexp.MatchString(
			"^[0-9]{2}/[0-9]{2}/[0-9]{4} [0-9]{2}:[0-9]{2}-[0-9]{2}/[0-9]{2}/[0-9]{4} [0-9]{2}:[0-9]{2}$",
			element)
		if match == false {
			fmt.Println("data/agosto file not validated by regular expressions.")
			os.Exit(-1)
		}
		validatedRows = append(validatedRows, validatedRow{content: element})
	}
	return validatedRows
}

func matchInitialDateTime(str string) string {
	matchDate, _ := regexp.Compile("^[0-9]{2}/[0-9]{2}/[0-9]{4} [0-9]{2}:[0-9]{2}")
	return matchDate.FindString(str)
}

func matchFinalDateTime(str string) string {
	matchDate, _ := regexp.Compile("[0-9]{2}/[0-9]{2}/[0-9]{4} [0-9]{2}:[0-9]{2}$")
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

func parseRowsToDateTimeIntervals(validatedRows []validatedRow) []dateTimeInterval {
	dateTimeIntervals := make([]dateTimeInterval, 0)
	for _, element := range validatedRows {
		initialDateTime := matchInitialDateTime(element.content)
		finalDateTime := matchFinalDateTime(element.content)
		initial, _ := time.Parse(dateTimeFormat, initialDateTime)
		final, _ := time.Parse(dateTimeFormat, finalDateTime)
		dateTimeIntervals = append(dateTimeIntervals, dateTimeInterval{initial: initial, final: final})
	}
	return dateTimeIntervals
}

func p(a ...interface{}) (n int, err error) {
	return fmt.Println(a...)
}

func calculateDurationPerDay(dateTimeIntervals []dateTimeInterval) []*durationPerDay {
	durationsPerDay := make([]*durationPerDay, 0)
	for _, dateTimeInterval := range dateTimeIntervals {
		// calculate the difference duration in minutes
		diff := dateTimeInterval.final.Sub(dateTimeInterval.initial).Minutes()

		day := dateTimeInterval.getCurrentDay()

		dayExists := false
		for _, d := range durationsPerDay {
			if d.day == day {
				d.addMinutes(diff)
				dayExists = true
				break
			}
		}
		if dayExists == false {
			// i create my durationPerDay var and what i must append to the slice is its pointer to atualize it before in the nexts iterations
			durationsPerDay = append(durationsPerDay, &durationPerDay{day: day, minutes: diff})
		}
	}
	return durationsPerDay
}

func calculateAdditionalNight(dateTimeIntervals []dateTimeInterval) []*additionalNight {
	additionalsNight := make([]*additionalNight, 0)
	for _, dateTimeInterval := range dateTimeIntervals {
		day := dateTimeInterval.getCurrentDay()

		beginAddNight := dateTimeInterval.getSpecificTimeFromInitialDay("22:00")

		// (1) if the initial hour is after the 10pm, the whole duration has addNight
		if dateTimeInterval.initial.After(beginAddNight) {
			additional := dateTimeInterval.final.Sub(dateTimeInterval.initial).Minutes()

			dayExists := false
			for _, a := range additionalsNight {
				if a.day == day {
					dayExists = true
					a.addAdditionalMinutes(additional)
					break
				}
			}
			if dayExists == false {
				additionalsNight = append(additionalsNight, &additionalNight{day: day, minutes: additional})
			}
			continue
		}

		// if the final hour is after the 10pm, there is addNight on duration between 10pm and the final hour
		if dateTimeInterval.final.After(beginAddNight) {
			additional := dateTimeInterval.final.Sub(beginAddNight).Minutes()

			dayExists := false
			for _, a := range additionalsNight {
				if a.day == day {
					dayExists = true
					a.addAdditionalMinutes(additional)
					break
				}
			}
			if dayExists == false {
				additionalsNight = append(additionalsNight, &additionalNight{day: day, minutes: additional})
			}
			continue
		}
	}
	return additionalsNight
}

func getOverTimeFiles() []os.FileInfo {
	files, err := ioutil.ReadDir(dirOverTimeData)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

type dayResult struct {
	day             string
	minutes         float64
	time50          float64
	time100         float64
	additionalNight float64
}

type generalResult struct {
	days []dayResult
}

func calculateResult(durationsPerDay []*durationPerDay, additionsNight []*additionalNight) generalResult {
	days := make([]dayResult, 0)
	for _, d := range durationsPerDay {
		dr := &dayResult{
			day:     d.day,
			minutes: d.minutes,
			time50:  d.calculateTime50Time100PerDay().time50,
			time100: d.calculateTime50Time100PerDay().time100,
		}
		for _, a := range additionsNight {
			if a.day == d.day {
				dr.additionalNight = a.minutes
				break
			}
		}
		days = append(days, *dr)
	}
	return generalResult{days: days}
}

func (r generalResult) makeDataResult() {
	p(r)
}

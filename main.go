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

const dateTimeFormat = "02/01/2006 15:04"

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

type additionalNight struct {
	minutes int
}

type resultDay struct {
	day     time.Time
	initial time.Time
	final   time.Time
	diff    float64
	time50  float64
	time100 float64
}

func calculate(validatedRows []validatedRow) []dateTimeInterval {
	sliceDateTimes := make([]dateTimeInterval, 0)
	//for _, element := range validatedRows {

	//durationPerDay := calculateDurationPerDay(dateTimeInterval)
	//
	//dateTimeInterval.calculateAdditionalNight()
	// fmt.Println(time50Time100PerDay.day, time50Time100PerDay.time50, time50Time100PerDay.time100, additionalNight.minutes)
	//}
	return sliceDateTimes
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

type dateTimeInterval struct {
	initial time.Time
	final   time.Time
}

func calculateDurationPerDay(dateTimeIntervals []dateTimeInterval) []*durationPerDay {
	durationsPerDay := make([]*durationPerDay, 0)
	var day string
	for _, dateTimeInterval := range dateTimeIntervals {
		// calculate the difference duration in minutes
		diff := dateTimeInterval.final.Sub(dateTimeInterval.initial).Minutes()

		midNight, _ := time.Parse(dateTimeFormat, dateTimeInterval.initial.Format("02/01/2006 ")+"00:00")
		sixAm, _ := time.Parse(dateTimeFormat, dateTimeInterval.initial.Format("02/01/2006 ")+"06:00")

		// if the initial time is between 06:00 and 00:00, the day of overtime is the same
		// if the initial time is between 00:00 and 03:00, the day of overtime is one before
		if dateTimeInterval.initial.Before(midNight) && dateTimeInterval.initial.After(sixAm) {
			day = dateTimeInterval.initial.AddDate(0, 0, -1).Format("02/01/2006")
		} else {
			day = dateTimeInterval.initial.Format("02/01/2006")
		}

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
			d := durationPerDay{day: day, minutes: diff}
			var dPoint *durationPerDay = &d
			durationsPerDay = append(durationsPerDay, dPoint)
			// this is the same: durationsPerDay = append(durationsPerDay, &d)
		}
	}
	return durationsPerDay
}

func (d *durationPerDay) addMinutes(diff float64) {
	d.minutes = d.minutes + diff
}

type time50Time100PerDay struct {
	day     string
	time50  float64
	time100 float64
}

func (d durationPerDay) calculateTime50Time100PerDay() time50Time100PerDay {
	time50 := d.minutes
	time100 := 0.00
	if d.minutes > 200 {
		time50 = 200.00
		time100 = d.minutes - 200.00
	}
	return time50Time100PerDay{day: d.day, time50: time50, time100: time100}
}

func (r dateTimeInterval) calculateAdditionalNight() {
	beginAddNight, _ := time.Parse(dateTimeFormat, r.initial.Format("02/01/2006 ")+"22:00")
	midNight, _ := time.Parse(dateTimeFormat, r.initial.Format("02/01/2006 ")+"00:00")
	endAddNight, _ := time.Parse(dateTimeFormat, r.initial.Format("02/01/2006 ")+"05:00")

	// if the initial hour is after the 22pm, the whole duration has addNight
	if r.initial.After(beginAddNight) {
		r.final.Sub(r.initial).Minutes()
		return
	}

	// if the final hour is after the 22pm, there is addNight on duration between 22pm and the final hour
	if r.final.After(beginAddNight) {
		r.final.Sub(beginAddNight).Minutes()
		return
	}

	// if the final hour is after the 00am and before endAddNight (5am), there is addNight on duration between 00am and the final hour
	if r.final.After(midNight) && r.final.Before(endAddNight) {
		r.final.Sub(midNight).Minutes()
		return
	}

	//fmt.Println(r.initial.Before(beginAddNight), r.initial, beginAddNight)
}

type durationPerDay struct {
	day     string
	minutes float64
}

func main() {
	file := openFile("data/agosto")
	rows := breakRows(file)
	validatedRows := validateRows(rows)
	dateTimeIntervals := parseRowsToDateTimeIntervals(validatedRows)
	durationsPerDay := calculateDurationPerDay(dateTimeIntervals)
	p(durationsPerDay[0].calculateTime50Time100PerDay().time50)
	p(durationsPerDay[0].calculateTime50Time100PerDay().time100)
	//fmt.Println(durationPerDay)
	// calculate(validatedRows)
}

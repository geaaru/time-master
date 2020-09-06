/*

Copyright (C) 2020  Daniele Rondina <geaaru@sabayonlinux.org>
Credits goes also to Gogs authors, some code portions and re-implemented design
are also coming from the Gogs project, which is using the go-macaron framework
and was really source of ispiration. Kudos to them!

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/
package time

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	date "github.com/rickb777/date"
)

// Parse duration and return number of seconds
func ParseDuration(duration string, workHour int) (int64, error) {
	var ans int64 = -1
	var regexDays = regexp.MustCompile(`[0-9]*d$`)

	if regexDays.MatchString(duration) {

		duration = strings.ReplaceAll(duration, "d", "")

		if strings.Contains(duration, ".") {
			days, err := strconv.ParseFloat(duration, 64)
			if err != nil {
				return -1, err
			}

			ans = int64(days * float64(workHour) * 60 * 60)
		} else {
			days, err := strconv.ParseInt(duration, 10, 64)
			if err != nil {
				return -1, err
			}

			ans = days * int64(workHour) * 60 * 60

		}
	} else {

		m, err := time.ParseDuration(duration)
		if err != nil {
			return -1, err
		}
		ans = int64(m.Seconds())

	}

	return ans, nil
}

func Seconds2Duration(sec int64) (string, error) {

	if sec <= 0 {
		return "", errors.New("Seconds must be greather then 0")
	}

	m, err := time.ParseDuration(fmt.Sprintf("%ds", sec))
	if err != nil {
		return "", err
	}

	ans := ""
	if m.Hours() > 0 {
		ans = fmt.Sprintf("%dh", int64(m.Hours()))
	}

	if m.Minutes() > 0 {
		mm := int64(m.Minutes()) % 60

		if mm > 0 {
			ans = ans + fmt.Sprintf("%dm", int64(mm))
		}
	}

	if m.Seconds() > 0 {
		s := int64(m.Seconds()) % 60

		if s > 0 {
			ans = ans + fmt.Sprintf("%ds", int64(s))
		}
	}
	return ans, nil
}

func Seconds2Date(sec int64, onlyDate bool) (string, error) {
	ans := ""

	if sec <= 0 {
		return ans, errors.New("Seconds must be greather then 0")
	}

	m := time.Unix(sec, int64(0))

	if onlyDate {
		ans = m.Format("2006-01-02")
	} else {
		ans = m.Format("2006-01-02 15:04:05")
	}

	return ans, nil
}

func ParseTimestamp(t string, onlyDate bool) (time.Time, error) {
	var layout string

	if onlyDate {
		// PRE: The first word MUST YYYY-MM-DD
		words := strings.Split(t, " ")
		t = words[0]
		layout = "2006-01-02"
	} else {
		layout = "2006-01-02 15:04:00"
	}

	return time.Parse(layout, t)
}

func IsAWorkDay(dstr string) (bool, error) {
	d, err := date.ParseISO(dstr)
	if err != nil {
		return false, err
	}

	switch d.Weekday() {
	case time.Saturday, time.Sunday:
		return false, nil
	default:
		return true, nil
	}
}

func GetNextWorkDay(dstr string) (string, error) {
	var nextDay date.Date

	d, err := date.ParseISO(dstr)
	if err != nil {
		return "", err
	}

	switch d.Weekday() {
	case time.Friday:
		nextDay = d.AddDate(0, 0, 3)
	case time.Saturday:
		nextDay = d.AddDate(0, 0, 2)
	default:
		nextDay = d.AddDate(0, 0, 1)
	}

	return fmt.Sprintf("%d-%02d-%02d",
		nextDay.Year(), nextDay.Month(), nextDay.Day()), nil
}

func GetNextWeekFirstWorkDay(dstr string) (string, error) {
	var nextDay date.Date

	d, err := date.ParseISO(dstr)
	if err != nil {
		return "", err
	}

	switch d.Weekday() {
	case time.Sunday:
		nextDay = d.AddDate(0, 0, 1)
	default:
		nextDay = d.AddDate(0, 0, 7-int(d.Weekday())+1)
	}

	return fmt.Sprintf("%d-%02d-%02d",
		nextDay.Year(), nextDay.Month(), nextDay.Day()), nil
}

func GetNextMonthFirstWorkDay(dstr string) (string, error) {
	var nextDay date.Date

	d, err := date.ParseISO(dstr)
	if err != nil {
		return "", err
	}

	nextDay = d.AddDate(0, 1, 0)
	// Set first of the month
	nextDay, err = date.ParseISO(fmt.Sprintf("%d-%02d-%02d",
		nextDay.Year(),
		nextDay.Month(),
		1,
	))

	switch nextDay.Weekday() {
	case time.Sunday, time.Saturday:
		return GetNextWorkDay(
			fmt.Sprintf("%d-%02d-%02d",
				nextDay.Year(), nextDay.Month(), nextDay.Day()))
	}

	return fmt.Sprintf("%d-%02d-%02d",
		nextDay.Year(), nextDay.Month(), nextDay.Day()), nil
}

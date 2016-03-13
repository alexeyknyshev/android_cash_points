package main

import (
	"encoding/json"
// 	"fmt"
	"testing"
	"reflect"
	"github.com/alexeyknyshev/gojsondiff"
	"github.com/alexeyknyshev/gojsondiff/formatter"
)

func TestSplitTime(t *testing.T) {
	checkSplitTime := func(s string, expFrom, expTo int, expOk bool) {
		from, to, ok := SplitTime(s)
		if ok != expOk {
			t.Errorf("Expected '%v' but got '%v' => %s", expOk, ok, s)
		}

		if from != expFrom {
			t.Errorf("Expected from '%d' but got '%d' => %s", expFrom, from, s)
		}

		if to != expTo {
			t.Errorf("Expected to '%d' but got '%d' => %s", expTo, to, s)
		}
	}

	checkSplitTime("10:00-10:01", 600, 601, true)
	checkSplitTime("00:00-23:59", 0, 1439, true)
	checkSplitTime("15:30-20:30", 930, 1230, true)
	checkSplitTime("08:30—21:10", 510, 1270, true)
	checkSplitTime("01:31-01:31", 91, 91, true)
	checkSplitTime("blah-blah", 0, 0, false)
	checkSplitTime("20:30—20:30", 1230, 1230, true)
	checkSplitTime("20:15-03:40", 1215, 220, true)
}

func TestDayRange(t *testing.T) {
	checkDayRange := func(s string, expRange []int, expOk bool) {
		result, err := ParseDayRange(s)
		if expOk && err != nil {
			t.Errorf("Expected right day range %v but got wrong %v => %v", expRange, result, err)
		}

		if !reflect.DeepEqual(result, expRange) {
			t.Errorf("Diff day ranges. Expected %v got %v", expRange, result)
		}
	}

	checkDayRange("пн.", []int{ 0 }, true)
	checkDayRange("сб.", []int{ 5 }, true)
	checkDayRange("пн.-пт.", []int{ 0, 1, 2, 3, 4 }, true)
	checkDayRange("чт.-сб.", []int{ 3, 4, 5 }, true)
	checkDayRange("сб.-вс.", []int{ 5, 6 }, true)
	checkDayRange("пн.,ср.,пт.", []int{ 0, 2, 4 }, true)
	checkDayRange("вт.,чт.-сб.", []int{ 1, 3, 4, 5 }, true)
	checkDayRange("вт.—ср.", []int{ 1, 2 }, true)
	checkDayRange("пн.,ср.,чт.,вс.,вт.,пт.-сб.", []int{ 0, 2, 3, 6, 1, 4, 5 }, true)
	checkDayRange("пон.", []int{ }, false)
	checkDayRange("пн.-пн.", []int{ }, false)
	checkDayRange("чт.-вт.", []int{ }, false)
}

func TestSchedule(t *testing.T) {
	checkSchedule := func(s string, expSchedule *Schedule) {
		schedule, err := ParseSchedule(s)
		if err != nil {
			t.Errorf("Cannot parse schedule: %v => %s", err, s)
		}

		scheduleJson, _  := json.Marshal(schedule)
		expScheduleJson, _ := json.Marshal(expSchedule)

		differ := gojsondiff.New()

		conf := &gojsondiff.CompareConfig{FloatEpsilon: 0.0001}
		d, err := differ.Compare(expScheduleJson, scheduleJson, conf)
		if err != nil {
			t.Errorf("Cannot compare json data: %v", err)
			return
		}

		if !d.Modified() {
			return
		}

		var expectedJson map[string]interface{}
		json.Unmarshal(expScheduleJson, &expectedJson)
		formatter := formatter.NewAsciiFormatter(expectedJson)
		formatter.ShowArrayIndex = true
		diffString, err := formatter.Format(d)
		if err != nil {
			// No error can occur
		}

		t.Errorf("json diff:\n%s", diffString)
	}

	checkSchedule("пн.: 08:30-21:00", &Schedule{
		Mon: &ScheduleDay{
			From: 510,
			To:   1260,
		},
	})

	sDay := &ScheduleDay{}

	sDay.From = 540
	sDay.To = 1110
	checkSchedule("пн.-пт.: 09:00-18:30", &Schedule{
		Mon: sDay,
		Tue: sDay,
		Wed: sDay,
		Thu: sDay,
		Fri: sDay,
	})

	sDay.From = 615
	sDay.To = 870
	checkSchedule("пн.,ср.-сб.: 10:15-14:30", &Schedule{
		Mon: sDay,
		Wed: sDay,
		Thu: sDay,
		Fri: sDay,
		Sat: sDay,
	})

	checkSchedule("пн.: 10:30-17:00\nср.: 10:30-18:00", &Schedule{
		Mon: &ScheduleDay{
			From: 630,
			To:   1020,
		},
		Wed: &ScheduleDay{
			From: 630,
			To:   1080,
		},
	})

	sDay.From = 720
	sDay.To = 1260
	checkSchedule("пн.,ср.,пт.: 12:00-21:00\nсб.: 15:00-16:00", &Schedule{
		Mon: sDay,
		Wed: sDay,
		Fri: sDay,
		Sat: &ScheduleDay{
			From: 900,
			To:   960,
		},
	})

	sDay.From = 600
	sDay.To = 1080
	checkSchedule("вт.: 11:00-18:40\n" +
		      "ср.-пт.: 10:00-18:00\n" +
		      "вс.: 12:00-16:30", &Schedule{
		Tue: &ScheduleDay{
			From: 660,
			To:   1120,
		},
		Wed: sDay,
		Thu: sDay,
		Fri: sDay,
		Sun: &ScheduleDay{
			From: 720,
			To:   990,
		},
	})
}

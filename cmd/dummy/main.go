// Package main provides a dummy application for calculating shift rewards
// based on special dates and hourly rates.
package main

import (
	"fmt"
	"time"
)

const (
	HourlyRate = 5
	Multiplier = 2
)

type Shift struct {
	Start time.Time
	End   time.Time
}

type Shifts []Shift

func main() {
	specialDates := []time.Time{
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 5, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 10, 31, 0, 0, 0, 0, time.UTC),
	}

	shifts := Shifts{
		Shift{Start: time.Date(2024, 12, 31, 11, 0, 0, 0, time.UTC), End: time.Date(2025, 1, 10, 8, 0, 0, 0, time.UTC)},
		Shift{Start: time.Date(2025, 3, 21, 10, 0, 0, 0, time.UTC), End: time.Date(2025, 6, 22, 9, 45, 0, 0, time.UTC)},
	}

	/*
	 * Calculate the total reward for my shifts.
	 * For each shift, if any of the dates are special dates, the reward is multiplied by the multiplier.
	 * The reward is the number of hours worked * the hourly rate.
	 */
	totalReward := 0
	for _, shift := range shifts {
		// Calculate hours worked in this shift
		hours := int(shift.End.Sub(shift.Start).Hours())
		fmt.Printf("Shift from %v to %v: %d hours worked\n", shift.Start, shift.End, hours)

		// Count the number of special days in the shift
		specialDaysCount := 0
		current := shift.Start
		for current.Before(shift.End) {
			for _, specialDate := range specialDates {
				if current.Year() == specialDate.Year() &&
					current.Month() == specialDate.Month() &&
					current.Day() == specialDate.Day() {
					specialDaysCount++
					fmt.Printf("\tSpecial date %v found within shift\n", specialDate)
					break
				}
			}
			current = current.Add(24 * time.Hour)
		}
		fmt.Printf("\tTotal special days in shift: %d\n", specialDaysCount)

		// Calculate reward for this shift
		shiftReward := hours * HourlyRate
		if specialDaysCount > 0 {
			shiftReward *= Multiplier * specialDaysCount
			fmt.Printf("\tShift reward with multiplier: %d\n", shiftReward)
		} else {
			fmt.Printf("\tShift reward without multiplier: %d\n", shiftReward)
		}
		totalReward += shiftReward
	}

	fmt.Println(totalReward)
}

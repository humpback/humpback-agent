package utils

import (
	"fmt"
	"time"
)

func HumanDuration(d time.Duration) string {
	if seconds := int(d.Seconds()); seconds < 1 {
		return "Less than a second"
	} else if seconds == 1 {
		return "Up 1 second"
	} else if seconds < 60 {
		return fmt.Sprintf("Up %d seconds", seconds)
	} else if minutes := int(d.Minutes()); minutes == 1 {
		return "About a minute"
	} else if minutes < 60 {
		return fmt.Sprintf("Up %d minutes", minutes)
	} else if hours := int(d.Hours() + 0.5); hours == 1 {
		return "About an hour"
	} else if hours < 48 {
		return fmt.Sprintf("Up %d hours", hours)
	} else if hours < 24*7*2 {
		return fmt.Sprintf("Up %d days", hours/24)
	} else if hours < 24*30*2 {
		return fmt.Sprintf("Up %d weeks", hours/24/7)
	} else if hours < 24*365*2 {
		return fmt.Sprintf("Up %d months", hours/24/30)
	}
	return fmt.Sprintf("Up %d years", int(d.Hours())/24/365)
}

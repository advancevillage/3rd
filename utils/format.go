//author: richard
package utils

import "regexp"

func ValidateEmail(email string) bool {
	re := regexp.MustCompile("^([a-z0-9_\\.-]+)@([\\da-z\\.-]+)\\.([a-z\\.]{2,6})$")
	return re.MatchString(email)
}

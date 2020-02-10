//author: richard
package utils

import "regexp"

func ValidateEmail(email string) bool {
	re := regexp.MustCompile(`^\w[-\w.+]*@([A-Za-z0-9][-A-Za-z0-9]+\.)+[A-Za-z]{2,14}$`)
	return re.MatchString(email)
}

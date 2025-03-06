package helpers

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Handler is a global variable for the Helper instance
var (
	Handler *Helper
)

// Boot initializes the Handler instance with default values
func Boot() {
	Handler = &Helper{
		Orm:  *NewOrm(),
		Bind: *NewBind(),
	}
}

// Helper is a struct that aggregates various helper functionalities
type Helper struct {
	Orm  Orm
	Bind Bind
}

var (
	ErrParseInt       = errors.New("error parsing Int")          // HTTP 400 Bad Request
	ErrParseUUID      = errors.New("error parsing UUID")         // HTTP 400 Bad Request
	ErrParseDate      = errors.New("error parsing Date")         // HTTP 400 Bad Request
	ErrBirthDateToAge = errors.New("birthdate is in the future") // HTTP 400 Bad Request
	ErrFieldNotFound  = errors.New("field not found")            // HTTP 404 Not Found
)

// StringToInt converts a string to an int.
// Returns: (int, error)
func StringToInt(str string) (int, error) {
	result, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrParseInt, err)
	}
	return result, nil
}

// BirthDateToAge converts a time.Time (birthdate) to an int (age).
// Returns: (int, error)
func BirthDateToAge(birthdate time.Time) (int, error) {
	now := time.Now()
	if birthdate.After(now) {
		return 0, ErrBirthDateToAge
	}

	age := now.Year() - birthdate.Year()

	// Adjust if the birthday hasn't occurred yet this year
	if now.Month() < birthdate.Month() || (now.Month() == birthdate.Month() && now.Day() < birthdate.Day()) {
		age--
	}

	return age, nil
}

// IsValidInt checks if a string is a valid integer.
// Returns: error
func IsValidInt(id string) error {
	_, err := strconv.Atoi(id)
	return err
}

// CleanText removes extra spaces from a string and trims it.
// Returns: string
func CleanText(text string) string {
	if text != "" {
		regexSpaces := regexp.MustCompile(`\s+`)
		text = regexSpaces.ReplaceAllString(text, " ")
		text = strings.TrimSpace(text)
	}
	return text
}

// IntegerToEsEs converts an integer to a string in Spanish
// Returns: string
func IntegerToEsEs(input int) string {
	var spanishMegasPlural = []string{"", "mil", "millones", "mil millones", "billones"}
	var spanishUnits = []string{"", "uno", "dos", "tres", "cuatro", "cinco", "seis", "siete", "ocho", "nueve"}
	var spanishHundreds = []string{"", "ciento", "doscientos", "trescientos", "cuatrocientos", "quinientos", "seiscientos", "setecientos", "ochocientos", "novecientos"}
	var spanishTens = []string{"", "diez", "veinte", "treinta", "cuarenta", "cincuenta", "sesenta", "setenta", "ochenta", "noventa"}
	var spanishTeens = []string{"diez", "once", "doce", "trece", "catorce", "quince", "dieciséis", "diecisiete", "dieciocho", "diecinueve"}
	var spanishTwenties = []string{"veinte", "veintiuno", "veintidós", "veintitrés", "veinticuatro", "veinticinco", "veintiséis", "veintisiete", "veintiocho", "veintinueve"}
	words := []string{}
	if input < 0 {
		words = append(words, "menos")
		input *= -1
	}
	if input == 0 {
		return "cero"
	}

	triplets := integerToTriplets(input)

	for idx := len(triplets) - 1; idx >= 0; idx-- {
		triplet := triplets[idx]
		if triplet == 0 {
			continue
		}

		hundreds := triplet / 100 % 10
		tens := triplet / 10 % 10
		units := triplet % 10

		if hundreds > 0 {
			words = append(words, spanishHundreds[hundreds])
		}

		if tens == 0 && units == 0 {

			continue
		}

		switch tens {
		case 0:

			words = append(words, spanishUnits[units])
		case 1:

			words = append(words, spanishTeens[units])
		case 2:

			if units == 0 {
				words = append(words, spanishTens[tens])
			} else {
				words = append(words, spanishTwenties[units])
			}
		default:

			if units > 0 {
				words = append(words, fmt.Sprintf("%s y %s", spanishTens[tens], spanishUnits[units]))
			} else {
				words = append(words, spanishTens[tens])
			}
		}

		if idx > 0 {
			mega := spanishMegasPlural[idx]
			if mega != "" {
				words = append(words, mega)
			}
		}
	}
	return strings.Join(words, " ")
}

// integerToTriplets divides a number into triplets (thousands, millions, etc.)
// Returns: a slice of integers representing the triplets
func integerToTriplets(input int) []int {
	var triplets []int
	for input > 0 {
		triplets = append(triplets, input%1000)
		input /= 1000
	}
	return triplets
}

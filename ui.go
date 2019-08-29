package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

type errorType string

func (et errorType) Error() string { return string(et) }

const errorUserTerminated = errorType("user terminated")
const errorChoices = errorType("choice string format")
const errorIllegalColumn = errorType("column value must be one or zero")

func readGrid(r io.Reader, w io.Writer) ([][]int, error) {
	rows, cols, err := getDimensions(r, w)
	if err != nil {
		return nil, err
	}

	var grid [][]int
	for i := 0; i < rows; i++ {
		row, err := getRow(r, w, cols)
		if err != nil {
			return nil, err
		}
		grid = append(grid, row)
	}

	return grid, nil
}

// Substrings that are surrounded by parenthesis are the valid response.  For example (A) with yield a valid response of
// A as in "Choose (A) or (B)"
func parseChoices(prompt string) ([]string, error) {
	re := regexp.MustCompile(`\(.\)`)
	choices := re.FindAllString(prompt, -1)
	for i := 0; i < len(choices); i++ {
		choices[i] = strings.Trim(choices[i], "()")
	}
	if len(choices) == 0 {
		return nil, errorChoices
	}
	return choices, nil
}

// Prompt only returns valid choices, will prompt user to retry if choice is invalid.
func prompt(r io.Reader, w io.Writer, prompt string) (string, error) {
	choices, err := parseChoices(prompt)
	if err != nil {
		return "", err
	}
	for {
		_, err := fmt.Fprint(w, prompt)
		if err != nil {
			return "", err
		}
		var response string
		fmt.Fscanln(r, &response)
		if validChoice(response, choices...) {
			return response, nil
		}
		fmt.Fprintf(w, "%q invalid choice, try again\n", response)
	}
}

func validChoice(response string, choices ...string) bool {
	for _, c := range choices {
		if c == response {
			return true
		}
	}
	return false
}

func getDimensions(r io.Reader, w io.Writer) (rows, cols int, err error) {
	fmt.Fprint(w, "Enter dimensions row count and column count separated by a space. ")
	_, err = fmt.Fscanln(r, &rows, &cols)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(w, "You entered %d rows and %d columns.\n", rows, cols)

	var response string
	response, err = prompt(r, w, "Continue (C), Retry (R) Cancel (X)? ")
	if err != nil {
		return
	}
	switch response {
	case "R":
		return getDimensions(r, w)
	case "X":
		return 0, 0, errorUserTerminated
	case "C":
	}
	return

}

func readRow(rdr io.Reader, cols int) ([]int, error) {
	row := make([]int, cols)
	var read []interface{}
	for i := 0; i < cols; i++ {
		read = append(read, &row[i])
	}

	if _, err := fmt.Fscanln(rdr, read...); err != nil {
		return nil, err
	}

	for i := 0; i < cols; i++ {
		switch row[i] {
		case 0, 1:
		default:
			return nil, errorIllegalColumn
		}
	}

	return row, nil
}

func getRow(r io.Reader, w io.Writer, cols int)([]int,error){
	for {
		var entries []int
		var err error

		for {
			_, _ = fmt.Fprintf(w, "Enter a %d element row containing space separated ones or zeros\n", cols)
			entries, err = readRow(r, cols)
			if err != nil {
				choice, err := prompt(r, w, fmt.Sprintf("Error: %q Retry (R) Cancel (X)? ", err))
				if err != nil {
					return nil, err
				}
				switch choice {
				case "X":
					return nil, errorUserTerminated
				case "R":
					continue
				}
			}
			break
		}

		_, _ = fmt.Fprintf(w, "You entered %v\n", entries)
		choice, err := prompt(r,w, "Continue (C) Retry (R) Cancel (X)? " )
		if err != nil {
			return nil, err
		}
		switch choice {
		case "C":
			return entries, nil
		case "X":
			return nil, errorUserTerminated
		}
	}
}


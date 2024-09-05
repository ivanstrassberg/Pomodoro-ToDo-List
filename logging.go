package main

import (
	"bufio"
	"os"
)

func CreateLogFile() (fp *os.File) {
	fp, err := os.Create("logs.txt")

	if err != nil {
		return nil
	}

	return fp
}

func WriteToLog(scanner bufio.Scanner, value string) error {
	return nil
}

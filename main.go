package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Syntax: %s <PrinterName> <FileName>\n", os.Args[0])
		os.Exit(1)
	}

	fmt.Printf(
		"Attempting to send file [%s] to printer [%s].\n",
		os.Args[2],
		os.Args[1],
	)

	printer, err := OpenNewPrinter(os.Args[1])
	if err != nil {
		fmt.Println("[Error] Could not open printer", err)
		return
	}

	printer.openDoc("DKNCK Receipt")
	printer.openPage()

	err = printer.writeFile(os.Args[2])
	if err != nil {
		fmt.Println("[Error] Could not print file", err)
	}

	printer.closePage()
	printer.closeDoc()

	err = printer.Close()
	if err != nil {
		fmt.Println("[Error] Could not correctly close printer", err)
	}
}

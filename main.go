package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/hennedo/escpos"
)

var (
	action  string
	port    int
	outFile string
)

func init() {
	flag.StringVar(&action, "action", "listen", "listen, print, or generate?")
	flag.IntVar(&port, "port", 35625, "Port to use when serving")
	flag.StringVar(&outFile, "outfile", "output.bin", "Output filename for the generated escpos code")
}

func printFile(printerName string, file string) {
	fmt.Printf(
		"Attempting to send file [%s] to printer [%s].\n",
		file,
		printerName,
	)

	printer, err := OpenNewPrinter(printerName)
	if err != nil {
		fmt.Println("[Error] Could not open printer", err)
		return
	}
	defer func() {
		err := printer.Close()
		if err != nil {
			fmt.Println("[Error] Could not correctly close printer", err)
		}
	}()

	printer.openDoc("DKNCK Receipt")
	printer.openPage()

	defer func() {
		printer.closePage()
		printer.closeDoc()
	}()

	err = printer.writeFile(file)
	if err != nil {
		fmt.Println("[Error] Could not print file", err)
		return
	}
}

func listen(printerName string, file string) {
	http.HandleFunc("/print-escpos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}
		// r.
		// printFile(printerName, file)
	})

	fmt.Printf("Starting server on port %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server closed\n")
	} else if err != nil {
		fmt.Printf("[Error] Could not start server: %s\n", err)
	}
}

func generate() {
	f, _ := os.Create(outFile)
	defer f.Close()

	p := escpos.New(f)

	p.Bold(true).Size(2, 2).Justify(escpos.JustifyCenter).Write("DK & NCK")
	p.LineFeed()
	p.Bold(false).Size(1, 1).Justify(escpos.JustifyCenter).Write("Receipt ID: 518")
	p.LineFeed()
	p.Bold(false).Size(1, 1).Justify(escpos.JustifyCenter).Write("Date: 23/02/2024")
	p.LineFeed()
	p.LineFeed()

	table := `
ID       Item                   Price Qty Total
-----------------------------------------------
00000000 JACKET CANADA GOES      3000   7 21000
-----------------------------------------------
                               Sub Total: 21000
`

	p.Write(table)
	p.LineFeed()
	p.LineFeed()

	byteCode := append([]byte("DKNCKS518"), 0)
	p.WriteRaw(append([]byte{0x1d, 0x6b, 0x04}, byteCode...))

	p.PrintAndCut()
	printFile(flag.Args()[0], "output.bin")
}

func main() {
	flag.Parse()
	args := flag.Args()

	switch action {
	case "print":
		printFile(args[0], args[1])
	case "listen":
		listen(args[0], args[1])
	case "generate":
		generate()
	}
}

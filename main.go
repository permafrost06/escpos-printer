package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
)

var (
	action  string
	port    int
	outFile string
	file    string
)

var printerName = "Receipt Printer"

const docName = "DK & NCK Receipt"

func init() {
	flag.StringVar(&action, "action", "listen", "listen, print, or generate?")
	flag.IntVar(&port, "port", 35625, "Port to use when serving")
	flag.StringVar(&outFile, "outfile", "output.bin", "Output filename for the generated escpos code")
}

type InvoiceItem struct {
	Product_id  string
	Name        string
	Unit_price  int
	Quantity    int
	Total_price int
}

type Invoice struct {
	ID       int
	Date     string
	Items    []InvoiceItem
	Subtotal int
}

type PrintRequest struct {
	Secret_key string
	Invoice    Invoice
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func listen() {
	http.HandleFunc("/print-escpos", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		if r.Method != "POST" {
			http.Error(w, "method not supported", http.StatusBadRequest)
			return
		}

		var printReq PrintRequest
		err := json.NewDecoder(r.Body).Decode(&printReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if printReq.Secret_key != "supersecret" {
			http.Error(w, "not authorized", http.StatusBadRequest)
			return
		}

		bytes := GetInvoiceBytes(printReq)
		PrintBytes(printerName, bytes, docName)

		fmt.Fprintf(w, "everything looks okay")
	})

	fmt.Printf("Starting server on port %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server closed\n")
	} else if err != nil {
		fmt.Printf("[Error] Could not start server: %s\n", err)
	}
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 1 {
		printerName = args[0]
	}

	if len(args) == 2 {
		file = args[1]
	}

	switch action {
	case "print":
		fmt.Printf(
			"Attempting to send file [%s] to printer [%s].\n",
			file,
			printerName,
		)

		err := PrintFile(printerName, file, "DKNCK Receipt")

		if err != nil {
			fmt.Println("[Error] Could not print file", err)
			return
		}
	case "listen":
		listen()
	case "generate":
		GenerateFile(outFile)
	}
}

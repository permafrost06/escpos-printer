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
	Unit_price  string
	Quantity    string
	Total_price string
}

type Invoice struct {
	ID       string
	Date     string
	Items    []InvoiceItem
	Subtotal string
}

type PrintRequest struct {
	Secret_key string
	Invoice    Invoice
}

func listen() {
	http.HandleFunc("/print-receipt", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			fmt.Println("OPTIONS request received")
			w.Header().Add("Connection", "keep-alive")
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Methods", "POST")
			w.Header().Add("Access-Control-Allow-Headers", "content-type")
			w.Header().Add("Access-Control-Max-Age", "86400")
			fmt.Println("handled preflight")
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method != "POST" {
			fmt.Println("not POST or OPTIONS method, abort")
			http.Error(w, "method not supported", http.StatusBadRequest)
			return
		}

		var printReq PrintRequest
		err := json.NewDecoder(r.Body).Decode(&printReq)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println("request:", printReq)

		if printReq.Secret_key != "supersecret" {
			fmt.Println("not authorized")
			http.Error(w, "not authorized", http.StatusBadRequest)
			return
		}

		fmt.Println("secret key matches")

		bytes := GetInvoiceBytes(printReq)
		fmt.Printf("sending receipt to printer %s\n", printerName)
		PrintBytes(printerName, bytes, docName)
		fmt.Println("receipt sent to printer")

		fmt.Fprintf(w, "printing receipt")
	})

	fmt.Printf("Starting server on port %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	fmt.Printf("listening for print requests on http://localhost:%d\n/35625", port)

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
	}
}

package main

import (
	"encoding/json"
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

func printShopName(p *escpos.Escpos) {
	p.Bold(true).Size(2, 2).Justify(escpos.JustifyCenter).Write("DK & NCK")
	p.LineFeed()
}

func printAddress(p *escpos.Escpos) {
	address := `Shop: 32 & 44, 4th Floor, Anexco Tower
8 Phoenix Road, Fulbaria, Shahbag
Dhaka-1000`

	p.Bold(false).Size(1, 1).Justify(escpos.JustifyCenter).Write(address)
	p.LineFeed()
}

func printPhone(p *escpos.Escpos) {
	phone := `Phone: 01556341569, 01832775999`

	p.Bold(false).Size(1, 1).Justify(escpos.JustifyCenter).Write(phone)
	p.LineFeed()
}

func printInvoiceId(p *escpos.Escpos, id int) {
	idString := fmt.Sprintf("Invoice No: %d", id)

	p.Bold(false).Size(1, 1).Justify(escpos.JustifyCenter).Write(idString)
	p.LineFeed()
}

func printDate(p *escpos.Escpos, date string) {
	dateString := fmt.Sprintf("Date: %s", date)

	p.Bold(false).Size(1, 1).Justify(escpos.JustifyCenter).Write(dateString)
	p.LineFeed()
}

func printTableHeader(p *escpos.Escpos) {
	header := `
ID       Item                   Price Qty Total
-----------------------------------------------`
	// 00000000 TROWSER SHARPA          3000   7 21000
	// -----------------------------------------------
	//                                Sub Total: 21000

	p.Write(header)
	p.LineFeed()
}

func printItems(p *escpos.Escpos, items []InvoiceItem) {
	for _, item := range items {
		row := fmt.Sprintf(
			"%s %s %d %d %d",
			item.Product_id,
			item.Name,
			item.Unit_price,
			item.Quantity,
			item.Total_price,
		)
		p.Bold(false).Size(1, 1).Justify(escpos.JustifyCenter).Write(row)
		p.LineFeed()
	}
}

func printTableFooter(p *escpos.Escpos, subtotal int) {
	footer := "-----------------------------------------------"
	footer += fmt.Sprintf("                                Sub Total: %d", subtotal)

	p.Write(footer)
	p.LineFeed()
}

func printBarcode(p *escpos.Escpos, id int) {
	byteCode := append([]byte(fmt.Sprintf("DKNCKS%d", id)), 0)
	p.WriteRaw(append([]byte{0x1d, 0x6b, 0x04}, byteCode...))
}

func printMessage(p *escpos.Escpos) {
	message := `Please bring cash memo for returning products`

	p.Bold(false).Size(1, 1).Justify(escpos.JustifyCenter).Write(message)
	p.LineFeed()

	p.Bold(true).Size(1, 1).Justify(escpos.JustifyCenter).Write("Thank you for shopping with DK & NCK")
	p.LineFeed()
}

func printBottomPadding(p *escpos.Escpos) {
	p.LineFeed()
	p.LineFeed()
	p.LineFeed()
}

func printInvoice(req PrintRequest) {
	f, _ := os.Create(outFile)

	p := escpos.New(f)
	printShopName(p)
	printAddress(p)
	printPhone(p)

	p.LineFeed()

	printInvoiceId(p, req.Invoice.ID)
	printDate(p, req.Invoice.Date)

	p.LineFeed()

	printTableHeader(p)
	printItems(p, req.Invoice.Items)
	printTableFooter(p, req.Invoice.Subtotal)

	p.LineFeed()

	printBarcode(p, req.Invoice.ID)
	printMessage(p)

	printBottomPadding(p)

	p.PrintAndCut()
	f.Close()
	printFile(flag.Args()[0], "output.bin")
}

func listen(printerName string, file string) {
	http.HandleFunc("/print-escpos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}

		var printReq PrintRequest
		err := json.NewDecoder(r.Body).Decode(&printReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		printInvoice(printReq)
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

	address := `Shop: 32 & 44, 4th Floor, Anexco Tower
8 Phoenix Road, Fulbaria, Shahbag
Dhaka-1000`

	p.Bold(false).Size(1, 1).Justify(escpos.JustifyCenter).Write(address)
	p.LineFeed()

	phone := `Phone: 01556341569, 01832775999`

	p.Bold(false).Size(1, 1).Justify(escpos.JustifyCenter).Write(phone)
	p.LineFeed()
	p.LineFeed()

	p.Bold(false).Size(1, 1).Justify(escpos.JustifyCenter).Write("Invoice No: 518")
	p.LineFeed()
	p.Bold(false).Size(1, 1).Justify(escpos.JustifyCenter).Write("Date: 23/02/2024")
	p.LineFeed()
	p.LineFeed()

	table := `
ID       Item                   Price Qty Total
-----------------------------------------------
00000066 TROWSER SHARPA           900   2  1800
00000065 SHOE RED TAPE           2800   1  2800
00000062 HAND GLOVES              500   1   500
-----------------------------------------------
                               Sub Total:  5100
`

	p.Write(table)
	p.LineFeed()
	p.LineFeed()

	byteCode := append([]byte("DKNCKS518"), 0)
	p.WriteRaw(append([]byte{0x1d, 0x6b, 0x04}, byteCode...))

	message := `Please bring cash memo for returning products`

	p.Bold(false).Size(1, 1).Justify(escpos.JustifyCenter).Write(message)
	p.LineFeed()

	p.Bold(true).Size(1, 1).Justify(escpos.JustifyCenter).Write("Thank you for shopping with DK & NCK")
	p.LineFeed()
	p.LineFeed()
	p.LineFeed()
	p.LineFeed()

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

package main

import (
	"bytes"
	"fmt"

	"github.com/permafrost06/escpos"
)

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
ID       Item                    Price Qty Total
------------------------------------------------`

	p.Write(header)
	p.LineFeed()
}

func trim(str string, length int) string {
	if len(str) < length {
		return str
	}

	return str[:length]
}

func printItems(p *escpos.Escpos, items []InvoiceItem) {
	for _, item := range items {
		row := fmt.Sprintf(
			"%08s %-23s %5s %3s %5s",
			trim(item.Product_id, 8),
			trim(item.Name, 23),
			trim(fmt.Sprint(item.Unit_price), 5),
			trim(fmt.Sprint(item.Quantity), 3),
			trim(fmt.Sprint(item.Total_price), 5),
		)
		p.Bold(false).Size(1, 1).Justify(escpos.JustifyCenter).Write(row)
		p.LineFeed()
	}
}

func printTableFooter(p *escpos.Escpos, subtotal int) {
	footer := "------------------------------------------------"
	footer += fmt.Sprintf("                                 Sub Total: %d", subtotal)

	p.Write(footer)
	p.LineFeed()
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

func GetInvoiceBytes(req PrintRequest) []byte {
	f := new(bytes.Buffer)

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

	p.WriteRaw([]byte{0x1d, 0x68, 0x50})
	p.WriteRaw([]byte{0x1d, 0x77, 0x04})
	p.WriteRaw([]byte{0x1d, 0x66, 0x00})
	p.WriteRaw([]byte{0x1d, 0x48, 0x02})
	p.WriteRaw([]byte{0x1d, 0x6b, 0x04})
	barcode := fmt.Sprintf("S%08d", req.Invoice.ID)
	byteCode := append([]byte(barcode), 0)
	p.WriteRaw(byteCode)

	printMessage(p)

	printBottomPadding(p)

	p.PrintAndCut()

	return f.Bytes()
}

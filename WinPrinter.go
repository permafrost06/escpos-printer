// credit goes to https://gist.github.com/gerkirill/35f5d1cf4abd4f569e33
package main

import (
	"os"
	"syscall"
	"unsafe"
)

var winspool = syscall.NewLazyDLL("Winspool.drv")

type docInfo struct {
	pDocName    uintptr
	pOutputFile uintptr
	pDatatype   uintptr
}

type Printer struct {
	printer syscall.Handle
}

func stringPtr(str string) *uint16 {
	ptr, err := syscall.UTF16PtrFromString(str)

	if err != nil {
		panic(err)
	}

	return ptr
}

func OpenNewPrinter(name string) (*Printer, error) {
	printer := new(Printer)
	err := printer.open(name)
	if err != nil {
		return nil, err
	}
	return printer, nil
}

func (p *Printer) open(name string) error {
	var openPrinter = winspool.NewProc("OpenPrinterW")
	ret, _, err := openPrinter.Call(
		uintptr(unsafe.Pointer(stringPtr(name))),
		uintptr(unsafe.Pointer(&p.printer)),
		uintptr(unsafe.Pointer(nil)))
	if ret != 1 {
		return err
	}
	return nil
}

func (p *Printer) Close() error {
	var closePrinter = winspool.NewProc("ClosePrinter")
	ret, _, err := closePrinter.Call(uintptr(unsafe.Pointer(p.printer)))
	if ret != 1 {
		return err
	}
	return nil
}

func (p *Printer) openPage() error {
	var startPagePrinter = winspool.NewProc("StartPagePrinter")
	ret, _, err := startPagePrinter.Call(uintptr(unsafe.Pointer(p.printer)))
	if ret != 1 {
		return err
	}
	return nil
}

func (p *Printer) closePage() {
	var endDocPrinter = winspool.NewProc("EndDocPrinter")
	endDocPrinter.Call(uintptr(unsafe.Pointer(p.printer)))
}

func (p *Printer) openDoc(name string) error {
	var startDocPrinter = winspool.NewProc("StartDocPrinterW")
	var level uint32 = 1

	var doc docInfo

	doc.pDocName = uintptr(unsafe.Pointer(stringPtr(name)))
	doc.pOutputFile = uintptr(unsafe.Pointer(nil))
	doc.pDatatype = uintptr(unsafe.Pointer(stringPtr("RAW")))

	ret, _, err := startDocPrinter.Call(
		uintptr(unsafe.Pointer(p.printer)),
		uintptr(level),
		uintptr(unsafe.Pointer(&doc)))
	if ret == 0 {
		return err
	}

	return nil
}

func (p *Printer) closeDoc() {
	var endPagePrinter = winspool.NewProc("EndPagePrinter")
	endPagePrinter.Call(uintptr(unsafe.Pointer(p.printer)))
}

func (p *Printer) writeFile(path string) error {
	var writePrinter = winspool.NewProc("WritePrinter")
	document, err := os.ReadFile(path)
	if nil != err {
		return err
	}
	var bytesWritten uint32 = 0
	var docSize uint32 = uint32(len(document))
	ret, _, err := writePrinter.Call(
		uintptr(unsafe.Pointer(p.printer)),
		uintptr(unsafe.Pointer(&document[0])),
		uintptr(docSize),
		uintptr(unsafe.Pointer(&bytesWritten)))
	if ret != 1 {
		return err
	}

	return nil
}

func (p *Printer) writeBytes(document []byte) error {
	var writePrinter = winspool.NewProc("WritePrinter")
	var bytesWritten uint32 = 0
	var docSize uint32 = uint32(len(document))
	ret, _, err := writePrinter.Call(
		uintptr(unsafe.Pointer(p.printer)),
		uintptr(unsafe.Pointer(&document[0])),
		uintptr(docSize),
		uintptr(unsafe.Pointer(&bytesWritten)))
	if ret != 1 {
		return err
	}

	return nil
}

func PrintFile(printerName string, fileName string, docName string) error {
	printer, err := OpenNewPrinter(printerName)
	if err != nil {
		return err
	}
	defer func() error {
		err := printer.Close()
		if err != nil {
			return err
		}
		return nil
	}()

	printer.openDoc(docName)
	printer.openPage()

	defer func() {
		printer.closePage()
		printer.closeDoc()
	}()

	err = printer.writeFile(fileName)
	if err != nil {
		return err
	}

	return nil
}

func PrintBytes(printerName string, document []byte, docName string) error {
	printer, err := OpenNewPrinter(printerName)
	if err != nil {
		return err
	}
	defer func() error {
		err := printer.Close()
		if err != nil {
			return err
		}
		return nil
	}()

	printer.openDoc(docName)
	printer.openPage()

	defer func() {
		printer.closePage()
		printer.closeDoc()
	}()

	err = printer.writeBytes(document)
	if err != nil {
		return err
	}

	return nil
}

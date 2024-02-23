# escpos Printer

Use this program to send escpos commands in a file to a printer. **(Windows only)**

## Credits

Windows printer API implementation in go https://gist.github.com/gerkirill/35f5d1cf4abd4f569e33

## Usage

```powershell
escpos-printer <printer-name> <file-to-print>
```

The file to be printed must be a binary file. The extension does not matter.
Two example files are provided in the `escpos_examples` directory.

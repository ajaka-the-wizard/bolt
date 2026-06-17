package domain

import "path/filepath"

var (
	// We'll be storing the file to the server, it should be noted that it is more appropriate to store in an object storage
	BoltInvoiceOutPutPath = filepath.Join("bolt", "invoice")
)

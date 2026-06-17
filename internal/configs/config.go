package configs

import (
	"log/slog"
	"os"

	"github.com/ajaka-the-wizard/bolt/internal/domain"
)

func EnsureAllIsFine(logger *slog.Logger) {
	err := os.MkdirAll(domain.BoltInvoiceOutPutPath, os.ModeDir.Perm())
	if err != nil {
		logger.Error("Couldnt ensure directories are created", "error", err)
		panic(err)
	}
	logger.Info("Invoice output directory is fine")
}

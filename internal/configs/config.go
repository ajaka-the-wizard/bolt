package configs

import (
	"log/slog"
	"os"

	"github.com/ajaka-the-wizard/bolt/internal/domain"
)

// This is a simple function to ensure the output directories for invoice generation exists. This is because writing to a file creates the file but not the directories resulting in a runtime error. This prevents that.
func EnsureAllIsFine(logger *slog.Logger) {
	err := os.MkdirAll(domain.BoltInvoiceOutputPath, os.ModeDir.Perm())
	if err != nil {
		logger.Error("Couldn't ensure directories are created", "error", err)
		panic(err)
	}
	logger.Info("Invoice output directory is fine")
}

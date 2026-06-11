package configs

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

// writeEnvFile writes the given content to a file named ".env" inside dir and
// returns a cleanup function.  Tests must call os.Chdir(dir) before invoking
// LoadEnv so that viper resolves ".env" relative to the working directory.
func writeEnvFile(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(content), 0600); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}
}

// chdirTemp creates a temp directory, changes the process working directory to
// it, and restores the original directory when the test finishes.
func chdirTemp(t *testing.T) string {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
	return dir
}

// validEnvContent returns a .env file that satisfies all required fields.
func validEnvContent() string {
	return `PORT=8080
DATABASE_URL=postgres://user:pass@localhost:5432/db
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=secret
SHARED_SECRET=supersecret
PRODUCTION=false
`
}

// ---- Tests -------------------------------------------------------------------

// TestLoadEnv_ValidEnvFile verifies that LoadEnv successfully parses a
// well-formed .env file and returns an Env struct with the expected values.
func TestLoadEnv_ValidEnvFile(t *testing.T) {
	dir := chdirTemp(t)
	writeEnvFile(t, dir, validEnvContent())

	env := LoadEnv(slog.Default())

	if env.PORT != "8080" {
		t.Errorf("PORT = %q, want %q", env.PORT, "8080")
	}
	if env.DATABASE_URL != "postgres://user:pass@localhost:5432/db" {
		t.Errorf("DATABASE_URL = %q", env.DATABASE_URL)
	}
	if env.REDIS_ADDR != "localhost:6379" {
		t.Errorf("REDIS_ADDR = %q", env.REDIS_ADDR)
	}
	if env.REDIS_PASSWORD != "secret" {
		t.Errorf("REDIS_PASSWORD = %q", env.REDIS_PASSWORD)
	}
	if env.SHARED_SECRET != "supersecret" {
		t.Errorf("SHARED_SECRET = %q", env.SHARED_SECRET)
	}
	if env.PRODUCTION != false {
		t.Errorf("PRODUCTION = %v, want false", env.PRODUCTION)
	}
}

// TestLoadEnv_UnmarshalExact_RejectsUnknownKeys verifies the PR change:
// UnmarshalExact (as opposed to Unmarshal) panics when the .env file contains
// keys that have no corresponding field in the Env struct.
func TestLoadEnv_UnmarshalExact_RejectsUnknownKeys(t *testing.T) {
	dir := chdirTemp(t)
	// UNKNOWN_KEY has no corresponding field in Env.
	content := validEnvContent() + "UNKNOWN_KEY=should_fail\n"
	writeEnvFile(t, dir, content)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected LoadEnv to panic on unknown key, but it did not")
		}
	}()

	LoadEnv(slog.Default())
}

// TestLoadEnv_MissingRequiredField_PORT panics when PORT is absent.
func TestLoadEnv_MissingRequiredField_PORT(t *testing.T) {
	dir := chdirTemp(t)
	writeEnvFile(t, dir, `DATABASE_URL=postgres://user:pass@localhost:5432/db
REDIS_ADDR=localhost:6379
SHARED_SECRET=supersecret
`)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing PORT, but none occurred")
		}
	}()

	LoadEnv(slog.Default())
}

// TestLoadEnv_MissingRequiredField_DatabaseURL panics when DATABASE_URL is absent.
func TestLoadEnv_MissingRequiredField_DatabaseURL(t *testing.T) {
	dir := chdirTemp(t)
	writeEnvFile(t, dir, `PORT=8080
REDIS_ADDR=localhost:6379
SHARED_SECRET=supersecret
`)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing DATABASE_URL, but none occurred")
		}
	}()

	LoadEnv(slog.Default())
}

// TestLoadEnv_MissingRequiredField_SharedSecret panics when SHARED_SECRET is absent.
func TestLoadEnv_MissingRequiredField_SharedSecret(t *testing.T) {
	dir := chdirTemp(t)
	writeEnvFile(t, dir, `PORT=8080
DATABASE_URL=postgres://user:pass@localhost:5432/db
REDIS_ADDR=localhost:6379
`)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing SHARED_SECRET, but none occurred")
		}
	}()

	LoadEnv(slog.Default())
}

// TestLoadEnv_NoEnvFile_PanicsEvenWithEnvVarsSet documents that when no .env
// file is present, LoadEnv panics even if all required variables exist in the
// OS environment.  This is because viper's AutomaticEnv does not populate
// UnmarshalExact for keys it has not previously "seen" (via config file,
// SetDefault, or BindEnv).  The log message "using system environment
// variables" is therefore aspirational, not functional.
func TestLoadEnv_NoEnvFile_PanicsEvenWithEnvVarsSet(t *testing.T) {
	// Use an empty temp dir so viper cannot find ".env".
	chdirTemp(t)

	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://env:pass@host:5432/testdb")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("REDIS_PASSWORD", "")
	t.Setenv("SHARED_SECRET", "envSecret")
	t.Setenv("PRODUCTION", "true")

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected LoadEnv to panic without a .env file, but it did not")
		}
	}()

	LoadEnv(slog.Default())
}

// TestLoadEnv_OptionalRedisPassword_CanBeEmpty verifies that REDIS_PASSWORD
// is optional (no validate:"required" tag) and LoadEnv succeeds when it is empty.
func TestLoadEnv_OptionalRedisPassword_CanBeEmpty(t *testing.T) {
	dir := chdirTemp(t)
	writeEnvFile(t, dir, `PORT=8080
DATABASE_URL=postgres://user:pass@localhost:5432/db
REDIS_ADDR=localhost:6379
SHARED_SECRET=supersecret
`)

	env := LoadEnv(slog.Default())
	if env.REDIS_PASSWORD != "" {
		t.Errorf("REDIS_PASSWORD = %q, want empty", env.REDIS_PASSWORD)
	}
}

// TestLoadEnv_ProductionDefaultsFalse verifies that PRODUCTION defaults to
// false when not specified (zero value of bool).
func TestLoadEnv_ProductionDefaultsFalse(t *testing.T) {
	dir := chdirTemp(t)
	writeEnvFile(t, dir, `PORT=8080
DATABASE_URL=postgres://user:pass@localhost:5432/db
REDIS_ADDR=localhost:6379
SHARED_SECRET=supersecret
`)

	env := LoadEnv(slog.Default())
	if env.PRODUCTION {
		t.Error("PRODUCTION should default to false when not set")
	}
}

// TestLoadEnv_UnmarshalExact_VsUnmarshal_BehaviouralRegression is a
// regression test documenting the specific behavioural difference introduced
// by the PR.  Unmarshal silently ignores unknown keys; UnmarshalExact rejects
// them.  This test asserts the *new* strict behaviour panics on unknown keys.
func TestLoadEnv_UnmarshalExact_VsUnmarshal_BehaviouralRegression(t *testing.T) {
	dir := chdirTemp(t)
	// EXTRA_FIELD is not a field on Env.
	writeEnvFile(t, dir, validEnvContent()+"EXTRA_FIELD=boom\n")

	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		LoadEnv(slog.Default())
	}()

	if !panicked {
		t.Error("regression: UnmarshalExact should reject unknown keys, but LoadEnv succeeded")
	}
}
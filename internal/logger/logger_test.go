package logger_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"testing"
	"time"

	"github.com/Alessandro-Massarotti-Jr/go-api/internal/logger"
)

const isoDateFormat = "2006-01-02T15:04:05.000Z07:00"

var logLinePattern = regexp.MustCompile(`^\[(.+?)\]\[([A-Z]+)\] (.+)\n$`)

// captureStdout replaces os.Stdout while fn runs and returns everything written to it.
func captureStdout(test *testing.T, fn func()) string {
	test.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		test.Fatalf("failed to create pipe: %v", err)
	}

	original := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = original }()

	fn()

	writer.Close()
	output, err := io.ReadAll(reader)
	if err != nil {
		test.Fatalf("failed to read output: %v", err)
	}
	return string(output)
}

// useLevel sets the logger level for one test and restores the default level afterwards.
func useLevel(test *testing.T, level logger.LogLevel) *logger.Logger {
	test.Helper()

	appLogger := logger.GetInstance()
	appLogger.SetLevel(level)
	test.Cleanup(func() { appLogger.SetLevel(logger.LevelInfo) })
	return appLogger
}

func TestLogger(test *testing.T) {

	test.Run("Get instance", func(test *testing.T) {
		test.Run("Should be the same instance", func(test *testing.T) {
			instance1 := logger.GetInstance()
			instance2 := logger.GetInstance()
			if instance1 != instance2 {
				test.Errorf("Expected the same instance, got different instances")
			}
		})
	})

	test.Run("SetLevel", func(test *testing.T) {
		test.Run("Should only print messages at or above the configured level", func(test *testing.T) {
			methods := []struct {
				name string
				call func(*logger.Logger, logger.LogMessage)
			}{
				{name: "Debug", call: (*logger.Logger).Debug},
				{name: "Info", call: (*logger.Logger).Info},
				{name: "Warn", call: (*logger.Logger).Warn},
				{name: "Error", call: (*logger.Logger).Error},
			}

			levels := []struct {
				level       logger.LogLevel
				wantVisible []string
			}{
				{level: logger.LevelDebug, wantVisible: []string{"Debug", "Info", "Warn", "Error"}},
				{level: logger.LevelInfo, wantVisible: []string{"Info", "Warn", "Error"}},
				{level: logger.LevelWarn, wantVisible: []string{"Warn", "Error"}},
				{level: logger.LevelError, wantVisible: []string{"Error"}},
				{level: logger.LevelSilent, wantVisible: []string{}},
			}

			for _, tc := range levels {
				test.Run(string(tc.level), func(test *testing.T) {
					appLogger := useLevel(test, tc.level)

					for _, method := range methods {
						output := captureStdout(test, func() {
							method.call(appLogger, logger.LogMessage{Action: "main", Message: "set level"})
						})

						printed := output != ""
						wantPrinted := slices.Contains(tc.wantVisible, method.name)
						if printed != wantPrinted {
							test.Errorf("%s with level %s: printed = %v, want %v", method.name, tc.level, printed, wantPrinted)
						}
					}
				})
			}
		})

		test.Run("Should allow changing the level more than once", func(test *testing.T) {
			appLogger := useLevel(test, logger.LevelError)
			appLogger.SetLevel(logger.LevelDebug)

			output := captureStdout(test, func() {
				appLogger.Debug(logger.LogMessage{Action: "main", Message: "visible"})
			})

			if output == "" {
				test.Error("Expected Debug to print after changing the level from error to debug, got nothing")
			}
		})

		invalidLevels := []logger.LogLevel{"banana", "", "INFO", " info"}
		for _, invalid := range invalidLevels {
			test.Run(fmt.Sprintf("Should ignore invalid level %q and keep the previous one", invalid), func(test *testing.T) {
				appLogger := useLevel(test, logger.LevelWarn)
				appLogger.SetLevel(invalid)

				// if the invalid level had been applied, its weight would be 0 and Info would print
				infoOutput := captureStdout(test, func() {
					appLogger.Info(logger.LogMessage{Action: "main", Message: "hidden"})
				})
				if infoOutput != "" {
					test.Errorf("Expected Info to stay hidden with level warn, got %q", infoOutput)
				}

				warnOutput := captureStdout(test, func() {
					appLogger.Warn(logger.LogMessage{Action: "main", Message: "visible"})
				})
				if warnOutput == "" {
					test.Error("Expected Warn to still print with level warn, got nothing")
				}
			})
		}
	})

	test.Run("Info", func(test *testing.T) {
		test.Run("Should print the message in the [date][INFO] json format", func(test *testing.T) {
			appLogger := useLevel(test, logger.LevelInfo)

			output := captureStdout(test, func() {
				appLogger.Info(logger.LogMessage{
					Action:  "user.create",
					Message: "user created",
					Data:    map[string]any{"id": 1},
				})
			})

			matches := logLinePattern.FindStringSubmatch(output)
			if matches == nil {
				test.Fatalf("Output %q does not match the expected format", output)
			}
			date, level, payload := matches[1], matches[2], matches[3]

			if _, err := time.Parse(isoDateFormat, date); err != nil {
				test.Errorf("Date %q is not in ISO format: %v", date, err)
			}

			if level != "INFO" {
				test.Errorf("Level = %q, want %q", level, "INFO")
			}

			var got map[string]any
			if err := json.Unmarshal([]byte(payload), &got); err != nil {
				test.Fatalf("Payload %q is not valid JSON: %v", payload, err)
			}
			if got["action"] != "user.create" {
				test.Errorf("action = %v, want %q", got["action"], "user.create")
			}
			if got["message"] != "user created" {
				test.Errorf("message = %v, want %q", got["message"], "user created")
			}
			data, ok := got["data"].(map[string]any)
			if !ok {
				test.Fatalf("data = %v, want an object", got["data"])
			}
			// encoding/json decodes every number as float64
			if data["id"] != float64(1) {
				test.Errorf("data.id = %v, want 1", data["id"])
			}
		})

		test.Run("Should omit data when it is nil", func(test *testing.T) {
			appLogger := useLevel(test, logger.LevelInfo)

			output := captureStdout(test, func() {
				appLogger.Info(logger.LogMessage{Action: "main", Message: "no data"})
			})

			matches := logLinePattern.FindStringSubmatch(output)
			if matches == nil {
				test.Fatalf("Output %q does not match the expected format", output)
			}

			want := `{"action":"main","message":"no data"}`
			if matches[3] != want {
				test.Errorf("Payload = %s, want %s", matches[3], want)
			}
		})

		test.Run("Should print when level is debug", func(test *testing.T) {
			appLogger := useLevel(test, logger.LevelDebug)

			output := captureStdout(test, func() {
				appLogger.Info(logger.LogMessage{Action: "main", Message: "visible"})
			})

			if output == "" {
				test.Error("Expected Info to print with level debug, got nothing")
			}
		})

		levelsAboveInfo := []logger.LogLevel{logger.LevelWarn, logger.LevelError, logger.LevelSilent}
		for _, level := range levelsAboveInfo {
			test.Run("Should not print when level is "+string(level), func(test *testing.T) {
				appLogger := useLevel(test, level)

				output := captureStdout(test, func() {
					appLogger.Info(logger.LogMessage{Action: "main", Message: "hidden"})
				})

				if output != "" {
					test.Errorf("Expected no output with level %s, got %q", level, output)
				}
			})
		}
	})

	test.Run("Debug", func(test *testing.T) {
		testLogMethod(test, logMethodCase{
			call:      (*logger.Logger).Debug,
			label:     "DEBUG",
			visibleAt: []logger.LogLevel{logger.LevelDebug},
			hiddenAt:  []logger.LogLevel{logger.LevelInfo, logger.LevelWarn, logger.LevelError, logger.LevelSilent},
		})
	})

	test.Run("Warn", func(test *testing.T) {
		testLogMethod(test, logMethodCase{
			call:      (*logger.Logger).Warn,
			label:     "WARN",
			visibleAt: []logger.LogLevel{logger.LevelDebug, logger.LevelInfo, logger.LevelWarn},
			hiddenAt:  []logger.LogLevel{logger.LevelError, logger.LevelSilent},
		})
	})

	test.Run("Error", func(test *testing.T) {
		testLogMethod(test, logMethodCase{
			call:      (*logger.Logger).Error,
			label:     "ERROR",
			visibleAt: []logger.LogLevel{logger.LevelDebug, logger.LevelInfo, logger.LevelWarn, logger.LevelError},
			hiddenAt:  []logger.LogLevel{logger.LevelSilent},
		})
	})

}

type logMethodCase struct {
	call      func(*logger.Logger, logger.LogMessage)
	label     string
	visibleAt []logger.LogLevel
	hiddenAt  []logger.LogLevel
}

// testLogMethod runs the same set of subtests for any of the logger methods.
func testLogMethod(test *testing.T, tc logMethodCase) {
	test.Helper()

	test.Run("Should print the message in the [date]["+tc.label+"] json format", func(test *testing.T) {
		appLogger := useLevel(test, tc.visibleAt[0])

		output := captureStdout(test, func() {
			tc.call(appLogger, logger.LogMessage{
				Action:  "user.create",
				Message: "user created",
				Data:    map[string]any{"id": 1},
			})
		})

		matches := logLinePattern.FindStringSubmatch(output)
		if matches == nil {
			test.Fatalf("Output %q does not match the expected format", output)
		}
		date, level, payload := matches[1], matches[2], matches[3]

		if _, err := time.Parse(isoDateFormat, date); err != nil {
			test.Errorf("Date %q is not in ISO format: %v", date, err)
		}
		if level != tc.label {
			test.Errorf("Level = %q, want %q", level, tc.label)
		}

		want := `{"action":"user.create","message":"user created","data":{"id":1}}`
		if payload != want {
			test.Errorf("Payload = %s, want %s", payload, want)
		}
	})

	for _, level := range tc.visibleAt {
		test.Run("Should print when level is "+string(level), func(test *testing.T) {
			appLogger := useLevel(test, level)

			output := captureStdout(test, func() {
				tc.call(appLogger, logger.LogMessage{Action: "main", Message: "visible"})
			})

			if output == "" {
				test.Errorf("Expected output with level %s, got nothing", level)
			}
		})
	}

	for _, level := range tc.hiddenAt {
		test.Run("Should not print when level is "+string(level), func(test *testing.T) {
			appLogger := useLevel(test, level)

			output := captureStdout(test, func() {
				tc.call(appLogger, logger.LogMessage{Action: "main", Message: "hidden"})
			})

			if output != "" {
				test.Errorf("Expected no output with level %s, got %q", level, output)
			}
		})
	}
}

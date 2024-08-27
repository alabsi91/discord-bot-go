package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type logStyles struct {
	Success      lipgloss.Style
	Warning      lipgloss.Style
	Error        lipgloss.Style
	Info         lipgloss.Style
	Tip          lipgloss.Style
	Log          lipgloss.Style
	PaddingStyle lipgloss.Style
}

type logLevel struct {
	Info    string
	Success string
	Warning string
	Error   string
	Fatal   string
}
type log struct {
	Style         logStyles
	Level         logLevel
	LogToFile     bool
	LogFilePath   string
	LogFileWriter *os.File
}

var Log = log{
	Style: logStyles{
		Success:      lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")),
		Warning:      lipgloss.NewStyle().Foreground(lipgloss.Color("#F1C40F")),
		Error:        lipgloss.NewStyle().Foreground(lipgloss.Color("#E74C3C")),
		Info:         lipgloss.NewStyle().Foreground(lipgloss.Color("#5eb9ff")),
		Tip:          lipgloss.NewStyle().Foreground(lipgloss.Color("#FF69B4")),
		Log:          lipgloss.NewStyle(),
		PaddingStyle: lipgloss.NewStyle().Faint(true),
	},

	Level: logLevel{
		Info:    "INFO",
		Success: "SUCCESS",
		Warning: "WARNING",
		Error:   "ERROR",
		Fatal:   "FATAL",
	},

	LogToFile:     false,
	LogFilePath:   "",
	LogFileWriter: nil,
}

const titleWidth = 13
const newlineSeparator = "- "
const newlinePaddingWidth = titleWidth + 3 // titleWidth + "|" + "|" + " "

// splitOnNewline splits a given string into leading newlines, content, and trailing newlines
func splitOnNewline(input string) (string, string, string) {
	// Check for leading newlines
	newlineStart := 0
	for newlineStart < len(input) && input[newlineStart] == '\n' {
		newlineStart++
	}

	// Check for trailing newlines
	newlineEnd := len(input)
	for newlineEnd > newlineStart && input[newlineEnd-1] == '\n' {
		newlineEnd--
	}

	// Return leading newlines, content, and trailing newlines
	return input[:newlineStart], input[newlineStart:newlineEnd], input[newlineEnd:]
}

// formatTitle formats a title with the given style
func formatTitle(title string, style lipgloss.Style) string {
	return style.Render("|") +
		lipgloss.NewStyle().Inherit(style).Reverse(true).Bold(true).Align(lipgloss.Center).Width(titleWidth).Render(title) +
		style.Render("|")
}

func (l log) printLog(title string, style lipgloss.Style, strs ...string) {
	title = formatTitle(title, style)

	fullString := strings.Join(strs, " ")
	leadingNewlines, msg, trailingNewlines := splitOnNewline(fullString)

	split := strings.Split(msg, "\n")

	// format new lines so that they are aligned
	var content string
	for i, s := range split {
		isFirst := i == 0
		if !isFirst {
			paddingString := strings.Repeat(newlineSeparator, newlinePaddingWidth/len(newlineSeparator))
			content += "\n" + l.Style.PaddingStyle.Render(paddingString)
		}
		content += style.Render(s)
	}

	fmt.Println(leadingNewlines+title, content+trailingNewlines)
}

func (l log) Success(strs ...string) {
	l.printLog("SUCCESS", l.Style.Success, strs...)
}

func (l log) Error(strs ...string) {
	l.printLog("ERROR", l.Style.Error, strs...)
}

// FATAL logs a fatal error and exits the program
func (l log) Fatal(strs ...string) {
	l.printLog("FATAL", l.Style.Error, strs...)
	os.Exit(1)
}

func (l log) Info(strs ...string) {
	l.printLog("INFO", l.Style.Info, strs...)
}

func (l log) Warning(strs ...string) {
	l.printLog("WARNING", l.Style.Warning, strs...)
}

func (l log) Tip(strs ...string) {
	l.printLog("Tip", l.Style.Tip, strs...)
}

func (l log) Log(strs ...string) {
	l.printLog("LOG", l.Style.Log, strs...)
}

func (l *log) SetLogToFile(enabled bool) {
	l.LogToFile = enabled
}

func (l *log) SetLogFilePath(path string) {
	l.LogFilePath = path
}

// Debug logs a debug message to the log file only (not to the console)
func (l *log) Debug(level string, strs ...string) {
	if !l.LogToFile {
		return
	}

	// Ensure the log file writer is initialized
	if l.LogFileWriter == nil {
		// Open the log file with append mode, create if it does not exist.
		file, err := os.OpenFile(l.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			Log.Error("\nLogger:", err.Error())
			return
		}
		l.LogFileWriter = file

		_, err = l.LogFileWriter.WriteString("\n--------------------\n\n")
		if err != nil {
			Log.Error("\nLogger:", err.Error())
		}
	}

	recordTime := time.Now().Format("2006-01-02 15:04:05")
	fullString := strings.Join(strs, " ")
	_, msg, _ := splitOnNewline(fullString)
	logMsg := fmt.Sprintf("[%s] [%s] %s", recordTime, level, msg)

	_, err := l.LogFileWriter.WriteString(logMsg + "\n")
	if err != nil {
		Log.Error("\nLogger:", err.Error())
	}
}

// FileExists checks if a file exists at the given path.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true // File exists
	}
	if os.IsNotExist(err) {
		return false // File does not exist
	}
	// For other errors (like permission issues), handle them accordingly
	return false
}

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"gopkg.in/natefinch/lumberjack.v2"
)

var cfg Config

func main() {

	// Load config
	cfg = LoadConfig("SymProxy.json")
	fmt.Println(cfg)

	// Echo instance
	app := fiber.New(fiber.Config{
		CaseSensitive: false,
	})

	// Middleware
	app.Use(recover.New())

	// Logger
	app.Use(logger.New(logger.Config{
		TimeFormat: "2006-01-02 15:04:05.00000",
		Output: &lumberjack.Logger{
			Filename:   "./logs/SymProxy.log",
			MaxSize:    10, // megabytes
			MaxBackups: 10,
		},
	}))

	// Routes
	app.Get("/download/symbols/*", downloadSymbolsHandler)

	// Start server
	hostPort := cfg.Ip + ":" + cfg.Port
	err := app.Listen(hostPort)
	if err != nil {
		panic(err)
	}
}

func fileExist(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func downloadFileByUrl(url string, filePath string) error {
	// Prepare all dirs
	err := os.MkdirAll(path.Dir(filePath), os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directories: ", err)
		return err
	}

	// Donwload symbol file
	client := &http.Client{
		// Timeout for poor network
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("MS server error: %d", resp.StatusCode)
	}

	// Cache symbol file
	cacheFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer cacheFile.Close()

	_, err = io.Copy(cacheFile, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func securePath(dirPath string) string {
	cleanedPath := filepath.Clean(dirPath)
	return cleanedPath
}

func mergeSlashes(input string) string {
	re := regexp.MustCompile(`[\\\/]+`)
	return re.ReplaceAllString(input, "/")
}

// Route handler
func downloadSymbolsHandler(c fiber.Ctx) error {
	req := c.Request()

	urlStr := "http://msdl.microsoft.com" + string(req.RequestURI())
	fmt.Println("GET: ", urlStr)

	requestURILower := strings.ToLower(string(req.RequestURI()))
	shortSubPath := strings.TrimPrefix(requestURILower, "/download/symbols")
	filePath := cfg.Root + securePath(shortSubPath)
	filePath = mergeSlashes(filePath)
	fmt.Println("Path:", filePath)

	// Skip too long path for safety
	if len(filePath) > 255 {
		fmt.Println("File path too long. ", filePath)
		return fiber.NewError(http.StatusNotFound, "File path too long")
	}

	// Not found, download from url
	if !fileExist(filePath) {
		err := downloadFileByUrl(urlStr, filePath)
		if err != nil {
			fmt.Println("Download failed: ", err)
			return fiber.NewError(http.StatusNotFound, "Download failed")
		}
	}

	// Not found, report error
	if !fileExist(filePath) {
		fmt.Println("File Not Found")
		return fiber.NewError(http.StatusNotFound, "File Not Found")
	}

	// Found, send file to client
	return c.SendFile(filePath)
}

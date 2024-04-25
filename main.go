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

	// Fiber instance
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
	routeURI := cfg.Route + "*"
	app.Get(routeURI, downloadSymbolsHandler)
	fmt.Println("AddRoute:", routeURI)

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
	// Donwload symbol file
	client := &http.Client{
		// Timeout for poor network
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Failed to create request: ", err)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to request: ", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("error - symbols server status code: %d", resp.StatusCode)
	}

	// Prepare all dirs
	err = os.MkdirAll(path.Dir(filePath), os.ModePerm)
	if err != nil {
		fmt.Println("Error create directories: ", err)
		return err
	}

	// Cache symbol file
	cacheFile, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error create file: ", err)
		return err
	}
	defer cacheFile.Close()

	_, err = io.Copy(cacheFile, resp.Body)
	if err != nil {
		fmt.Println("Error copy file: ", err)
		return err
	}

	return nil
}

func securePath(dirPath string) string {
	cleanedPath := filepath.Clean(dirPath)
	return mergeSlashes(cleanedPath)
}

func mergeSlashes(input string) string {
	re := regexp.MustCompile(`[\\\/]+`)
	return re.ReplaceAllString(input, "/")
}

// Route handler
func downloadSymbolsHandler(c fiber.Ctx) error {
	req := c.Request()

	requestURILower := strings.ToLower(string(req.RequestURI()))
	subRUI := strings.TrimPrefix(requestURILower, cfg.Route)
	securedSubURI := securePath(subRUI)

	// File path
	filePath := cfg.Root + "/" + securedSubURI
	filePath = mergeSlashes(filePath)

	// Skip too long path for safety
	if len(filePath) > 255 {
		fmt.Println("File path too long. ", filePath)
		return fiber.NewError(http.StatusNotFound, "File path too long")
	}

	// Not found, download from url
	if !fileExist(filePath) {
		// URL
		urlStr := "http://msdl.microsoft.com/download/symbols/" + securedSubURI
		fmt.Println("GET: ", urlStr)

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
	fmt.Println("Path:", filePath)
	return c.SendFile(filePath)
}

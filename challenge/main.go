package main

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// generate a random string
func generateRandomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	password_real := generateRandomString(4)
	flag_file := "flag.txt"

	// Write the password_real to .env
	godotenv.Write(map[string]string{
		"PASSWORD": password_real,
	}, ".env")

	// Get current files in dir that match flag_*.txt
	files, err := os.ReadDir("./")
	if err != nil {
		log.Fatal("Error reading directory")
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "flag_") && strings.HasSuffix(file.Name(), ".txt") {
			flag_file = file.Name()
		}
	}

	fmt.Println(flag_file)

	flag_txt, err := os.ReadFile(flag_file)
	if err != nil {
		log.Fatal("Error reading flag.txt file")
	}

	inputFile := "./static/horse.jpeg"
	ffmpeg.LogCompiledCommand = false // be quiet

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := fiber.New()

	app.Get("/top-secret/nobody/will/ever/find-1337/this/:password", func(c *fiber.Ctx) error {
		if c.Params("password") == password_real {
			return c.SendString(string(flag_txt))
		}

		return c.JSON(fiber.Map{
			"error": "Invalid password",
		})
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./static/index.html")
	})

	app.Get("/robots.txt", func(c *fiber.Ctx) error {
		c.Response().Header.Set("Content-Type", "text/plain")
		return c.Send([]byte("# Note to self: Don't let anyone see the sceret file.!\nUser-agent: *\n\nDisallow: /dev-notepad"))
	})

	app.Get("/img", func(c *fiber.Ctx) error {
		c.Response().Header.Set("Content-Type", "image/jpeg")
		return c.SendFile("./static/horse.jpeg")
	})

	app.Get("/dev-notepad", func(c *fiber.Ctx) error {
		return c.SendFile("./static/notepad.html")
	})

	app.Get("/create-watermark", func(c *fiber.Ctx) error {
		// get ffmpeg version on the server
		ffmpeg_version_bytes, err := exec.Command("ffmpeg", "-version").Output()
		if err != nil {
			return c.JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		ffmpeg_version := string(ffmpeg_version_bytes)
		// Output is too long - get the first line only
		ffmpeg_version = strings.Split(ffmpeg_version, "\n")[0]

		fontcolor := c.Query("fontcolor", "black")
		fontsize_str := c.Query("fontsize", "10")
		fontsize, err := strconv.Atoi(fontsize_str)
		if err != nil {
			return c.JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if c.Query("text") != "" {
			text := c.Query("text")
			buf := bytes.NewBuffer(nil)
			cmd := ffmpeg.Input(inputFile).
				Filter("drawtext", nil, ffmpeg.KwArgs{
					"text":      text,
					"fontsize":  fontsize,
					"fontcolor": fontcolor,
					"x":         10,
					"y":         10,
				}).
				Filter("drawtext", nil, ffmpeg.KwArgs{
					"text":      fmt.Sprintf("Powered by %s", ffmpeg_version),
					"fontsize":  5,
					"fontcolor": "white",
					"x":         140,
					"y":         470,
				}).
				Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
				WithOutput(buf)

			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
				return c.JSON(fiber.Map{
					"error": err.Error(),
				})
			}
			c.Response().Header.Set("Cache-Control", "no-cache")
			c.Response().Header.Set("Content-Type", "image/jpg")
			return c.SendStream(buf)
		} else if c.Query("textfile") != "" {
			_, err := os.Stat(c.Query("textfile"))
			if err != nil {
				return c.JSON(fiber.Map{
					"error": err.Error(),
				})
			}

			text := c.Query("textfile")
			buf := bytes.NewBuffer(nil)
			cmd := ffmpeg.Input(inputFile).
				Filter("drawtext", nil, ffmpeg.KwArgs{
					"textfile":  text,
					"fontsize":  fontsize,
					"fontcolor": fontcolor,
					"x":         10,
					"y":         10,
				}).
				Filter("drawtext", nil, ffmpeg.KwArgs{
					"text":      fmt.Sprintf("Powered by %s", ffmpeg_version),
					"fontsize":  5,
					"fontcolor": "white",
					"x":         140,
					"y":         470,
				}).
				Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
				WithOutput(buf)

			err = cmd.Run()
			if err != nil {
				fmt.Println(err)
				return c.JSON(fiber.Map{
					"error": err.Error(),
				})
			}

			c.Response().Header.Set("Cache-Control", "no-cache")
			c.Response().Header.Set("Content-Type", "image/jpg")
			return c.SendStream(buf)
		}
		return c.JSON(fiber.Map{
			"error": errors.New("no watermark text-based parameters were provided").Error(),
		})
	})

	err = app.Listen(":1337")
	if err != nil {
		panic(err)
	}
}

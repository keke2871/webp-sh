package main

import (
	"fmt"
	"webp-sh/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
)

var app = fiber.New(fiber.Config{
	ServerHeader:          "WebP Server Go",
	AppName:               "WebP Server Go",
	DisableStartupMessage: true,
	ProxyHeader:           "X-Real-IP",
	ReadBufferSize:        4096,  // per-connection buffer size for requests' reading. This also limits the maximum header size. Increase this buffer if your clients send multi-KB RequestURIs and/or multi-KB headers (for example, BIG cookies).
	Concurrency:           20,    // Maximum number of concurrent connections.
	DisableKeepalive:      false, // Disable keep-alive connections, the server will close incoming connections after sending the first response to the client
})

func main() {
	fmt.Println("Webp start ...")
	listenAddress := "127.0.0.1:9000"
	app.Use(etag.New(etag.Config{
		Weak: true,
	}))

	app.Get("/healthz", handler.Healthz)

	fmt.Printf("WebP Server Go is Running on http://%s\n", listenAddress)

	_ = app.Listen(listenAddress)

}

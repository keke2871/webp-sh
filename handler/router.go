package handler

import (
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"webp-sh/config"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

// Convert image schema to webp
func Convert(c *fiber.Ctx) error {

	// this function need to do:
	// 1. get request path, query string
	// 2. generate rawImagePath, could be local path or remote url(possible with query string)
	// 3. pass it to encoder, get the result, send it back

	// normal http request will start with /
	log.Infof("c.Path-->>%s\n", c.Path())
	if !strings.HasPrefix(c.Path(), "/") {
		_ = c.SendStatus(http.StatusBadRequest)
		return nil
	}

	var (
		reqHostname = c.Hostname()
		reqHost     = c.Protocol() + "://" + reqHostname // http://www.example.com:8000
		reqHeader   = &c.Request().Header

		reqURIRaw, _          = url.QueryUnescape(c.Path())        // /mypic/123.jpg
		reqURIwithQueryRaw, _ = url.QueryUnescape(c.OriginalURL()) // /mypic/123.jpg?someother=200&somebugs=200
		reqURI                = path.Clean(reqURIRaw)              // delete ../ in reqURI to mitigate directory traversal
		reqURIwithQuery       = path.Clean(reqURIwithQueryRaw)     // Sometimes reqURIwithQuery can be https://example.tld/mypic/123.jpg?someother=200&somebugs=200, we need to extract it

		filename       = path.Base(reqURI)
		realRemoteAddr = ""
		targetHostName = config.LocalHostAlias
		targetHost     = config.Config.ImgPath
		proxyMode      = config.ProxyMode
		mapMode        = false

		width, _     = strconv.Atoi(c.Query("width"))      // Extra Params
		height, _    = strconv.Atoi(c.Query("height"))     // Extra Params
		maxHeight, _ = strconv.Atoi(c.Query("max_height")) // Extra Params
		maxWidth, _  = strconv.Atoi(c.Query("max_width"))  // Extra Params
		// extraParams,
		_ = config.ExtraParams{
			Width:     width,
			Height:    height,
			MaxWidth:  maxWidth,
			MaxHeight: maxHeight,
		}
	)

	log.Debugf("Incoming connection from %s %s %s", c.IP(), reqHostname, reqURIwithQuery)
	log.Debugf("Header %s, header=%s, filename=%s, realRemoteAddr=%s, targetHostName=%s, targetHost=%s,proxyMode=%s,mapMode=%v", reqHost, reqHeader, filename, realRemoteAddr, targetHostName, targetHost, proxyMode, mapMode)

	return c.SendString("WebP Server Router running!")
}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type CaddyLog struct {
	Level   string  `json:"level"`
	TS      float64 `json:"ts"`
	Request struct {
		ClientIP string `json:"client_ip"`
		Proto    string `json:"proto"`
		Method   string `json:"method"`
		Host     string `json:"host"`
		URI      string `json:"uri"`
		Headers  struct {
			Referer   []string `json:"Referer"`
			UserAgent []string `json:"User-Agent"`
		} `json:"headers"`
	} `json:"request"`
	UserID string `json:"user_id"`
	Size   int    `json:"size"`
	Status int    `json:"status"`
}

func (c *CaddyLog) toCombinedLogFormat() string {
	clientIP := c.Request.ClientIP

	logname := "-"

	user := c.UserID
	if user == "" {
		user = "-"
	}

	sec, dec := int64(c.TS), int64((c.TS-float64(int64(c.TS)))*1e9)
	ts := time.Unix(sec, dec)
	timestamp := ts.Format("02/Jan/2006:15:04:05 -0700")

	requestLine := fmt.Sprintf("%s %s %s", c.Request.Method, c.Request.URI, c.Request.Proto)

	status := c.Status

	size := c.Size

	referer := "-"
	if len(c.Request.Headers.Referer) > 0 {
		referer = c.Request.Headers.Referer[0]
	}

	userAgent := "-"
	if len(c.Request.Headers.UserAgent) > 0 {
		userAgent = c.Request.Headers.UserAgent[0]
	}

	return fmt.Sprintf(`%s %s %s [%s] "%s" %d %d "%s" "%s"`,
		clientIP,
		logname,
		user,
		timestamp,
		requestLine,
		status,
		size,
		referer,
		userAgent,
	)
}

func processLogs(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Bytes()
		var logEntry CaddyLog

		if err := json.Unmarshal(line, &logEntry); err != nil {
			fmt.Fprintf(os.Stderr, "error parsing JSON: %v\n", err)
			continue
		}

		fmt.Println(logEntry.toCombinedLogFormat())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading source: %v\n", err)
	}
}

func main() {
	if len(os.Args) > 1 {
		for _, filename := range os.Args[1:] {
			file, err := os.Open(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error opening file %s: %v\n", filename, err)
				continue
			}
			processLogs(file)
			file.Close()
		}
	} else {
		processLogs(os.Stdin)
	}
}

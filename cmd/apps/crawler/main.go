package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func logAction(msg string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		log.Println(msg)
		return nil
	}
}

func main() {
	var (
		sites = []string{
			"https://tree.test:4044/",
			//"https://thetreecenter.com",
			//"https://google.com",
		}
	)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// Set headless option
		chromedp.Flag("headless", true),
		//chromedp.Flag("headless", false),
		// Disable SSL certificate error checks
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.WindowSize(1920, 1080),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	for _, site := range sites {
		// strips the protocol and replaces / with _
		file := strings.Replace(strings.Replace(site, "://", "", 1), "/", "_", -1)
		// also replace :
		file = strings.Replace(file, ":", "_", -1)
		if err := screenshot(ctx, site, fmt.Sprintf("%s.png", file)); err != nil {
			log.Fatalf("failed to take screenshot of %s: %v", site, err)
		}
	}
}

func screenshot(ctx context.Context, urlStr, file string) error {
	var buf []byte

	if err := chromedp.Run(ctx,
		logAction("Starting Chrome"),
		chromedp.Navigate(urlStr),
		logAction("Navigated to "+urlStr),
		chromedp.WaitReady(`body`, chromedp.BySearch),
		logAction("Body is ready"),
		chromedp.FullScreenshot(&buf, 100),
		//chromedp.Screenshot(`body`, &buf, chromedp.NodeVisible, chromedp.BySearch),
		logAction("Screenshot taken"),
	); err != nil {
		return err
	}

	return os.WriteFile(file, buf, 0644)
}

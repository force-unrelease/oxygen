package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/test", ScreenShot)
	// mux.HandleFunc("/api/falcon/", falcon2img)
	// mux.HandleFunc("/api/gitlab-ci/")
	// mux.HandleFunc("/api/prometheus/")
	// mux.HandleFunc("/api/grafana/")

	srv := &http.Server{
		Addr:    "127.0.0.1:8090",
		Handler: mux,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}

func genImg(w http.ResponseWriter, r *http.Request) {
	// id := uuid.New().String()
	err := ioutil.WriteFile("/tmp/dat1", []byte{}, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ScreenShot(w http.ResponseWriter, r *http.Request) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.Flag("headless", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	// create context
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	// capture screenshot of an element
	var buf []byte
	if err := chromedp.Run(ctx,
		setCookie("falconSessionId", "b8bf552a2c04e4330ce25297ba8f5f42", "falcon.dun.mi.com", "/", true, false),
		elementScreenshot(`https://s.mi.cn/nPQVKko2`, `#core-view`, &buf)); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("elementScreenshot.png", buf, 0644); err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "https://www.google.com/ screenshot is captured!")
}

// elementScreenshot takes a screenshot of a specific element.
func elementScreenshot(urlstr, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.WaitVisible(sel, chromedp.ByID),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible, chromedp.ByID),
	}
}

func setCookie(name, value, domain, path string, httpOnly, secure bool) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
		success, err := network.SetCookie(name, value).
			WithExpires(&expr).
			WithDomain(domain).
			WithPath(path).
			WithHTTPOnly(httpOnly).
			WithSecure(secure).
			Do(ctx)
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("could not set cookie %s", name)
		}
		return nil
	})
}

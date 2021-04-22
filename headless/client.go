package headless

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/page"
	cdpruntime "github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"

	"chromedpsample/utils"
)

// Counter keep number files are dowloading
type Counter struct {
	counter uint64
	mutex   sync.Mutex
}

// Value -
func (c *Counter) Value() uint64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.counter
}

// Increase -
func (c *Counter) Increase() {
	c.mutex.Lock()
	c.counter++
	c.mutex.Unlock()
}

// Decrease -
func (c *Counter) Decrease() {
	c.mutex.Lock()
	c.counter--
	c.mutex.Unlock()
}

type Client struct {
	browserCtx context.Context
}

func NewClient() *Client {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	// create a new browser
	if os.Getenv("HEADLESS_ENV") == "container" {
		opts := []chromedp.ExecAllocatorOption{
			chromedp.ExecPath("/headless-shell/headless-shell"),
		}

		ctx, _ = chromedp.NewExecAllocator(context.Background(), opts...)
		ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf)) //chromedp.WithLogf(log.Printf))
	} else {
		ctx, cancel = chromedp.NewContext(context.Background(), chromedp.WithDebugf(log.Printf)) //chromedp.WithLogf(log.Printf))
	}

	if err := chromedp.Run(ctx); err != nil {
		cancel()
		log.Fatal("Failed to start browser:", err)
	}

	return &Client{
		browserCtx: ctx,
	}
}

func (c *Client) Close() {
	if c.browserCtx != nil {
		if err := chromedp.Cancel(c.browserCtx); err != nil {
			err = errors.Wrap(err, "failed to close browser")
			log.Println(err)
		}
	}
}

func (c *Client) Download(ID, dlURL, dlDir string) error {
	var (
		errChan = make(chan error, 1)
		counter Counter
	)

	if c.browserCtx == nil {
		err := errors.Errorf("Not start browser")
		log.Println(err)
		return err
	}

	log.Printf("ID:%s, browser has %d tabs", ID, c.getNumOfTabs())

	uri, err := url.ParseRequestURI(dlURL)
	if err != nil {
		err = errors.Wrapf(err, "ID:%s, URL %s has query params invalid", ID, dlURL)
		log.Println(err)
		return err
	}

	// Create dowload dir if not exist
	if err = os.MkdirAll(dlDir, 0755); err != nil {
		err = errors.Wrapf(err, "ID:%s failed to create download dir: %s", ID, dlDir)
		log.Println(err)
		return err
	}

	// Download dir have to absolute path
	absDir, err := filepath.Abs(dlDir)
	if err != nil {
		err = errors.Wrapf(err, "ID:%s, download dir not found %s", ID, dlDir)
		log.Println(err)
		return err
	}
	dlDir = absDir

	// Create a timeout as a safety net to prevent any infinite wait loops
	duration := 5 * time.Minute

	// create a new tab
	ctxTimeout, cancel := context.WithTimeout(c.browserCtx, duration)
	defer cancel()
	ctxTab, cancel := chromedp.NewContext(ctxTimeout)
	defer func() {
		cancel()
		if err := chromedp.Cancel(ctxTab); err != nil {
			err = errors.Wrapf(err, "ID:%s failed to close tab", ID)
			log.Println(err)
		}
	}()

	chromedp.ListenTarget(ctxTab, func(ev interface{}) {
		switch ev := ev.(type) {
		case *cdpruntime.EventConsoleAPICalled:
			data, _ := ev.MarshalJSON()
			details := string(data)
			log.Printf("ID:%s received event EventConsoleAPICalled:%s", ID, details)
			if ev.Type == "error" {
				if len(ev.Args) > 0 && ev.Args[0].Description != "" {
					errChan <- errors.New(string(ev.Args[0].Description))
				} else {
					errChan <- errors.New("Received unknown error")
				}
			}

		case *cdpruntime.EventExceptionThrown:
			data, _ := ev.MarshalJSON()
			details := string(data)
			log.Printf("ID:%s received event EventExceptionThrown:%s", ID, details)
			errChan <- errors.New(string(ev.ExceptionDetails.Exception.Value))

		case *target.EventTargetCrashed:
			data, _ := ev.MarshalJSON()
			details := string(data)
			log.Printf("ID:%s received event EventTargetCrashed:%s", ID, details)
			errChan <- errors.New(string(ev.Status))

		case *page.EventDownloadWillBegin:
			data, _ := ev.MarshalJSON()
			details := string(data)
			log.Printf("ID:%s download start:%s", ID, details)

			// Increase download counter when download start
			counter.Increase()

		case *page.EventDownloadProgress:
			data, _ := ev.MarshalJSON()
			details := string(data)

			switch ev.State {
			case page.DownloadProgressStateCanceled, page.DownloadProgressStateCompleted:
				if ev.State == page.DownloadProgressStateCanceled {
					log.Printf("ID:%s download canceled:%s", ID, details)
				} else {
					log.Printf("ID:%s download completed:%s", ID, details)
				}

				chromedp.Sleep(30 * time.Second)

				// Decrease download counter after download ended
				counter.Decrease()

				// Notify main routine download progress is done
				if counter.Value() == 0 {
					errChan <- nil
				}
			}
		}
	})

	err = chromedp.Run(
		ctxTab,
		page.SetDownloadBehavior(page.SetDownloadBehaviorBehaviorAllow).WithDownloadPath(dlDir),
		chromedp.Navigate(dlURL),
	)
	if err != nil && !strings.Contains(err.Error(), "net::ERR_ABORTED") {
		err = errors.Wrapf(err, "ID:%s failed to download file at URL %s", ID, dlURL)
		log.Println(err)
		return err
	}

	// Wait for the download to finish
	select {
	case <-ctxTimeout.Done():
		err = fmt.Errorf("Download not finish in %s", duration.String())
	case err = <-errChan:
	}

	if err != nil {
		err = errors.Wrapf(err, "ID:%s downloaded failed at URL:%s", ID, uri.RawQuery)
		log.Println(err)
	} else {
		// Check there are any files in download dir
		files, err := utils.ListFilesInDir(dlDir)
		if err == nil && len(files) > 0 {
			log.Printf("ID:%s, download dir:%s has %d files: %v", ID, dlDir, len(files), strings.Join(files, ","))
		} else {
			err = errors.Errorf("ID:%s not found any file in download dir: %s", ID, dlDir)
			log.Println(err)
			return err
		}
	}

	return err
}

func (c *Client) getNumOfTabs() int {
	infos, err := chromedp.Targets(c.browserCtx)
	if err != nil {
		err = errors.Wrap(err, "Failed to lists all the targets in the browser")
		log.Println(err)
		return 0
	}

	var pages []*target.Info
	for _, info := range infos {
		if info.Type == "page" {
			pages = append(pages, info)
		}
	}

	return len(pages)
}

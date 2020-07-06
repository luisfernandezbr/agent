package util

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/pkg/browser"
)

// WaitForRedirect will open a url with a `redirect_to` query string param that gets handled by handler
func WaitForRedirect(rawURL string, handler func(w http.ResponseWriter, r *http.Request)) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return fmt.Errorf("error listening to port: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	q := u.Query()
	q.Set("redirect_to", fmt.Sprintf("http://localhost:%d/", port))
	u.RawQuery = q.Encode()

	done := make(chan bool, 1)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
		done <- true
	})

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}
	defer server.Close()
	go server.Serve(listener)

	if err := browser.OpenURL(u.String()); err != nil {
		return fmt.Errorf("error opening url: %w", err)
	}

	<-done
	return nil
}

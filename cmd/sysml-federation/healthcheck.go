package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// healthcheck GETs each URL and fails on the first that does not answer
// 200. The image's HEALTHCHECK runs it against the router's readiness path
// and the viewer (C-55, C-56).
func healthcheck(ctx context.Context, urls ...string) error {
	client := &http.Client{Timeout: 3 * time.Second}
	for _, u := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s answered %d", u, resp.StatusCode)
		}
	}
	return nil
}

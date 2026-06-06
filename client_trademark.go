package dpmaconnect

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/patent-dev/dpma-connect-plus/generated"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// SearchTrademarks executes a trademark expert search query
func (c *Client) SearchTrademarks(ctx context.Context, query string) ([]byte, error) {
	resp, err := c.generated.SearchTrademarksWithResponse(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search trademarks: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "search failed")
}

// GetTrademarkInfo retrieves trademark information by application number
func (c *Client) GetTrademarkInfo(ctx context.Context, applicationNumber string) ([]byte, error) {
	resp, err := c.generated.GetTrademarkInfoWithResponse(ctx, applicationNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get trademark info: %w", err)
	}
	return resourceResult(resp.Body, resp.StatusCode(), "trademark info", applicationNumber, "failed to get trademark info")
}

// GetTrademarkImage downloads a trademark image by application number
func (c *Client) GetTrademarkImage(ctx context.Context, applicationNumber string) ([]byte, error) {
	resp, err := c.generated.GetTrademarkImageWithResponse(ctx, applicationNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get trademark image: %w", err)
	}
	return resourceResult(resp.Body, resp.StatusCode(), "trademark image", applicationNumber, "failed to get trademark image")
}

// GetTrademarkThumbnail downloads a trademark thumbnail by application number
func (c *Client) GetTrademarkThumbnail(ctx context.Context, applicationNumber string) ([]byte, error) {
	resp, err := c.generated.GetTrademarkThumbnailWithResponse(ctx, applicationNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get trademark thumbnail: %w", err)
	}
	return resourceResult(resp.Body, resp.StatusCode(), "trademark thumbnail", applicationNumber, "failed to get trademark thumbnail")
}

// GetTrademarkBibDataApplied downloads trademark bibliographic data (applied) for a publication week
func (c *Client) GetTrademarkBibDataApplied(ctx context.Context, year, week int) ([]byte, error) {
	return fetchWeeklyBulk(year, week,
		"failed to get trademark bib data (applied)", "failed to download trademark bib data (applied)",
		func(pw string) (*generated.GetTrademarkBibDataAppliedResponse, error) {
			return c.generated.GetTrademarkBibDataAppliedWithResponse(ctx, pw)
		},
		func(r *generated.GetTrademarkBibDataAppliedResponse) []byte { return r.Body })
}

// GetTrademarkBibDataRegistered downloads trademark bibliographic data (registered) for a publication week
func (c *Client) GetTrademarkBibDataRegistered(ctx context.Context, year, week int) ([]byte, error) {
	return fetchWeeklyBulk(year, week,
		"failed to get trademark bib data (registered)", "failed to download trademark bib data (registered)",
		func(pw string) (*generated.GetTrademarkBibDataRegisteredResponse, error) {
			return c.generated.GetTrademarkBibDataRegisteredWithResponse(ctx, pw)
		},
		func(r *generated.GetTrademarkBibDataRegisteredResponse) []byte { return r.Body })
}

// GetTrademarkBibDataRejected downloads trademark bibliographic data (rejected) for a publication week
func (c *Client) GetTrademarkBibDataRejected(ctx context.Context, year, week int) ([]byte, error) {
	return fetchWeeklyBulk(year, week,
		"failed to get trademark bib data (rejected)", "failed to download trademark bib data (rejected)",
		func(pw string) (*generated.GetTrademarkBibDataRejectedResponse, error) {
			return c.generated.GetTrademarkBibDataRejectedWithResponse(ctx, pw)
		},
		func(r *generated.GetTrademarkBibDataRejectedResponse) []byte { return r.Body })
}

// GetTrademarkRegisterExtract downloads trademark register extract data for a date and period
func (c *Client) GetTrademarkRegisterExtract(ctx context.Context, date time.Time, period string) ([]byte, error) {
	if err := ValidatePeriod(period); err != nil {
		return nil, err
	}
	d := openapi_types.Date{Time: date}
	resp, err := c.generated.GetTrademarkRegisterExtractWithResponse(ctx, d, generated.GetTrademarkRegisterExtractParamsPeriod(period))
	if err != nil {
		return nil, fmt.Errorf("failed to get trademark register extract: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download trademark register extract")
}

// SearchTrademarksParsed executes a trademark search and returns parsed results.
func (c *Client) SearchTrademarksParsed(ctx context.Context, query string) (*TrademarkSearchResult, error) {
	data, err := c.SearchTrademarks(ctx, query)
	if err != nil {
		return nil, err
	}
	return ParseTrademarkSearch(data)
}

// GetTrademarkInfoParsed retrieves trademark info and returns parsed data.
func (c *Client) GetTrademarkInfoParsed(ctx context.Context, applicationNumber string) (*TrademarkInfo, error) {
	data, err := c.GetTrademarkInfo(ctx, applicationNumber)
	if err != nil {
		return nil, err
	}
	return ParseTrademarkInfo(data)
}

// GetTrademarkBibDataAppliedStream downloads trademark bib data (applied) and writes to dst
func (c *Client) GetTrademarkBibDataAppliedStream(ctx context.Context, year, week int, dst io.Writer) error {
	return streamWeekly(year, week, "failed to get trademark bib data (applied)", dst,
		func(pw string) (*http.Response, error) { return c.generated.GetTrademarkBibDataApplied(ctx, pw) })
}

// GetTrademarkBibDataRegisteredStream downloads trademark bib data (registered) and writes to dst
func (c *Client) GetTrademarkBibDataRegisteredStream(ctx context.Context, year, week int, dst io.Writer) error {
	return streamWeekly(year, week, "failed to get trademark bib data (registered)", dst,
		func(pw string) (*http.Response, error) { return c.generated.GetTrademarkBibDataRegistered(ctx, pw) })
}

// GetTrademarkBibDataRejectedStream downloads trademark bib data (rejected) and writes to dst
func (c *Client) GetTrademarkBibDataRejectedStream(ctx context.Context, year, week int, dst io.Writer) error {
	return streamWeekly(year, week, "failed to get trademark bib data (rejected)", dst,
		func(pw string) (*http.Response, error) { return c.generated.GetTrademarkBibDataRejected(ctx, pw) })
}

// GetTrademarkRegisterExtractStream downloads trademark register extract data and writes to dst
func (c *Client) GetTrademarkRegisterExtractStream(ctx context.Context, date time.Time, period string, dst io.Writer) error {
	if err := ValidatePeriod(period); err != nil {
		return err
	}
	d := openapi_types.Date{Time: date}
	resp, err := c.generated.GetTrademarkRegisterExtract(ctx, d, generated.GetTrademarkRegisterExtractParamsPeriod(period))
	return streamResponse(resp, err, "failed to get trademark register extract", dst)
}

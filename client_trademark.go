package dpmaconnect

import (
	"context"
	"fmt"
	"io"
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
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return nil, err
	}
	resp, err := c.generated.GetTrademarkBibDataAppliedWithResponse(ctx, pubWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get trademark bib data (applied): %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download trademark bib data (applied)")
}

// GetTrademarkBibDataRegistered downloads trademark bibliographic data (registered) for a publication week
func (c *Client) GetTrademarkBibDataRegistered(ctx context.Context, year, week int) ([]byte, error) {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return nil, err
	}
	resp, err := c.generated.GetTrademarkBibDataRegisteredWithResponse(ctx, pubWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get trademark bib data (registered): %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download trademark bib data (registered)")
}

// GetTrademarkBibDataRejected downloads trademark bibliographic data (rejected) for a publication week
func (c *Client) GetTrademarkBibDataRejected(ctx context.Context, year, week int) ([]byte, error) {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return nil, err
	}
	resp, err := c.generated.GetTrademarkBibDataRejectedWithResponse(ctx, pubWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get trademark bib data (rejected): %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download trademark bib data (rejected)")
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
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return err
	}
	resp, err := c.generated.GetTrademarkBibDataApplied(ctx, pubWeek)
	return streamResponse(resp, err, "failed to get trademark bib data (applied)", dst)
}

// GetTrademarkBibDataRegisteredStream downloads trademark bib data (registered) and writes to dst
func (c *Client) GetTrademarkBibDataRegisteredStream(ctx context.Context, year, week int, dst io.Writer) error {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return err
	}
	resp, err := c.generated.GetTrademarkBibDataRegistered(ctx, pubWeek)
	return streamResponse(resp, err, "failed to get trademark bib data (registered)", dst)
}

// GetTrademarkBibDataRejectedStream downloads trademark bib data (rejected) and writes to dst
func (c *Client) GetTrademarkBibDataRejectedStream(ctx context.Context, year, week int, dst io.Writer) error {
	pubWeek, err := FormatPublicationWeek(year, week)
	if err != nil {
		return err
	}
	resp, err := c.generated.GetTrademarkBibDataRejected(ctx, pubWeek)
	return streamResponse(resp, err, "failed to get trademark bib data (rejected)", dst)
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

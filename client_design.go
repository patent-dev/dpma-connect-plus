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

// SearchDesigns executes a design expert search query
func (c *Client) SearchDesigns(ctx context.Context, query string) ([]byte, error) {
	resp, err := c.generated.SearchDesignsWithResponse(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search designs: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "search failed")
}

// GetDesignInfo retrieves design information by design number
func (c *Client) GetDesignInfo(ctx context.Context, designNumber string) ([]byte, error) {
	resp, err := c.generated.GetDesignInfoWithResponse(ctx, designNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get design info: %w", err)
	}
	return resourceResult(resp.Body, resp.StatusCode(), "design info", designNumber, "failed to get design info")
}

// GetDesignImage downloads a design image by design number and image number
func (c *Client) GetDesignImage(ctx context.Context, designNumber, imageNumber string) ([]byte, error) {
	resp, err := c.generated.GetDesignImageWithResponse(ctx, designNumber, imageNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get design image: %w", err)
	}
	return resourceResult(resp.Body, resp.StatusCode(), "design image", designNumber+"/"+imageNumber, "failed to get design image")
}

// GetDesignThumbnail downloads a design thumbnail by design number and thumbnail number
func (c *Client) GetDesignThumbnail(ctx context.Context, designNumber, thumbnailNumber string) ([]byte, error) {
	resp, err := c.generated.GetDesignThumbnailWithResponse(ctx, designNumber, thumbnailNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get design thumbnail: %w", err)
	}
	return resourceResult(resp.Body, resp.StatusCode(), "design thumbnail", designNumber+"/"+thumbnailNumber, "failed to get design thumbnail")
}

// GetDesignBibliographicDataXML downloads design bibliographic data as XML for a publication week
func (c *Client) GetDesignBibliographicDataXML(ctx context.Context, year, week int) ([]byte, error) {
	return fetchWeeklyBulk(year, week,
		"failed to get design bibliographic data", "failed to download design bibliographic data",
		func(pw string) (*generated.GetDesignBibliographicDataXMLResponse, error) {
			return c.generated.GetDesignBibliographicDataXMLWithResponse(ctx, pw)
		},
		func(r *generated.GetDesignBibliographicDataXMLResponse) []byte { return r.Body })
}

// GetDesignImages downloads design images for a publication week
func (c *Client) GetDesignImages(ctx context.Context, year, week int) ([]byte, error) {
	return fetchWeeklyBulk(year, week,
		"failed to get design images", "failed to download design images",
		func(pw string) (*generated.GetDesignImagesResponse, error) {
			return c.generated.GetDesignImagesWithResponse(ctx, pw)
		},
		func(r *generated.GetDesignImagesResponse) []byte { return r.Body })
}

// GetDesignRegisterExtract downloads design register extract data for a date and period
func (c *Client) GetDesignRegisterExtract(ctx context.Context, date time.Time, period string) ([]byte, error) {
	if err := ValidatePeriod(period); err != nil {
		return nil, err
	}
	d := openapi_types.Date{Time: date}
	resp, err := c.generated.GetDesignRegisterExtractWithResponse(ctx, d, generated.GetDesignRegisterExtractParamsPeriod(period))
	if err != nil {
		return nil, fmt.Errorf("failed to get design register extract: %w", err)
	}
	return bulkResult(resp.Body, resp.StatusCode(), "failed to download design register extract")
}

// SearchDesignsParsed executes a design search and returns parsed results.
func (c *Client) SearchDesignsParsed(ctx context.Context, query string) (*DesignSearchResult, error) {
	data, err := c.SearchDesigns(ctx, query)
	if err != nil {
		return nil, err
	}
	return ParseDesignSearch(data)
}

// GetDesignInfoParsed retrieves design info and returns parsed data.
func (c *Client) GetDesignInfoParsed(ctx context.Context, designNumber string) (*DesignInfo, error) {
	data, err := c.GetDesignInfo(ctx, designNumber)
	if err != nil {
		return nil, err
	}
	return ParseDesignInfo(data)
}

// GetDesignBibliographicDataXMLStream downloads design bibliographic data as XML and writes to dst
func (c *Client) GetDesignBibliographicDataXMLStream(ctx context.Context, year, week int, dst io.Writer) error {
	return streamWeekly(year, week, "failed to get design bibliographic data", dst,
		func(pw string) (*http.Response, error) { return c.generated.GetDesignBibliographicDataXML(ctx, pw) })
}

// GetDesignImagesStream downloads design images and writes to dst
func (c *Client) GetDesignImagesStream(ctx context.Context, year, week int, dst io.Writer) error {
	return streamWeekly(year, week, "failed to get design images", dst,
		func(pw string) (*http.Response, error) { return c.generated.GetDesignImages(ctx, pw) })
}

// GetDesignRegisterExtractStream downloads design register extract data and writes to dst
func (c *Client) GetDesignRegisterExtractStream(ctx context.Context, date time.Time, period string, dst io.Writer) error {
	if err := ValidatePeriod(period); err != nil {
		return err
	}
	d := openapi_types.Date{Time: date}
	resp, err := c.generated.GetDesignRegisterExtract(ctx, d, generated.GetDesignRegisterExtractParamsPeriod(period))
	return streamResponse(resp, err, "failed to get design register extract", dst)
}

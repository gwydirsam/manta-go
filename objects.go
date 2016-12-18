package manta

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
)

// GetObjectInput represents parameters to a GetObject operation.
type GetObjectInput struct {
	ObjectPath string
}

// GetObjectOutput contains the outputs for a GetObject operation. It is your
// responsibility to ensure that the io.ReadCloser ObjectReader is closed.
type GetObjectOutput struct {
	ContentLength uint64
	ContentType   string
	LastModified  time.Time
	ContentMD5    string
	ETag          string
	Metadata      map[string]string
	ObjectReader  io.ReadCloser
}

// GetObject retrieves an object from the Manta service. If error is nil (i.e.
// the call returns successfully), it is your responsibility to close the io.ReadCloser
// named ObjectReader in the operation output.
func (c *Client) GetObject(input *GetObjectInput) (*GetObjectOutput, error) {
	path := fmt.Sprintf("/%s/stor/%s", c.accountName, input.ObjectPath)

	respBody, respHeaders, err := c.executeRequest(http.MethodGet, path, nil, nil, nil)
	if err != nil {
		respBody.Close()
		return nil, errwrap.Wrapf("Error executing GetDirectory request: {{err}}", err)
	}

	response := &GetObjectOutput{
		ContentType:  respHeaders.Get("Content-Type"),
		ContentMD5:   respHeaders.Get("Content-MD5"),
		ETag:         respHeaders.Get("Etag"),
		ObjectReader: respBody,
	}

	lastModified, err := time.Parse(time.RFC1123, respHeaders.Get("Last-Modified"))
	if err == nil {
		response.LastModified = lastModified
	}

	contentLength, err := strconv.ParseUint(respHeaders.Get("Content-Length"), 10, 64)
	if err == nil {
		response.ContentLength = contentLength
	}

	metadata := map[string]string{}
	for key, values := range respHeaders {
		if strings.HasPrefix(key, "m-") {
			metadata[key] = strings.Join(values, ", ")
		}
	}
	response.Metadata = metadata

	return response, nil
}

// DeleteObjectInput represents parameters to a DeleteObject operation.
type DeleteObjectInput struct {
	ObjectPath string
}

// DeleteObject deletes an object.
func (c *Client) DeleteObject(input *DeleteObjectInput) error {
	path := fmt.Sprintf("/%s/stor/%s", c.accountName, input.ObjectPath)

	respBody, _, err := c.executeRequest(http.MethodDelete, path, nil, nil, nil)
	if respBody != nil {
		defer respBody.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing DeleteObject request: {{err}}", err)
	}

	return nil
}

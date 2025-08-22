//====================================================================================================
// Copyright (C) 2016-present ShIroRRen <http://shiror.ren>.                                         =
//                                                                                                   =
// Part of the NGA project.                                                                          =
// Licensed under the F2DLPR License.                                                                =
//                                                                                                   =
// YOU MAY NOT USE THIS FILE EXCEPT IN COMPLIANCE WITH THE LICENSE.                                  =
// Provided "AS IS", WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,                                   =
// unless required by applicable law or agreed to in writing.                                        =
//                                                                                                   =
// For the NGA project, visit: <http://app.niggergo.work>.                                           =
// For the F2DLPR License terms and conditions, visit: <http://license.fileto.download>.             =
//====================================================================================================

package nga

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type HttpReader struct {
	Url    string
	client *http.Client
	Size   int64
}

func NewHttpReader(url string) (*HttpReader, error) {
	client := &http.Client{}
	resp, err := client.Head(url)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	if resp.ContentLength < 0 {
		return nil, errors.New("ContentLength < 0")
	}
	return &HttpReader{Url: url, client: client, Size: resp.ContentLength}, nil
}

func (_reader *HttpReader) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= _reader.Size {
		return 0, io.EOF
	}
	end := off + int64(len(p)) - 1
	if end >= _reader.Size {
		end = _reader.Size - 1
	}
	req, err := http.NewRequest("GET", _reader.Url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", off, end))
	resp, err := _reader.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusPartialContent, http.StatusOK:
		return io.ReadFull(resp.Body, p)
	case http.StatusRequestedRangeNotSatisfiable:
		return 0, io.EOF
	default:
		return 0, fmt.Errorf("StatusCode %d", resp.StatusCode)
	}
}

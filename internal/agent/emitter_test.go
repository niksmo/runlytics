package agent

import (
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeClient struct {
	url, contentType string
}

type fakeReadCloser struct{}

func (frc fakeReadCloser) Close() error {
	return nil
}
func (frc fakeReadCloser) Read(b []byte) (int, error) {
	return 0, nil
}

func (fc *fakeClient) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	fc.url = url
	fc.contentType = contentType

	return &http.Response{StatusCode: 200, Body: fakeReadCloser{}}, nil
}

func TestHttpEmittingFunc(t *testing.T) {

	type want struct {
		url, contentType string
	}

	type args struct {
		url         *url.URL
		client      *fakeClient
		metricType  string
		metricName  string
		metricValue string
	}

	type test struct {
		name    string
		args    args
		want    want
		wantErr error
	}

	tests := []test{
		{
			name: "Should `Post` to necessary address",
			args: args{
				url:         &url.URL{Host: "127.0.0.1:8080", Scheme: "http"},
				client:      &fakeClient{},
				metricType:  "testType",
				metricName:  "testName",
				metricValue: "testValue",
			},
			want: want{
				url:         "http://127.0.0.1:8080/update/testType/testName/testValue",
				contentType: "text/plain",
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			emittingFunc, err := HTTPEmittingFunc(test.args.url, test.args.client)
			if err != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			emittingFunc(test.args.metricType, test.args.metricName, test.args.metricValue)
			assert.Equal(t, test.want.url, test.args.client.url)
			assert.Equal(t, test.want.contentType, test.args.client.contentType)
		})
	}

}

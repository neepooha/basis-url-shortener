package delete_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	resp2 "url_shortener/internal/lib/api/response"
	"url_shortener/internal/lib/logger/handlers/slogdiscard"
	"url_shortener/internal/storage"

	"github.com/go-chi/chi/v5"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"url_shortener/internal/transport/handlers/url/delete"
	"url_shortener/internal/transport/handlers/url/delete/mocks"
)

func TestDeleteHandler(t *testing.T) {
	cases := []struct {
		name      string
		uri       string
		respError string
		mockError error
		code      int
	}{
		{
			name: "Success",
			uri:  "/url/youtube",
			code: http.StatusOK,
		},
		{
			name:      "Omitted alias",
			uri:       "/url/",
			respError: "invalid request",
			code:      http.StatusNotFound,
		},
		{
			name:      "Alias Not Found",
			uri:       "/url/geek",
			respError: "url by alias was not found",
			mockError: storage.ErrAliasNotFound,
			code:      http.StatusOK,
		},
		{
			name:      "Delete Error",
			uri:       "/url/10",
			respError: "internal error",
			mockError: errors.New("unexpected error"),
			code:      http.StatusOK,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlDeleterMock := mocks.NewURLDeleter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlDeleterMock.On("DeleteURL", mock.AnythingOfType("string")).
					Return(tc.mockError).
					Once()
			}

			handler := chi.NewRouter()
			handler.Delete("/url/{alias}", delete.New(slogdiscard.NewDiscardLogger(), urlDeleterMock))

			req, err := http.NewRequest(http.MethodDelete, tc.uri, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, tc.code)

			if rr.Code == http.StatusOK {
				body := rr.Body.String()

				var resp resp2.Response

				require.NoError(t, json.Unmarshal([]byte(body), &resp))

				require.Equal(t, tc.respError, resp.Error)
			}
		})
	}
}

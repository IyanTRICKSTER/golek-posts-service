package responses

type HttpResponse struct {
	StatusCode int         `json:"status_code,omitempty"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data"`
}

type HttpPaginationResponse struct {
	PerPage int64 `json:"per_page"`
	Page    int64 `json:"page"`
	HttpResponse
}

type HttpErrorResponse struct {
	StatusCode int         `json:"status_code,omitempty"`
	Error      string      `json:"error,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

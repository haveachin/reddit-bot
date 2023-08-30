package reddit

import (
	"net/http"
)

func ConvertToDesktop(mobilelink string) (string, error) {

	response, err := http.Get(mobilelink)
	if err != nil {
		return "", err
	}
	finalURL := response.Request.URL.String()
	return finalURL, nil
}

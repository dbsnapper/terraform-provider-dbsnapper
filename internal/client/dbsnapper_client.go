package client

import (
	"github.com/joescharf/dbsnapper/v2/apiv1"
)

type DBSnapper struct {
	IsReady bool
	API     *apiv1.APIV1
}

func NewDBSnapper(authtoken, baseURL string) *DBSnapper {
	api := apiv1.NewClient(authtoken, baseURL)

	d := &DBSnapper{
		API: api,
	}

	d.IsReady = api.IsReady()

	return d
}

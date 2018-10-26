/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package component

import (
	"github.com/ortuman/jackal/component/httpapi"
	apiStorage "github.com/ortuman/jackal/component/httpapi/storage"
	"github.com/ortuman/jackal/component/httpupload"
	"github.com/ortuman/jackal/component/muc"
	mucStorage "github.com/ortuman/jackal/component/muc/storage"
)

type Config struct {
	HttpApi    *httpapi.Config    `yaml:"http_api"`
	HttpUpload *httpupload.Config `yaml:"http_upload"`
	Mucs       []muc.Config       `yaml:"muc_com"`
	Storage    *mucStorage.Config `yaml:"muc_storage"`
	ApiStorage *apiStorage.Config `yaml:"api_storage"`
}

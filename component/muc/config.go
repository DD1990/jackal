/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package muc

import (
	"errors"
)

const (
	defaultUploadPath = "/var/lib/jackal/httpupload"
	defaultSizeLimit  = 1048576
)

type Config struct {
	Host        string
}

type configProxy struct {
	Host        string `yaml:"host"`
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	p := configProxy{}
	if err := unmarshal(&p); err != nil {
		return err
	}
	// mandatory fields
	// 检查必填项
	if len(p.Host) == 0 {
		return errors.New("muc.Config: host value must be set")
	}
	c.Host = p.Host

	// optional fields
	// 可选字段

	return nil
}

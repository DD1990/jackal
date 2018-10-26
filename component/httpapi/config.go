/*
 * 组件配置.
 * See the LICENSE file for more information.
 */

package httpapi

import (
	"errors"
)

type Config struct {
	Host string
	Port string
}

type configProxy struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	p := configProxy{}
	if err := unmarshal(&p); err != nil {
		return err
	}
	// mandatory fields
	// 检查必填项
	if len(p.Host) == 0 {
		return errors.New("httpapi.Config: host value must be set")
	}
	if len(p.Port) == 0 {
		return errors.New("httpapi.Config: port value must be set")
	}
	c.Host = p.Host
	c.Port = p.Port

	// optional fields
	// 可选字段

	return nil
}

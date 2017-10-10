// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Authentication struct {
  Username string
  Password string
}

type Config struct {
	Period time.Duration `config:"period"`
  Labels []string `config:"labels"`
  OpenStatuses []string `config:"statuses.open"`
  Project string `config:"project"`
  Url string `config:"url"`
  Authentication Authentication `config:"authentication"`
}

var DefaultConfig = Config{
	Period: 10 * time.Second,
}

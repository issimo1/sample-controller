package server

import "encoding/json"

type Service struct {
	Name []string `json:"services"`
}

func (s *Service) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Service) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

package dao

import "gorm.io/gorm"

type Deployment struct {
	gorm.Model
	Name        string
	Namespace   string
	Description string
	Image       string
	Replicas    int
	Auto        int
}

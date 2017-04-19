package controllers

import "humpback-agent/config"

// FaqController - faq info
type FaqController struct {
	baseController
}

// Prepare - Override baseController
func (faq *FaqController) Prepare() {

}

// Get - Return config info
func (faq *FaqController) Get() {
	var conf = config.GetConfig()
	faq.JSON(conf)
}

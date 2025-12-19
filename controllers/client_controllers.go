package controllers

import (
	"wireguard_api/config"
	"wireguard_api/usecases"

	"github.com/gin-gonic/gin"
)

func NewController(service *usecases.Usecases, cfg *config.ServerConfig) *Controller {
	return &Controller{Service: service}
}

func (ctrl *Controller) GetStatus(c *gin.Context) {
	data, err := ctrl.Service.GetStatus()
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": data})
}

func (ctrl *Controller) GetAllClients(c *gin.Context) {
	data, err := ctrl.Service.GetAllClients()
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": data})
}

func (ctrl *Controller) AddClient(c *gin.Context) {
	var dataJson addClient
	err := c.BindJSON(&dataJson)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	data, err := ctrl.Service.NewClient(dataJson.Ifname, dataJson.Ip, dataJson.AllowedIp)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": data})
}

func (ctrl *Controller) AddInterface(c *gin.Context) {
	var dataJson addServer
	err := c.BindJSON(&dataJson)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	data, err := ctrl.Service.NewInterface(dataJson.Ifname, dataJson.Ip, dataJson.Endpoint, dataJson.Port)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": data})
}

func (ctrl *Controller) DeleteClient(c *gin.Context) {
	if !ctrl.cfg.ClientDelete {
		c.JSON(500, gin.H{"result": "Don't have permissions for delete client on this server"})
		return
	}
	var client deleteClient
	err := c.BindJSON(&client)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	err = ctrl.Service.DeleteClient(client.Public)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": "ok"})
}

func (ctrl *Controller) GetClientArchive(c *gin.Context) {
	data, err := ctrl.Service.GetClientArchive()
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": data})
}

func (ctrl *Controller) GetVersion(c *gin.Context) {
	c.JSON(200, gin.H{"result": config.Version})
}

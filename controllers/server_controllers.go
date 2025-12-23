package controllers

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) CtrlDeleteServer(c *gin.Context) {
	if !ctrl.cfg.DeleteInterface {
		c.JSON(500, gin.H{"result": "Don't have permissions for delete interface on this server"})
		return
	}
	var ser deleteServer
	err := c.BindJSON(&ser)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	err = ctrl.service.DeleteServer(ser.Private, ser.Ifname)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": "ok"})
}

func (ctrl *Controller) CtrlStopServer(c *gin.Context) {
	var ser ServerStartStop
	err := c.BindJSON(&ser)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	err = ctrl.service.StopInterface(ser.Ifname)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": "ok"})
}

func (ctrl *Controller) CtrlStartServer(c *gin.Context) {
	var ser ServerStartStop
	err := c.BindJSON(&ser)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	err = ctrl.service.StartInterface(ser.Ifname)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": "ok"})
}

func (ctrl *Controller) SetForward(c *gin.Context) {
	var ser ServerForward
	err := c.BindJSON(&ser)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	comment := strings.ReplaceAll(ser.Comment, " ", "_")
	err = ctrl.service.SetUsForward(ser.Position, ser.Action, ser.Command, ser.Source, ser.Destination, ser.Protocol, ser.Port, comment, ser.List, ser.Except)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": "ok"})
}

func (ctrl *Controller) SetForwardUpdateList(c *gin.Context) {
	var ser ServerForwardUpdateList
	err := c.BindJSON(&ser)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	err = ctrl.service.UpdateIpSetList(ser.Command, ser.IpsetName, ser.IpList, ser.Single)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": "ok"})
}

func (ctrl *Controller) SetMasquerade(c *gin.Context) {
	var ser ServerMasquerade
	err := c.BindJSON(&ser)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	comment := strings.ReplaceAll(ser.Comment, " ", "_")
	err = ctrl.service.SetUsMasquerade(ser.Command, ser.Source, ser.Ifname, comment)
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": "ok"})
}

func (ctrl *Controller) CtrlGetServerArchive(c *gin.Context) {
	data, err := ctrl.service.GetServerArchive()
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": data})
}

func (ctrl *Controller) CtrlGetInterfaces(c *gin.Context) {
	data, err := ctrl.service.GetServerInterfaces()
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": data})
}

func (ctrl *Controller) CtrlGetIptables(c *gin.Context) {
	data, err := ctrl.service.GetIptablesRules()
	if err != nil {
		c.JSON(500, gin.H{"result": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": data})
}

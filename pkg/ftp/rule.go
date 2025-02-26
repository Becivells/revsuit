package ftp

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/li4n0/revsuit/internal/database"
	"github.com/li4n0/revsuit/internal/rule"
	"gorm.io/gorm/clause"
	log "unknwon.dev/clog/v2"
)

// Rule FTP rule struct
type Rule struct {
	rule.BaseRule `yaml:",inline"`
	PasvAddress   string `gorm:"pasv_address" json:"pasv_address" form:"pasv_address" yaml:"pasv_address"`
	Data          []byte `json:"data" form:"data"`
}

func (Rule) TableName() string {
	return "ftp_rules"
}

// NewRule creates a new ftp rule struct
func NewRule(name, flagFormat, pasvAddress string, pushToClient, notice bool) *Rule {
	return &Rule{
		BaseRule: rule.BaseRule{
			Name:         name,
			FlagFormat:   flagFormat,
			PushToClient: pushToClient,
			Notice:       notice,
		},
		PasvAddress: pasvAddress,
	}
}

// CreateOrUpdate creates or updates the ftp rule in database and ruleSet
func (r *Rule) CreateOrUpdate() (err error) {
	db := database.DB.Model(r)
	err = db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns(
			[]string{
				"name",
				"flag_format",
				"base_rank",
				"pasv_address",
				"data",
				"push_to_client",
				"notice",
			}),
	}).Create(r).Error
	if err != nil {
		return
	}

	return GetServer().UpdateRules()
}

// Delete deletes the ftp rule in database and ruleSet
func (r *Rule) Delete() (err error) {
	db := database.DB.Model(r)
	err = db.Delete(r).Error
	if err != nil {
		return
	}

	return GetServer().UpdateRules()
}

// ListRules lists all ftp rules those satisfy the filter
func ListRules(c *gin.Context) {
	var (
		ftpRule  Rule
		res      []Rule
		count    int64
		order    = c.Query("order")
		pageSize = 10
	)

	if c.Query("pageSize") != "" {
		if n, err := strconv.Atoi(c.Query("pageSize")); err == nil {
			if n > 0 && n < 100 {
				pageSize = n
			}
		}
	}

	if err := c.ShouldBind(&ftpRule); err != nil {
		c.JSON(400, gin.H{
			"status": "failed",
			"error":  err.Error(),
			"result": nil,
		})
		return
	}

	db := database.DB.Model(&ftpRule)
	if ftpRule.Name != "" {
		db.Where("name = ?", ftpRule.Name)
	}
	db.Count(&count)

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		c.JSON(400, gin.H{
			"status": "failed",
			"error":  err.Error(),
			"result": nil,
		})
		return
	}

	if order != "asc" {
		order = "desc"
	}

	if err := db.Order("base_rank desc").Order("id" + " " + order).Count(&count).Offset((page - 1) * pageSize).Limit(pageSize).Find(&res).Error; err != nil {
		c.JSON(400, gin.H{
			"status": "failed",
			"error":  err.Error(),
			"data":   nil,
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "succeed",
		"error":  nil,
		"result": gin.H{"count": count, "data": res},
	})
}

// UpsertRules creates or updates ftp rule from user submit
func UpsertRules(c *gin.Context) {
	var (
		ftpRule Rule
		update  bool
	)

	if err := c.ShouldBind(&ftpRule); err != nil {
		c.JSON(400, gin.H{
			"status": "failed",
			"error":  err.Error(),
			"data":   nil,
		})
		return
	}

	if ftpRule.ID != 0 {
		update = true
	}

	if err := ftpRule.CreateOrUpdate(); err != nil {
		c.JSON(400, gin.H{
			"status": "failed",
			"error":  err.Error(),
			"result": nil,
		})
		return
	}

	if update {
		log.Trace("FTP rule[id:%d] has been updated", ftpRule.ID)
	} else {
		log.Trace("FTP rule[id:%d] has been created", ftpRule.ID)
	}

	c.JSON(200, gin.H{
		"status": "succeed",
		"error":  nil,
		"result": nil,
	})
}

// DeleteRules deletes ftp rule from user submit
func DeleteRules(c *gin.Context) {
	var ftpRule Rule

	if err := c.ShouldBind(&ftpRule); err != nil {
		c.JSON(400, gin.H{
			"status": "failed",
			"error":  err.Error(),
			"data":   nil,
		})
		return
	}

	if err := ftpRule.Delete(); err != nil {
		c.JSON(400, gin.H{
			"status": "failed",
			"error":  err.Error(),
			"data":   nil,
		})
		return
	}

	log.Trace("FTP rule[id:%d] has been deleted", ftpRule.ID)

	c.JSON(200, gin.H{
		"status": "succeed",
		"error":  nil,
		"data":   nil,
	})
}

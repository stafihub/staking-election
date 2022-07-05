package dao_election

import "github.com/stafihub/staking-election/db"

type AnnualRate struct {
	db.BaseModel
	RTokenDenom string `gorm:"type:varchar(10) not null;default:'';column:rtoken_denom;uniqueIndex"`
	AnnualRate  string `gorm:"type:varchar(80) not null;default:'';column:annual_rate"`
}

func (f AnnualRate) TableName() string {
	return "staking_election_annual_rate"
}

func UpOrInAnnualRate(db *db.WrapDb, c *AnnualRate) error {
	return db.Save(c).Error
}

func GetAnnualRate(db *db.WrapDb, denom string) (info *AnnualRate, err error) {
	info = &AnnualRate{}
	err = db.Take(info, "denom = ?", denom).Error
	return
}
func GetAnnualRateList(db *db.WrapDb) (infos []*AnnualRate, err error) {
	err = db.Find(&infos).Error
	return
}

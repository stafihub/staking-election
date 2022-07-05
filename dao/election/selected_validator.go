package dao_election

import "github.com/stafihub/staking-election/db"

type SelectedValidator struct {
	db.BaseModel
	RTokenDenom      string `gorm:"type:varchar(10) not null;default:'';column:rtoken_denom;uniqueIndex:uni_denom_pool_val"`
	PoolAddress      string `gorm:"type:varchar(80) not null;default:'';column:pool_address;uniqueIndex:uni_denom_pool_val"`
	ValidatorAddress string `gorm:"type:varchar(80) not null;default:'';column:validator_address;uniqueIndex:uni_denom_pool_val"`
	Moniker          string `gorm:"type:varchar(50) not null;default:'';column:moniker"`
}

func (f SelectedValidator) TableName() string {
	return "staking_election_selected_validator"
}

func UpOrInSelectedValidator(db *db.WrapDb, c *SelectedValidator) error {
	return db.Save(c).Error
}

func GetSelectedValidator(db *db.WrapDb, denom, poolAddress, validatorAddress string) (info *SelectedValidator, err error) {
	info = &SelectedValidator{}
	err = db.Take(info, "rtoken_denom = ? and pool_address = ? and validator_address = ?", denom, poolAddress, validatorAddress).Error
	return
}

func GetAllSelectedValidators(db *db.WrapDb) (infos []*SelectedValidator, err error) {
	err = db.Find(&infos).Error
	return
}

func DeleteSelectedValidator(db *db.WrapDb, denom, poolAddress, validatorAddress string) error {
	return db.Delete(&SelectedValidator{}, "rtoken_denom = ? and pool_address = ? and validator_address = ?", denom, poolAddress, validatorAddress).Error
}

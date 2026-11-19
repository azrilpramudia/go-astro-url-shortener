package entity

type LinkTag struct {
	ID     uint   `gorm:"column:id;primaryKey;autoIncrement"`
	LinkID uint   `gorm:"column:link_id;not null;uniqueIndex:link_tags_link_id_name_key"`
	Name   string `gorm:"column:name;size:32;not null;uniqueIndex:link_tags_link_id_name_key"`
}

func (t *LinkTag) TableName() string {
	return "link_tags"
}

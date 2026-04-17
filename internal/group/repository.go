package group

import "gorm.io/gorm"

type Repository interface {
	// 群操作
	Create(group *Group) error
	FindByID(id uint) (*Group, error)
	FindByCreatorID(creatorID uint) ([]Group, error)
	Update(group *Group) error
	Delete(id uint) error

	// 成员操作
	AddMember(member *GroupMember) error
	RemoveMember(groupID, userID uint) error
	FindMember(groupID, userID uint) (*GroupMember, error)
	FindMembersByGroupID(groupID uint) ([]GroupMember, error)
	FindGroupsByUserID(userID uint) ([]Group, error)
	UpdateMemberRole(groupID, userID uint, role int) error
	CountMembers(groupID uint) (int64, error)

	// 消息操作
	SaveMessage(msg *GroupMessage) error
	FindMessagesByGroupID(groupID uint, limit, offset int) ([]GroupMessage, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

//创建群聊
func (r *repository) Create(group Group) error {
	return r.db.Create(group).Error
}

//群ID查找
func (r *repository) FindByID(id uint) (*Group, error) {
	var group Group
	err := r.db.First(&group, id).Error
	return &group, err
}

//按创建者ID查找
func (r *repository) FindByCreatorID(creatorID uint) ([]Group, error) {
	var groups []Group
	err := r.db.Where("creator_id = ?", creatorID).Find(&groups).Error
	return groups, err
}

//群更新
func (r *repository) Update(group *Group) error {
	return r.db.Save(group).Error
}

//群删除
func (r *repository) Delete(id uint) error {
	return r.db.Delete(&Group{}, id).Error
}

//加入成员
func (r *repository) AddMember(member *GroupMember) error {
	return r.db.Create(member).Error
}

//移除成员
func (r *repository) RemoveMember(groupID, userID uint) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&GroupMember{}).Error
}

//ID查成员
func (r *repository) FindMember(groupID, userID uint) (*GroupMember, error) {
	var member GroupMember
	err := r.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error
	return &member, err
}

//查一个群的所有成员
func (r *repository) FindMembersByGroupID(groupID uint) ([]GroupMember, error) {
	var members []GroupMember
	err := r.db.Where("group_id = ?", groupID).Find(&members).Error
	return members, err
}

//查一个用户所在的群聊
func (r *repository) FindGroupsByUserID(userID uint) ([]Group, error) {
	var groups []Group
	err := r.db.Table("groups").
		Joins("JOIN group_members ON groups.id = group_members.group_id").
		Where("group_members.user_id = ?", userID).
		Where("groups.deleted_at IS NULL").
		Find(&groups).Error
	return groups, err
}

//更新用户群角色
func (r *repository) UpdateMemberRole(groupID, userID uint, role int) error {
	return r.db.Model(&GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("role", role).Error
}

//群成员数量
func (r *repository) CountMembers(groupID uint) (int64, error) {
	var count int64
	err := r.db.Model(&GroupMember{}).Where("group_id = ?", groupID).Count(&count).Error
	return count, err
}

//发消息
func (r *repository) SaveMessage(msg *GroupMessage) error {
	return r.db.Create(msg).Error
}

//查群消息
func (r *repository) FindMessagesByGroupID(groupID uint, limit, offset int) ([]GroupMessage, error) {
	var messages []GroupMessage
	err := r.db.Where("group_id = ?", groupID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&messages).Error
	return messages, err
}

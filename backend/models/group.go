package models

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Group MongoDB中的群组模型
type Group struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Avatar      string             `bson:"avatar" json:"avatar"`
	CreatorID   string             `bson:"creatorId" json:"creatorId"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
	Deleted     bool               `bson:"deleted" json:"-"`
}

// GroupMember MongoDB中的群组成员模型
type GroupMember struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID   string             `bson:"groupId" json:"groupId"`
	UserID    string             `bson:"userId" json:"userId"`
	Role      string             `bson:"role" json:"role"` // admin, member
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
	Deleted   bool               `bson:"deleted" json:"-"`
}

// CreateGroup 创建新群组
func CreateGroup(name, description, avatar, creatorID string) (*Group, error) {
	// 检查创建者是否存在
	_, err := GetUserByID(creatorID)
	if err != nil {
		return nil, errors.New("创建者不存在")
	}

	// 创建群组
	now := time.Now()
	defaultAvatar := "https://cn.bing.com/images/search?view=detailV2&ccid=rTMre%2bV%2b&id=63662EF8E2C07443A67E94BD782DCCB6DB48B0C6&thid=OIP.rTMre-V-rMq3T1EP9W7vlAHaEK&mediaurl=https%3a%2f%2fimg-baofun.zhhainiao.com%2fpcwallpaper_ugc%2fstatic%2f388379538bf2af745f3f7cfea82816a2.jpg%3fx-oss-process%3dimage%252fresize%252cm_lfit%252cw_3840%252ch_2160&exph=2160&expw=3840&q=%e9%a3%8e%e6%99%af%e5%9b%be&simid=607992754002226257&FORM=IRPRST&ck=0B63A81C8EF37FA62162CBB348088488&selectedIndex=14&itb=0"
	if avatar == "" {
		avatar = defaultAvatar
	}
	group := &Group{
		Name:        name,
		Description: description,
		Avatar:      avatar,
		CreatorID:   creatorID,
		CreatedAt:   now,
		UpdatedAt:   now,
		Deleted:     false,
	}

	collection := MongoDatabase.Collection("groups")
	result, err := collection.InsertOne(context.Background(), group)
	if err != nil {
		return nil, err
	}

	group.ID = result.InsertedID.(primitive.ObjectID)

	// 添加创建者为管理员
	_, err = AddGroupMember(group.ID.Hex(), creatorID, "admin")
	if err != nil {
		return nil, err
	}

	return group, nil
}

// GetGroupByID 通过ID获取群组
func GetGroupByID(id string) (*Group, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	collection := MongoDatabase.Collection("groups")
	var group Group
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID, "deleted": false}).Decode(&group)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

// UpdateGroup 更新群组信息
func UpdateGroup(group *Group) error {
	group.UpdatedAt = time.Now()

	collection := MongoDatabase.Collection("groups")
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": group.ID},
		bson.M{"$set": group},
	)

	return err
}

// DeleteGroup 删除群组
func DeleteGroup(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	collection := MongoDatabase.Collection("groups")
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"deleted": true, "updatedAt": time.Now()}},
	)

	return err
}

// GetUserGroups 获取用户所在的群组
func GetUserGroups(userID string) ([]*Group, error) {
	// 先获取用户所在的群组成员记录
	memberCollection := MongoDatabase.Collection("group_members")
	cursor, err := memberCollection.Find(context.Background(), bson.M{"userId": userID, "deleted": false})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var members []*GroupMember
	if err = cursor.All(context.Background(), &members); err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return []*Group{}, nil
	}

	// 获取群组ID列表
	var groupIDs []primitive.ObjectID
	for _, member := range members {
		groupID, err := primitive.ObjectIDFromHex(member.GroupID)
		if err != nil {
			continue
		}
		groupIDs = append(groupIDs, groupID)
	}

	// 获取群组信息
	groupCollection := MongoDatabase.Collection("groups")
	cursor, err = groupCollection.Find(context.Background(), bson.M{"_id": bson.M{"$in": groupIDs}, "deleted": false})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var groups []*Group
	if err = cursor.All(context.Background(), &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

// AddGroupMember 添加群组成员
func AddGroupMember(groupID, userID, role string) (*GroupMember, error) {
	// 检查群组是否存在
	_, err := GetGroupByID(groupID)
	if err != nil {
		return nil, errors.New("群组不存在")
	}

	// 检查用户是否存在
	_, err = GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户是否已经是群组成员
	collection := MongoDatabase.Collection("group_members")
	var existingMember GroupMember
	err = collection.FindOne(context.Background(), bson.M{
		"groupId": groupID,
		"userId":  userID,
		"deleted": false,
	}).Decode(&existingMember)

	if err == nil {
		return nil, errors.New("用户已经是群组成员")
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// 添加群组成员
	now := time.Now()
	member := &GroupMember{
		GroupID:   groupID,
		UserID:    userID,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
		Deleted:   false,
	}

	result, err := collection.InsertOne(context.Background(), member)
	if err != nil {
		return nil, err
	}

	member.ID = result.InsertedID.(primitive.ObjectID)
	return member, nil
}

// GetGroupMembers 获取群组成员
func GetGroupMembers(groupID string) ([]*GroupMember, error) {
	collection := MongoDatabase.Collection("group_members")
	cursor, err := collection.Find(context.Background(), bson.M{"groupId": groupID, "deleted": false})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var members []*GroupMember
	if err = cursor.All(context.Background(), &members); err != nil {
		return nil, err
	}

	return members, nil
}

// RemoveGroupMember 移除群组成员
func RemoveGroupMember(groupID, userID string) error {
	collection := MongoDatabase.Collection("group_members")
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"groupId": groupID, "userId": userID, "deleted": false},
		bson.M{"$set": bson.M{"deleted": true, "updatedAt": time.Now()}},
	)

	return err
}

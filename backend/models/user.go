package models

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// User MongoDB中的用户模型
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username"`
	Password  string             `bson:"password" json:"-"`
	Email     string             `bson:"email" json:"email"`
	Avatar    string             `bson:"avatar" json:"avatar"`
	Status    string             `bson:"status" json:"status"` // online, offline, away
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
	Deleted   bool               `bson:"deleted" json:"-"`
}

// Friendship MongoDB中的好友关系模型
type Friendship struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"userId" json:"userId"`
	FriendID  string             `bson:"friendId" json:"friendId"`
	Status    string             `bson:"status" json:"status"` // pending, accepted, rejected
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
	Deleted   bool               `bson:"deleted" json:"-"`
}

// CheckPassword 检查密码是否正确
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// CreateUser 创建新用户
func CreateUser(username, password, email string) (*User, error) {
	// 检查用户名是否已存在
	collection := MongoDatabase.Collection("users")
	var existingUser User
	err := collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&existingUser)
	if err == nil {
		return nil, errors.New("用户名已存在")
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	// 检查邮箱是否已存在
	if email != "" {
		err = collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&existingUser)
		if err == nil {
			return nil, errors.New("邮箱已存在")
		} else if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	now := time.Now()
	defaultAvatar := "https://cn.bing.com/images/search?view=detailV2&ccid=StrDRqen&id=92D4568BD21B0D03661242697B510D4D295C4727&thid=OIP.StrDRqennoZNbzSPZapKZwAAAA&mediaurl=https%3a%2f%2fimg.shetu66.com%2f2023%2f06%2f26%2f1687770031227597.png&exph=265&expw=474&q=%e9%a3%8e%e6%99%af%e5%9b%be&simid=608050023063974373&FORM=IRPRST&ck=69A4339CE32B1BCBD039D689CE35ACC3&selectedIndex=9&itb=0"
	user := &User{
		Username:  username,
		Password:  string(hashedPassword),
		Email:     email,
		Avatar:    defaultAvatar,
		Status:    "offline",
		CreatedAt: now,
		UpdatedAt: now,
		Deleted:   false,
	}

	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

// GetUserByID 通过ID获取用户
func GetUserByID(id string) (*User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	collection := MongoDatabase.Collection("users")
	var user User
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID, "deleted": false}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByUsername 通过用户名获取用户
func GetUserByUsername(username string) (*User, error) {
	collection := MongoDatabase.Collection("users")
	var user User
	err := collection.FindOne(context.Background(), bson.M{"username": username, "deleted": false}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser 更新用户信息
func UpdateUser(user *User) error {
	user.UpdatedAt = time.Now()

	collection := MongoDatabase.Collection("users")
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": user.ID},
		bson.M{"$set": user},
	)

	return err
}

// AddFriend 添加好友请求
func AddFriend(userID, friendID string) (*Friendship, error) {
	// 检查用户和好友是否存在
	_, err := GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	_, err = GetUserByID(friendID)
	if err != nil {
		return nil, errors.New("好友不存在")
	}

	// 检查是否已经是好友
	collection := MongoDatabase.Collection("friendships")
	var existingFriendship Friendship
	err = collection.FindOne(context.Background(), bson.M{
		"$or": []bson.M{
			{"userId": userID, "friendId": friendID},
			{"userId": friendID, "friendId": userID},
		},
		"deleted": false,
	}).Decode(&existingFriendship)

	if err == nil {
		return nil, errors.New("好友关系已存在")
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// 创建好友关系
	now := time.Now()
	friendship := &Friendship{
		UserID:    userID,
		FriendID:  friendID,
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
		Deleted:   false,
	}

	result, err := collection.InsertOne(context.Background(), friendship)
	if err != nil {
		return nil, err
	}

	friendship.ID = result.InsertedID.(primitive.ObjectID)
	return friendship, nil
}

// GetFriendships 获取用户的好友关系
func GetFriendships(userID string, status string) ([]*Friendship, error) {
	collection := MongoDatabase.Collection("friendships")

	filter := bson.M{
		"$or": []bson.M{
			{"userId": userID},
			{"friendId": userID},
		},
		"deleted": false,
	}

	if status != "" {
		filter["status"] = status
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var friendships []*Friendship
	if err = cursor.All(context.Background(), &friendships); err != nil {
		return nil, err
	}

	return friendships, nil
}

// UpdateFriendship 更新好友关系
func UpdateFriendship(friendship *Friendship) error {
	friendship.UpdatedAt = time.Now()

	collection := MongoDatabase.Collection("friendships")
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": friendship.ID},
		bson.M{"$set": friendship},
	)

	return err
}

// DeleteFriendship 删除好友关系
func DeleteFriendship(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	collection := MongoDatabase.Collection("friendships")
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"deleted": true, "updatedAt": time.Now()}},
	)

	return err
}

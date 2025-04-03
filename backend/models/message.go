package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 消息类型常量
const (
	MessageTypePrivate = "private" // 私聊消息
	MessageTypeGroup   = "group"   // 群聊消息
)

// Message MongoDB中的消息模型
type Message struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type       string             `bson:"type" json:"type"` // private, group
	SenderID   string             `bson:"senderId" json:"senderId"`
	ReceiverID string             `bson:"receiverId,omitempty" json:"receiverId,omitempty"` // 私聊时的接收者ID
	GroupID    string             `bson:"groupId,omitempty" json:"groupId,omitempty"`       // 群聊时的群组ID
	Content    string             `bson:"content" json:"content"`
	Timestamp  time.Time          `bson:"timestamp" json:"timestamp"`
	Read       bool               `bson:"read" json:"read"` // 消息是否已读
}

// SavePrivateMessage 保存私聊消息到MongoDB
func SavePrivateMessage(senderID, receiverID, content string) (*Message, error) {
	message := &Message{
		Type:       MessageTypePrivate,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		Timestamp:  time.Now(),
		Read:       false,
	}

	collection := MongoDatabase.Collection("messages")
	result, err := collection.InsertOne(context.Background(), message)
	if err != nil {
		return nil, err
	}

	message.ID = result.InsertedID.(primitive.ObjectID)
	return message, nil
}

// SaveGroupMessage 保存群聊消息到MongoDB
func SaveGroupMessage(senderID, groupID, content string) (*Message, error) {
	message := &Message{
		Type:      MessageTypeGroup,
		SenderID:  senderID,
		GroupID:   groupID,
		Content:   content,
		Timestamp: time.Now(),
		Read:      false,
	}

	collection := MongoDatabase.Collection("messages")
	result, err := collection.InsertOne(context.Background(), message)
	if err != nil {
		return nil, err
	}

	message.ID = result.InsertedID.(primitive.ObjectID)
	return message, nil
}

// GetPrivateMessages 获取两个用户之间的私聊消息
func GetPrivateMessages(userID1, userID2 string, limit, skip int64) ([]*Message, error) {
	collection := MongoDatabase.Collection("messages")

	// 构建查询条件：(sender=userID1 AND receiver=userID2) OR (sender=userID2 AND receiver=userID1)
	filter := bson.M{
		"type": MessageTypePrivate,
		"$or": []bson.M{
			{
				"senderId":   userID1,
				"receiverId": userID2,
			},
			{
				"senderId":   userID2,
				"receiverId": userID1,
			},
		},
	}

	// 设置排序和分页
	opts := options.Find().
		SetSort(bson.D{{"timestamp", -1}}). // 按时间降序
		SetLimit(limit).
		SetSkip(skip)

	cursor, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var messages []*Message
	if err = cursor.All(context.Background(), &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// GetGroupMessages 获取群组消息
func GetGroupMessages(groupID string, limit, skip int64) ([]*Message, error) {
	collection := MongoDatabase.Collection("messages")

	// 构建查询条件
	filter := bson.M{
		"type":    MessageTypeGroup,
		"groupId": groupID,
	}

	// 设置排序和分页
	opts := options.Find().
		SetSort(bson.D{{"timestamp", 1}}). // 按时间降序
		SetLimit(limit).
		SetSkip(skip)

	cursor, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var messages []*Message
	if err = cursor.All(context.Background(), &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// MarkMessagesAsRead 将消息标记为已读
func MarkMessagesAsRead(messageIDs []primitive.ObjectID) error {
	collection := MongoDatabase.Collection("messages")

	filter := bson.M{
		"_id": bson.M{"$in": messageIDs},
	}

	update := bson.M{
		"$set": bson.M{"read": true},
	}

	_, err := collection.UpdateMany(context.Background(), filter, update)
	return err
}

package handler

import (
	"context"
	"mx-shop-srvs/user_interaction_srv/global"
	"mx-shop-srvs/user_interaction_srv/model"
	"mx-shop-srvs/user_interaction_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MessageServer struct {
	proto.UnimplementedMessageServer
}

func (s *MessageServer) MessageList(ctx context.Context, req *proto.MessageRequest) (*proto.MessageListResponse, error) {
	var msgList []model.LeavingMessages

	if result := global.DB.Where("user = ?", req.UserId).Find(&msgList); result.Error != nil {
		return nil, status.Error(codes.Internal, "获取留言消息列表时出错")
	}

	var resp proto.MessageListResponse
	resp.Total = int32(len(msgList))
	for _, msg := range msgList {
		msgResp := proto.MessageResponse{
			Id:          msg.ID,
			UserId:      msg.User,
			MessageType: msg.MessageType,
			Subject:     msg.Subject,
			Message:     msg.Message,
			File:        msg.File,
		}
		resp.Data = append(resp.Data, &msgResp)
	}

	return &resp, nil
}

func (s *MessageServer) CreateMessage(ctx context.Context, req *proto.MessageRequest) (*proto.MessageResponse, error) {
	message := model.LeavingMessages{
		User:        req.UserId,
		MessageType: req.MessageType,
		Subject:     req.Subject,
		Message:     req.Message,
		File:        req.File,
	}

	if result := global.DB.Create(&message); result.RowsAffected == 0 {
		return nil, status.Error(codes.Internal, "创建留言消息时出错")
	}

	return &proto.MessageResponse{
		Id: message.ID,
	}, nil
}

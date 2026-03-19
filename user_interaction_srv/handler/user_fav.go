package handler

import (
	"context"
	"mx-shop-srvs/user_interaction_srv/global"
	"mx-shop-srvs/user_interaction_srv/model"
	"mx-shop-srvs/user_interaction_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserFavServer struct {
	proto.UnimplementedUserFavServer
}

func (s *UserFavServer) GetFavList(ctx context.Context, req *proto.UserFavRequest) (*proto.UserFavListResponse, error) {
	var userFavList []model.UserFav
	if result := global.DB.Model(&model.UserFav{}).Where("user = ?", req.UserId).Find(&userFavList); result.Error != nil {
		return nil, result.Error
	}

	var resp proto.UserFavListResponse
	resp.Total = int32(len(userFavList))
	for _, userFav := range userFavList {
		resp.Data = append(resp.Data, &proto.UserFavResponse{
			UserId:  userFav.User,
			GoodsId: userFav.Goods,
		})
	}

	return &resp, nil
}

func (s *UserFavServer) AddUserFav(ctx context.Context, req *proto.UserFavRequest) (*emptypb.Empty, error) {
	if result := global.DB.Model(&model.UserFav{}).Where("user = ? and goods = ?", req.UserId, req.GoodsId); result.RowsAffected == 1 {
		return &emptypb.Empty{}, status.Error(codes.InvalidArgument, "此收藏记录已经存在")
	}
	userFav := model.UserFav{
		User:  req.UserId,
		Goods: req.GoodsId,
	}

	if result := global.DB.Create(&userFav); result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Error(codes.Internal, "创换用户收藏失败")
	}

	return &emptypb.Empty{}, nil
}

func (s *UserFavServer) DeleteUserFav(ctx context.Context, req *proto.UserFavRequest) (*emptypb.Empty, error) {
	var userFav model.UserFav
	result := global.DB.Unscoped().Model(&model.UserFav{}).Where("user = ? and goods = ?", req.UserId, req.GoodsId).Find(&userFav)
	if result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Error(codes.InvalidArgument, "用户收藏不存在")
	}

	if result := global.DB.Unscoped().Where("user = ? and goods = ?", req.UserId, req.GoodsId).Delete(&model.UserFav{}); result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Error(codes.InvalidArgument, "删除用户收藏失败")
	}

	return &emptypb.Empty{}, nil
}

func (s *UserFavServer) GetUserFavDetail(ctx context.Context, req *proto.UserFavRequest) (*emptypb.Empty, error) {
	var userFav model.UserFav
	if result := global.DB.Model(&model.UserFav{}).Where("user = ? and goods = ?", req.UserId, req.GoodsId).Find(&userFav); result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Error(codes.InvalidArgument, "用户收藏不存在")
	}

	return &emptypb.Empty{}, nil
}

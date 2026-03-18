package handler

import (
	"context"
	"mx-shop-srvs/user_interaction_srv/global"
	"mx-shop-srvs/user_interaction_srv/model"
	"mx-shop-srvs/user_interaction_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type AddressServer struct {
	proto.UnimplementedAddressServer
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func (s *AddressServer) GetAddressList(ctx context.Context, req *proto.AddressRequest) (*proto.AddressListResponse, error) {
	var addresses []model.Address

	if result := global.DB.Where("user = ?", req.UserId).Find(&addresses); result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询地址信息时出错了")
	}

	var resp proto.AddressListResponse
	resp.Total = int32(len(addresses))

	for _, address := range addresses {
		resp.Data = append(resp.Data, &proto.AddressResponse{
			Id:           address.ID,
			UserId:       address.User,
			Province:     address.Province,
			City:         address.City,
			District:     address.District,
			Address:      address.Address,
			SignerName:   address.SignerName,
			SignerMobile: address.SignerMobile,
		})
	}

	return &resp, nil
}

func (s *AddressServer) CreateAddress(ctx context.Context, req *proto.AddressRequest) (*proto.AddressResponse, error) {
	address := model.Address{
		User:         req.UserId,
		Province:     req.Province,
		City:         req.City,
		District:     req.District,
		Address:      req.Address,
		SignerName:   req.SignerName,
		SignerMobile: req.SignerMobile,
		BaseModel:    model.BaseModel{},
	}

	result := global.DB.Create(&address)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.Internal, "新建地址时出错了")
	}

	return &proto.AddressResponse{
		Id: address.ID,
	}, nil
}

func (s *AddressServer) DeleteAddress(ctx context.Context, req *proto.AddressRequest) (*emptypb.Empty, error) {
	if result := global.DB.Where("id = ? and user = ?", req.Id, req.UserId).Find(&model.Address{}); result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Errorf(codes.NotFound, "找不到改地址信息")
	}

	if result := global.DB.Where("id = ? and user = ?", req.Id, req.UserId).Delete(&model.Address{}); result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Errorf(codes.Internal, "删除失败")
	}

	return &emptypb.Empty{}, nil
}

func (s *AddressServer) UpdateAddress(ctx context.Context, req *proto.AddressRequest) (*emptypb.Empty, error) {
	var address model.Address
	result := global.DB.Where("id = ? and user = ?", req.Id, req.UserId).Find(&address)
	if result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Errorf(codes.NotFound, "找不到改地址信息")
	}

	address.Province = req.Province
	address.City = req.City
	address.District = req.District
	address.Address = req.Address
	address.SignerMobile = req.SignerMobile
	address.SignerName = req.SignerName

	if result := global.DB.Save(&address); result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Errorf(codes.NotFound, "更新地址信息失败")
	}

	return &emptypb.Empty{}, nil
}

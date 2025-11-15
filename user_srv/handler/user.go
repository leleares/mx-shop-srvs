package handler

import (
	"context"

	"mx-shop-srvs/user_srv/global"
	"mx-shop-srvs/user_srv/model"
	"mx-shop-srvs/user_srv/proto"

	"gorm.io/gorm"
)

// 在此文件中实现生成的user.pb.go中的interface即可。
// 找到 type UserServer interface 将其中的方法实现即可。
type UserServer struct{}

func modelToResponse(user model.User) proto.UserInfoResponse {
	// 在grpc的message字段中如果有默认值的话，不能赋值nil进去，容易出错
	// 因此对于有可能为nil值的字段，要单独进行处理
	userInfoRsp := proto.UserInfoResponse{
		Id:       user.ID,
		Password: user.Password,
		NickName: user.NickName,
		Mobile:   user.Mobile,
		Gender:   user.Gender,
		Role:     int32(user.Role),
	}

	if user.Birthday != nil {
		userInfoRsp.BirthDay = uint64(user.Birthday.Unix())
	}

	return userInfoRsp
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

// 获取用户列表
func (s *UserServer) GetUserList(ctx context.Context, req *proto.PageInfo) (*proto.UserListResponse, error) {
	// 获取所有用户
	var users []model.User
	result := global.DB.Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}

	rsp := &proto.UserListResponse{}
	rsp.Total = uint32(result.RowsAffected)

	// 利用grom能力进行分页，重整users数据
	global.DB.Scopes(Paginate(int(req.Pn), int(req.PSize))).Find(&users)

	var userInfoList []*proto.UserInfoResponse // 每个元素都是指针
	for _, user := range users {
		userInfoRsp := modelToResponse(user)
		userInfoList = append(userInfoList, &userInfoRsp) // 这里也要传指针进去
	}

	rsp.Data = userInfoList

	return rsp, nil
}

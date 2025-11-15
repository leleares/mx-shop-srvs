package handler

import (
	"context"
	"crypto/sha512"
	"fmt"
	"strings"
	"time"

	"mx-shop-srvs/user_srv/global"
	"mx-shop-srvs/user_srv/model"
	"mx-shop-srvs/user_srv/proto"

	"github.com/anaskhan96/go-password-encoder"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *UserServer) GetUserByMobile(ctx context.Context, req *proto.MobileRequest) (*proto.UserInfoResponse, error) {
	var user model.User
	result := global.DB.Where("mobile = ?", req.Mobile).First(&user)

	if result.RowsAffected < 1 {
		return nil, status.Errorf(codes.NotFound, "用户不存在 ")
	}

	if result.Error != nil {
		return nil, result.Error
	}

	userInfo := modelToResponse(user)

	return &userInfo, nil
}

func (s *UserServer) GetUserById(ctx context.Context, req *proto.IdRequest) (*proto.UserInfoResponse, error) {
	var user model.User
	result := global.DB.Where("id = ?", req.Id).First(&user)

	if result.RowsAffected < 1 {
		return nil, status.Errorf(codes.NotFound, "用户不存在 ")
	}

	if result.Error != nil {
		return nil, result.Error
	}

	userInfo := modelToResponse(user)

	return &userInfo, nil
}

func (s *UserServer) CreateUser(ctx context.Context, req *proto.CreateUserInfo) (*proto.UserInfoResponse, error) {
	// 首先需要查询用户是否存在
	var user model.User
	result := global.DB.Where("mobile = ?", req.Mobile).First(&user)
	if result.RowsAffected >= 1 {
		return nil, status.Errorf(codes.AlreadyExists, "用户已经存在")
	}

	user.NickName = req.NickName
	user.Mobile = req.Mobile
	// 密码要加密
	options := &password.Options{16, 100, 32, sha512.New}
	salt, encodedPwd := password.Encode(req.Password, options)
	user.Password = fmt.Sprintf("$pbkdf2$%s$%s", salt, encodedPwd)

	result = global.DB.Create(&user)

	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, result.Error.Error())
	}

	if result.RowsAffected < 1 {
		return nil, status.Errorf(codes.Internal, "创建用户失败")
	}

	userInfo := modelToResponse(user)
	return &userInfo, nil
}

func (s *UserServer) UpdateUser(ctx context.Context, req *proto.UpdateUserInfo) (*empty.Empty, error) {
	var user model.User
	result := global.DB.First(&user, req.Id)
	if result.RowsAffected < 1 {
		return nil, status.Errorf(codes.NotFound, "用户不存在 ")
	}

	// 将uint类型的生日转换为time类型的生日
	birthDay := time.Unix(int64(req.BirthDay), 0)
	user.NickName = req.NickName
	user.Birthday = &birthDay
	user.Gender = req.Gender
	result = global.DB.Save(&user)

	if result.RowsAffected < 1 {
		return nil, status.Errorf(codes.Internal, "更新失败")
	}

	return &empty.Empty{}, nil
}

func (s *UserServer) CheckPassword(ctx context.Context, req *proto.PasswordCheckInfo) (*proto.CheckResponse, error) {
	options := &password.Options{16, 100, 32, sha512.New}
	mutiPasswordData := strings.Split(req.EncryptedPassword, "$")
	check := password.Verify(req.Password, mutiPasswordData[2], mutiPasswordData[3], options)

	var checkInfo proto.CheckResponse
	checkInfo.Success = check

	return &checkInfo, nil
}

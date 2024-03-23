package handler

// GetUserList(ctx context.Context, in *PageInfo, opts ...grpc.CallOption) (*UserListResponse, error)
// 	GetUserByMobile(ctx context.Context, in *MobileRequest, opts ...grpc.CallOption) (*UserInfoResponse, error)
// 	GetUserById(ctx context.Context, in *IdRequest, opts ...grpc.CallOption) (*UserInfoResponse, error)
// 	CreateUser(ctx context.Context, in *CreateUserInfo, opts ...grpc.CallOption) (*UserInfoResponse, error)
// 	UpdateUser(ctx context.Context, in *UpdateUserInfo, opts ...grpc.CallOption) (*emptypb.Empty, error)
// 	CheckPassword(ctx context.Context, in *CheckPasswordInfo, opts ...grpc.CallOption) (*CheckResponse, error)

import (
	"context"
	"crypto/sha512"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/anaskhan96/go-password-encoder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"

	"mxshop_srvs/user_srv/global"
	"mxshop_srvs/user_srv/model"
	"mxshop_srvs/user_srv/proto"
)

type UserServer struct{}

func Model2Response(user model.User) proto.UserInfoResponse {
	// 在GRPC中message有默认值，不能随便赋值nil，会报错
	userInfoRsp := proto.UserInfoResponse{
		Id:       user.ID,
		Password: user.Password,
		NickName: user.NickName,
		Gender:   user.Gender,
		Role:     int32(user.Role),
	}
	if user.BirthDay != nil {
		userInfoRsp.Birthday = uint64(user.BirthDay.Unix())
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

// GetUserList 分页获取用户列表
func (s *UserServer) GetUserList(ctx context.Context, in *proto.PageInfo) (*proto.UserListResponse, error) {
	// 查询分页数据
	var users []model.User
	if err := global.DB.Scopes(Paginate(int(in.GetPn()), int(in.GetPSize()))).Find(&users).Error; err != nil {
		return nil, err
	}
	var data []*proto.UserInfoResponse
	for _, user := range users {
		userInfoRsp := Model2Response(user)
		data = append(data, &userInfoRsp)
	}

	// 查询总数量
	var count int64
	if err := global.DB.Model(&model.User{}).Count(&count).Error; err != nil {
		return nil, err
	}

	return &proto.UserListResponse{
		Total: uint64(count),
		Data:  data,
	}, nil
}

// GetUserByMobile 通过手机号查询用户信息
func (s *UserServer) GetUserByMobile(ctx context.Context, in *proto.MobileRequest) (*proto.UserInfoResponse, error) {
	var user model.User
	if err := global.DB.Where(&model.User{
		Mobile: in.Mobile,
	}).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "用户不存在")
		} else {
			return nil, err
		}
	}
	userInfoRsp := Model2Response(user)
	return &userInfoRsp, nil
}

// GetUserById 根据ID查询用户
func (s *UserServer) GetUserById(ctx context.Context, in *proto.IdRequest) (*proto.UserInfoResponse, error) {
	var user model.User
	if err := global.DB.First(&user, in.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "用户不存在")
		} else {
			return nil, err
		}
	}
	userInfoRsp := Model2Response(user)
	return &userInfoRsp, nil
}

// CreateUser 创建用户
func (s *UserServer) CreateUser(ctx context.Context, in *proto.CreateUserInfo) (*proto.UserInfoResponse, error) {
	var user model.User
	result := global.DB.Where(&model.User{Mobile: in.GetMobile()}).First(&user)
	if result.RowsAffected > 0 {
		return nil, status.Errorf(codes.AlreadyExists, "用户已存在")
	}
	user.Mobile = in.Mobile
	user.NickName = in.NickName
	options := &password.Options{SaltLen: 16, Iterations: 100, KeyLen: 32, HashFunction: sha512.New}
	salt, encodedPwd := password.Encode(in.Password, options)
	user.Password = fmt.Sprintf("$pbkdf2-sha512$%s$%s$", salt, encodedPwd)

	if err := global.DB.Create(&user).Error; err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	userInfoRsp := Model2Response(user)
	return &userInfoRsp, nil
}

// UpdateUser 更新用户信息
func (s *UserServer) UpdateUser(ctx context.Context, in *proto.UpdateUserInfo) (*emptypb.Empty, error) {
	var user model.User
	result := global.DB.First(&user, in.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	birthday := time.Unix(int64(in.Birthday), 0)
	user.NickName = in.NickName
	user.BirthDay = &birthday
	user.Gender = in.Gender
	if err := global.DB.Save(&user).Error; err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

// CheckPassword 校验用户密码
func (s *UserServer) CheckPassword(ctx context.Context, in *proto.CheckPasswordInfo) (*proto.CheckResponse, error) {
	options := &password.Options{SaltLen: 16, Iterations: 100, KeyLen: 32, HashFunction: sha512.New}
	passwordInfo := strings.Split(in.EncryptedPassword, "$")
	checkResponse := password.Verify(in.Password, passwordInfo[2], passwordInfo[3], options)
	return &proto.CheckResponse{
		Success: checkResponse,
	}, nil
}
package handler

import (
	"context"
	"encoding/json"
	"mx-shop-srvs/goods_srv/global"
	"mx-shop-srvs/goods_srv/model"
	"mx-shop-srvs/goods_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// 商品分类
func (s *GoodsServer) GetAllCategorysList(ctx context.Context, empty *emptypb.Empty) (*proto.CategoryListResponse, error) {
	var categories []model.Category
	result := global.DB.Where(&model.Category{Level: 1}).Preload("SubCategory.SubCategory").Find(&categories)
	if result.RowsAffected == 0 {
		return nil, result.Error
	}

	b, _ := json.Marshal(categories)

	resp := proto.CategoryListResponse{
		JsonData: string(b),
	}
	return &resp, nil
}

// 获取子分类
func (s *GoodsServer) GetSubCategory(ctx context.Context, req *proto.CategoryListRequest) (*proto.SubCategoryListResponse, error) {
	var category model.Category
	result := global.DB.First(&category, req.Id)

	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "分类不存在")
	}

	var resp proto.SubCategoryListResponse
	resp.Info = &proto.CategoryInfoResponse{
		Id:             category.ID,
		Name:           category.Name,
		ParentCategory: category.ParentCategoryID,
		Level:          category.Level,
		IsTab:          category.IsTab,
	}

	var categories []model.Category
	result = global.DB.Find(&categories)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "无分类数据")
	}

	// 构建了一颗map树，key为分类id，value为分类对象
	nodeMap := make(map[int32]*model.Category)
	for _, cat := range categories {
		cat.SubCategory = nil
		nodeMap[cat.ID] = &cat
	}

	// 遍历所有map节点，构造成树形结构，找到root节点
	var root *model.Category
	for _, node := range nodeMap {
		if node.ParentCategoryID != 0 {
			parentNode := nodeMap[node.ParentCategoryID]
			parentNode.SubCategory = append(parentNode.SubCategory, node)
		}

		if req.Id == node.ID {
			root = node
		}
	}

	var subCategories []*proto.CategoryInfoResponse

	// dfs进行扁平化操作
	var dfs func(*model.Category)
	dfs = func(c *model.Category) {
		if c == nil {
			return
		}

		subCategories = append(subCategories, &proto.CategoryInfoResponse{
			Id:             c.ID,
			Level:          c.Level,
			Name:           c.Name,
			ParentCategory: c.ParentCategoryID,
			IsTab:          c.IsTab,
		})

		for _, v := range c.SubCategory {
			dfs(v)
		}
	}

	for _, v := range root.SubCategory {
		dfs(v)
	}

	resp.SubCategorys = subCategories

	return &resp, nil
}

func (s *GoodsServer) CreateCategory(ctx context.Context, req *proto.CategoryInfoRequest) (*proto.CategoryInfoResponse, error) {
	category := model.Category{}
	// 使用map的原因是当level为1的时候，不给数据库传parent_category_id这个字段，使其在库里为null
	// 倘若使用struct的方式来构建，由于parent_category_id的类型为int，则其缺省值为0，由于其作为了外键，因此数据库会查id为0的分类，找不到就会报错。
	cMap := map[string]interface{}{}
	cMap["name"] = req.Name
	cMap["level"] = req.Level
	cMap["is_tab"] = req.IsTab
	if req.Level != 1 {
		cMap["parent_category_id"] = req.ParentCategory
	}
	result := global.DB.Model(&category).Create(cMap)
	if result.RowsAffected == 0 {
		return nil, result.Error
	}
	return &proto.CategoryInfoResponse{Id: int32(category.ID)}, nil
}

func (s *GoodsServer) DeleteCategory(ctx context.Context, req *proto.DeleteCategoryRequest) (*emptypb.Empty, error) {

	result := global.DB.Where("id = ?", req.Id).Delete(&model.Category{})
	if result.RowsAffected == 0 {
		return &emptypb.Empty{}, result.Error
	}

	return &emptypb.Empty{}, nil
}

func (s *GoodsServer) UpdateCategory(ctx context.Context, req *proto.CategoryInfoRequest) (*emptypb.Empty, error) {
	var category model.Category
	result := global.DB.First(&category, req.Id)
	if result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Errorf(codes.NotFound, "未找到该分类")
	}

	cMap := map[string]interface{}{}
	cMap["name"] = req.Name
	cMap["level"] = req.Level
	cMap["is_tab"] = req.IsTab
	if req.Level != 1 && req.Level != 0 {
		cMap["parent_category_id"] = req.ParentCategory
	}

	result = global.DB.Model(&model.Category{}).Where("id = ?", req.Id).Updates(cMap)
	if result.RowsAffected == 0 {
		return &emptypb.Empty{}, status.Errorf(codes.Internal, "更新失败")
	}

	return &emptypb.Empty{}, nil
}

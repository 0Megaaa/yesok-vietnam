package main

import (
	"encoding/json"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
)

// seedCoreData 注入 Yesok 2.0 的后台账号、服务、流程、配置和演示客户数据。
// 1.意图 -> 保证首次启动即可完成 B/C 双端真实数据闭环验收。
// 2.步骤 -> 幂等写入 admin/123456、五类服务、流程节点、公开配置和演示客户。
// 3.返回 -> 无返回；写入失败时记录日志但不中断服务启动。
func seedCoreData(db *gorm.DB) {
	seedSysUser(db)
	services := seedServices(db)
	seedWorkflowNodes(db, services)
	seedConfigs(db)
	seedAppUser(db)
}

// seedSysUser 注入管家后台超级管理员。
// 1.意图 -> 提供用户要求的 admin / 123456 验收账号。
// 2.步骤 -> 若账号不存在则生成 bcrypt 哈希并创建 sys_users 记录。
// 3.返回 -> 无返回，失败写日志。
func seedSysUser(db *gorm.DB) {
	var count int64
	db.Model(&models.SysUser{}).Where("username = ?", "admin").Count(&count)
	if count > 0 {
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err := db.Create(&models.SysUser{Username: "admin", PasswordHash: string(hash), RealName: "Yesok 总管家", Role: models.RoleAdmin, Status: 1}).Error; err != nil {
		log.Printf("seed sys user failed: %v", err)
	}
}

// seedServices 注入基础服务配置。
// 1.意图 -> 让 C 端首页五宫格和热门卡片完全由 sys_services 驱动。
// 2.步骤 -> 按 service_code 幂等 upsert 五类高频服务。
// 3.返回 -> code 到服务模型的映射，供流程节点绑定 service_id。
func seedServices(db *gorm.DB) map[string]models.SysService {
	serviceSeeds := []models.SysService{
		{ServiceCode: "airport_transfer", ServiceName: "越南机场接送", DisplayName: "豪华接机", Icon: "✈️", CoverImage: "/static/img.png", Description: "双语管家举牌接机，商务车直达酒店。", BasePrice: 65000000, Currency: "VND", Unit: "次", SortOrder: 1, Status: 1, IsHot: true, FormSchema: formSchema("flight_no", "arrival_time", "hotel_address")},
		{ServiceCode: "visa", ServiceName: "越南签证加急", DisplayName: "签证加急", Icon: "🛂", CoverImage: "/static/img.png", Description: "商务、旅游、落地签资料审核与加急通道。", BasePrice: 120000000, Currency: "VND", Unit: "单", SortOrder: 2, Status: 1, IsHot: true, FormSchema: formSchema("passport_name", "passport_no", "entry_date")},
		{ServiceCode: "charter", ServiceName: "商务包车", DisplayName: "商务包车", Icon: "🚘", CoverImage: "/static/img.png", Description: "胡志明、河内、岘港商务包车与行程规划。", BasePrice: 180000000, Currency: "VND", Unit: "天", SortOrder: 3, Status: 1, IsHot: true, FormSchema: formSchema("city", "use_date", "route")},
		{ServiceCode: "translation", ServiceName: "商务翻译", DisplayName: "随行翻译", Icon: "🌐", CoverImage: "/static/img.png", Description: "中越英随行翻译、会议陪同与商务谈判支持。", BasePrice: 150000000, Currency: "VND", Unit: "天", SortOrder: 4, Status: 1, IsHot: false, FormSchema: formSchema("language", "meeting_time", "scene")},
		{ServiceCode: "business", ServiceName: "企业落地咨询", DisplayName: "企业落地", Icon: "🏢", CoverImage: "/static/img.png", Description: "公司注册、选址、财税和本地资源对接。", BasePrice: 350000000, Currency: "VND", Unit: "项", SortOrder: 5, Status: 1, IsHot: false, FormSchema: formSchema("company_name", "industry", "need")},
	}
	result := map[string]models.SysService{}
	for _, seed := range serviceSeeds {
		var item models.SysService
		if err := db.Where("service_code = ?", seed.ServiceCode).First(&item).Error; err != nil {
			db.Create(&seed)
			item = seed
		} else {
			db.Model(&item).Updates(seed)
			db.Where("service_code = ?", seed.ServiceCode).First(&item)
		}
		result[item.ServiceCode] = item
	}
	return result
}

func formSchema(fields ...string) string {
	b, _ := json.Marshal(map[string]interface{}{"fields": fields})
	return string(b)
}

// seedWorkflowNodes 注入流程大脑节点。
// 1.意图 -> 让后台操作按钮由 sys_workflow_nodes 动态渲染。
// 2.步骤 -> 为每个服务写入待处理、报价、收款、履约、完成等状态动作。
// 3.返回 -> 无返回，失败写日志。
func seedWorkflowNodes(db *gorm.DB, services map[string]models.SysService) {
	buttons := []struct {
		current, btn, target string
		pay                  bool
		material             bool
	}{{"pending", "去报价", "quoted", true, false}, {"quoted", "确认收款", "paid", true, false}, {"paid", "开始履约", "in_progress", false, false}, {"in_progress", "完成订单", "completed", false, false}}
	for _, service := range services {
		for idx, btn := range buttons {
			createNode(db, service.ID, btn.current, btn.btn, btn.target, btn.pay, btn.material, idx+1)
		}
		if service.ServiceCode == "visa" {
			createNode(db, service.ID, "pending", "审核资料", "reviewing", false, true, 0)
			createNode(db, service.ID, "reviewing", "资料通过并报价", "quoted", true, false, 1)
		}
	}
}

func createNode(db *gorm.DB, sid uint, current, button, target string, pay, material bool, sort int) {
	var count int64
	db.Model(&models.SysWorkflowNode{}).Where("service_id=? AND current_status=? AND target_status=?", sid, current, target).Count(&count)
	if count == 0 {
		db.Create(&models.SysWorkflowNode{ServiceID: sid, CurrentStatus: current, ButtonName: button, TargetStatus: target, TriggerPayment: pay, RequiredMaterial: material, SortOrder: sort, Remark: "系统种子流程"})
	}
}

// seedConfigs 注入 C 端公开全局配置。
// 1.意图 -> 支撑 /api/v1/configs 动态输出小程序全局配置。
// 2.步骤 -> 幂等写入品牌、Banner、热线、主题色等配置项。
// 3.返回 -> 无返回。
func seedConfigs(db *gorm.DB) {
	seeds := []models.SysConfig{{ConfigKey: "app_name", ConfigValue: "Yesok Vietnam", ValueType: "string", GroupName: "brand", Remark: "应用名称", IsPublic: true}, {ConfigKey: "hero_title", ConfigValue: "越南高端生活服务管家", ValueType: "string", GroupName: "home", Remark: "首页主标题", IsPublic: true}, {ConfigKey: "hero_subtitle", ConfigValue: "接机、签证、包车、翻译、企业落地一站式托管", ValueType: "string", GroupName: "home", Remark: "首页副标题", IsPublic: true}, {ConfigKey: "banner_image", ConfigValue: "/static/img.png", ValueType: "string", GroupName: "home", Remark: "首页 Banner 图", IsPublic: true}, {ConfigKey: "primary_color", ConfigValue: "#0F3D3E", ValueType: "string", GroupName: "theme", Remark: "主色", IsPublic: true}, {ConfigKey: "hotline", ConfigValue: "+84 888 666 168", ValueType: "string", GroupName: "contact", Remark: "管家热线", IsPublic: true}}
	for _, seed := range seeds {
		var item models.SysConfig
		if db.Where("config_key = ?", seed.ConfigKey).First(&item).Error != nil {
			db.Create(&seed)
		} else {
			db.Model(&item).Updates(seed)
		}
	}
}

// seedAppUser 注入演示客户画像。
// 1.意图 -> 让用户矩阵和订单链路具备初始客户数据。
// 2.步骤 -> 按手机号幂等创建一名 VIP 演示客户。
// 3.返回 -> 无返回。
func seedAppUser(db *gorm.DB) {
	var count int64
	db.Model(&models.AppUser{}).Where("phone = ?", "+84901234567").Count(&count)
	if count == 0 {
		db.Create(&models.AppUser{Phone: "+84901234567", Nickname: "陈先生", VipLevel: 2, Balance: 0})
	}
}

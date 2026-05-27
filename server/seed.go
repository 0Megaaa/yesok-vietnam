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
	seedDictTypes(db)
	seedArticles(db)
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
		{ServiceCode: "airport_transfer", ServiceName: "越南机场接送", DisplayName: "豪华接机", Icon: "✈️", CoverImage: "/static/img.png", Description: "双语管家举牌接机，商务车直达酒店。", BasePrice: 65000000, Currency: "VND", Unit: "次", SortOrder: 1, Status: 1, IsHot: true, FormSchema: makeFormSchema([]FieldSchema{
			{Name: "flight_no", Label: "航班号", Type: "text", Required: true, Placeholder: "如 VN123 或 CA901"},
			{Name: "arrival_date", Label: "到达日期", Type: "date", Required: true, Placeholder: "请选择到达日期"},
			{Name: "arrival_time", Label: "到达时间", Type: "text", Required: true, Placeholder: "如 14:30"},
			{Name: "hotel_address", Label: "酒店地址", Type: "textarea", Required: true, Placeholder: "请填写胡志明市区的酒店名称和地址"},
			{Name: "passenger_count", Label: "乘客人数", Type: "select", Required: true, Options: []string{"1人", "2人", "3人", "4人及以上"}},
			{Name: "luggage", Label: "行李件数", Type: "select", Required: false, Options: []string{"无行李", "1件", "2件", "3件及以上"}},
			{Name: "remark", Label: "补充说明", Type: "textarea", Required: false, Placeholder: "其他需求（选填）"},
		})},
		{ServiceCode: "visa", ServiceName: "越南签证加急", DisplayName: "签证加急", Icon: "🛂", CoverImage: "/static/img.png", Description: "商务、旅游、落地签资料审核与加急通道。", BasePrice: 120000000, Currency: "VND", Unit: "单", SortOrder: 2, Status: 1, IsHot: true, FormSchema: makeFormSchema([]FieldSchema{
			{Name: "passport_name", Label: "护照姓名（英文）", Type: "text", Required: true, Placeholder: "与护照一致的英文姓名"},
			{Name: "passport_no", Label: "护照号码", Type: "text", Required: true, Placeholder: "护照上的号码"},
			{Name: "nationality", Label: "国籍", Type: "select", Required: true, Options: []string{"中国", "其他国家"}},
			{Name: "entry_date", Label: "预计入境日期", Type: "date", Required: true, Placeholder: "请选择"},
			{Name: "visa_type", Label: "签证类型", Type: "select", Required: true, Options: []string{"旅游签", "商务签", "落地签", "贴纸签"}},
			{Name: "stay_days", Label: "预计停留天数", Type: "select", Required: true, Options: []string{"15天", "30天", "90天"}},
			{Name: "contact_phone", Label: "联系电话", Type: "phone", Required: true, Placeholder: "国内手机号"},
			{Name: "remark", Label: "备注", Type: "textarea", Required: false, Placeholder: "特殊需求（选填）"},
		})},
		{ServiceCode: "charter", ServiceName: "商务包车", DisplayName: "商务包车", Icon: "🚘", CoverImage: "/static/img.png", Description: "胡志明、河内、岘港商务包车与行程规划。", BasePrice: 180000000, Currency: "VND", Unit: "天", SortOrder: 3, Status: 1, IsHot: true, FormSchema: makeFormSchema([]FieldSchema{
			{Name: "city", Label: "目的地城市", Type: "select", Required: true, Options: []string{"胡志明市", "河内", "岘港", "芽庄", "其他"}},
			{Name: "use_date", Label: "用车日期", Type: "date", Required: true, Placeholder: "请选择日期"},
			{Name: "use_days", Label: "用车天数", Type: "select", Required: true, Options: []string{"1天", "2天", "3天", "4天及以上"}},
			{Name: "route", Label: "行程路线", Type: "textarea", Required: true, Placeholder: "描述大致行程，如：胡志明市→美奈一日游"},
			{Name: "passenger_count", Label: "乘客人数", Type: "select", Required: true, Options: []string{"1-3人", "4-6人", "7-12人", "13人以上"}},
			{Name: "vehicle_type", Label: "车型偏好", Type: "select", Required: false, Options: []string{"轿车", "SUV", "商务车", "小巴", "无要求"}},
			{Name: "remark", Label: "特殊要求", Type: "textarea", Required: false, Placeholder: "儿童座椅/行李规格等（选填）"},
		})},
		{ServiceCode: "translation", ServiceName: "商务翻译", DisplayName: "随行翻译", Icon: "🌐", CoverImage: "/static/img.png", Description: "中越英随行翻译、会议陪同与商务谈判支持。", BasePrice: 150000000, Currency: "VND", Unit: "天", SortOrder: 4, Status: 1, IsHot: false, FormSchema: makeFormSchema([]FieldSchema{
			{Name: "language", Label: "翻译语言", Type: "select", Required: true, Options: []string{"中越", "中英", "中英越"}},
			{Name: "scene", Label: "使用场景", Type: "select", Required: true, Options: []string{"会议陪同", "工厂参观", "商务谈判", "展会翻译", "旅行随行", "其他"}},
			{Name: "meeting_date", Label: "翻译日期", Type: "date", Required: true, Placeholder: "请选择日期"},
			{Name: "meeting_time", Label: "开始时间", Type: "text", Required: true, Placeholder: "如 09:00"},
			{Name: "duration", Label: "预计时长", Type: "select", Required: true, Options: []string{"半天（4小时内）", "全天（8小时内）", "多天"}},
			{Name: "address", Label: "服务地址", Type: "textarea", Required: true, Placeholder: "详细地址或会议地点"},
			{Name: "topic", Label: "会议主题", Type: "textarea", Required: false, Placeholder: "简要描述会议内容（选填）"},
		})},
		{ServiceCode: "business", ServiceName: "企业落地咨询", DisplayName: "企业落地", Icon: "🏢", CoverImage: "/static/img.png", Description: "公司注册、选址、财税和本地资源对接。", BasePrice: 350000000, Currency: "VND", Unit: "项", SortOrder: 5, Status: 1, IsHot: false, FormSchema: makeFormSchema([]FieldSchema{
			{Name: "company_name", Label: "公司名称（拟注册）", Type: "text", Required: true, Placeholder: "中英文名称"},
			{Name: "industry", Label: "所属行业", Type: "select", Required: true, Options: []string{"科技/互联网", "贸易/进出口", "制造/工厂", "餐饮/食品", "教育/培训", "地产/建筑", "其他"}},
			{Name: "registered_capital", Label: "注册资本", Type: "select", Required: true, Options: []string{"10万美元以下", "10-50万美元", "50-100万美元", "100万美元以上", "未定"}},
			{Name: "investor_country", Label: "投资方国籍/地区", Type: "select", Required: true, Options: []string{"中国大陆", "香港", "台湾", "新加坡", "其他"}},
			{Name: "need", Label: "需要哪些服务", Type: "select", Required: true, Options: []string{"注册公司", "工作签/投资签", "厂房/办公室选址", "财税代理", "法律顾问", "全托管"}},
			{Name: "contact_name", Label: "联系人", Type: "text", Required: true, Placeholder: "您的姓名"},
			{Name: "contact_phone", Label: "联系电话", Type: "phone", Required: true, Placeholder: "手机或微信"},
			{Name: "remark", Label: "补充说明", Type: "textarea", Required: false, Placeholder: "预算、特殊要求等（选填）"},
		})},
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

// FieldSchema 描述单个表单字段的元数据。
type FieldSchema struct {
	Name        string   `json:"name"`              // 表单字段名，对应 form_data 中的 key
	Label       string   `json:"label"`             // 前端展示标签
	Type        string   `json:"type"`              // 字段类型：text | select | textarea | date | phone
	Required    bool     `json:"required"`          // 是否必填
	Placeholder string   `json:"placeholder"`       // 输入框占位提示
	Options     []string `json:"options,omitempty"` // select 类型的选项列表
}

// makeFormSchema 构造服务动态表单的 schema JSON。
// 每个服务可自定义字段集合，前端根据 schema 动态渲染表单。
func makeFormSchema(fields []FieldSchema) []byte {
	b, _ := json.Marshal(map[string]interface{}{"fields": fields})
	return b
}

// WorkflowNodeTemplate 描述单个流程节点的静态配置模板。
type WorkflowNodeTemplate struct {
	StageCode       string // 当前节点编码（触发入口）
	StageName       string // 当前节点名称
	MacroStatus     string // 映射主状态
	ActionName      string // B 端操作按钮名称
	NextStageCode   string // 流转到的目标节点编码
	IsManual        bool   // 是否需要人工触发
	RequireMaterial bool   // 流转到下一步是否必传资料
	NotifyType      string // TG 通知类型，空则不发
	SortOrder       int64  // 按钮排序
}

// workflowTemplates 按服务编码聚合流程节点模板。
// 扩展方式：只需在此 map 中新增 key 或追加节点列表，无需修改注入逻辑。
var workflowTemplates = map[string][]WorkflowNodeTemplate{
	"airport_transfer": {
		{StageCode: "start", StageName: "待受理", MacroStatus: "pending", ActionName: "接单", NextStageCode: "quoted", IsManual: true, RequireMaterial: false, SortOrder: 1},
		{StageCode: "quoted", StageName: "已报价", MacroStatus: "quoted", ActionName: "确认收款", NextStageCode: "paid", IsManual: true, RequireMaterial: false, SortOrder: 2},
		{StageCode: "paid", StageName: "已收款", MacroStatus: "paid", ActionName: "开始履约", NextStageCode: "in_progress", IsManual: true, RequireMaterial: true, SortOrder: 3},
		{StageCode: "in_progress", StageName: "履约中", MacroStatus: "in_progress", ActionName: "完成订单", NextStageCode: "completed", IsManual: true, RequireMaterial: false, SortOrder: 4},
	},
	"visa": {
		{StageCode: "start", StageName: "待受理", MacroStatus: "pending", ActionName: "接单", NextStageCode: "reviewing", IsManual: true, RequireMaterial: true, SortOrder: 1},
		{StageCode: "reviewing", StageName: "资料审核中", MacroStatus: "reviewing", ActionName: "审核通过并报价", NextStageCode: "quoted", IsManual: true, RequireMaterial: false, SortOrder: 2},
		{StageCode: "quoted", StageName: "已报价", MacroStatus: "quoted", ActionName: "确认收款", NextStageCode: "paid", IsManual: true, RequireMaterial: false, SortOrder: 3},
		{StageCode: "paid", StageName: "已收款", MacroStatus: "paid", ActionName: "开始履约", NextStageCode: "in_progress", IsManual: true, RequireMaterial: true, SortOrder: 4},
		{StageCode: "in_progress", StageName: "履约中", MacroStatus: "in_progress", ActionName: "完成订单", NextStageCode: "completed", IsManual: true, RequireMaterial: false, SortOrder: 5},
	},
	"charter": {
		{StageCode: "start", StageName: "待受理", MacroStatus: "pending", ActionName: "接单", NextStageCode: "quoted", IsManual: true, RequireMaterial: false, SortOrder: 1},
		{StageCode: "quoted", StageName: "已报价", MacroStatus: "quoted", ActionName: "确认收款", NextStageCode: "paid", IsManual: true, RequireMaterial: false, SortOrder: 2},
		{StageCode: "paid", StageName: "已收款", MacroStatus: "paid", ActionName: "开始履约", NextStageCode: "in_progress", IsManual: true, RequireMaterial: true, SortOrder: 3},
		{StageCode: "in_progress", StageName: "履约中", MacroStatus: "in_progress", ActionName: "完成订单", NextStageCode: "completed", IsManual: true, RequireMaterial: false, SortOrder: 4},
	},
	"translation": {
		{StageCode: "start", StageName: "待受理", MacroStatus: "pending", ActionName: "接单", NextStageCode: "quoted", IsManual: true, RequireMaterial: false, SortOrder: 1},
		{StageCode: "quoted", StageName: "已报价", MacroStatus: "quoted", ActionName: "确认收款", NextStageCode: "paid", IsManual: true, RequireMaterial: false, SortOrder: 2},
		{StageCode: "paid", StageName: "已收款", MacroStatus: "paid", ActionName: "开始履约", NextStageCode: "in_progress", IsManual: true, RequireMaterial: false, SortOrder: 3},
		{StageCode: "in_progress", StageName: "履约中", MacroStatus: "in_progress", ActionName: "完成订单", NextStageCode: "completed", IsManual: true, RequireMaterial: false, SortOrder: 4},
	},
	"business": {
		{StageCode: "start", StageName: "待受理", MacroStatus: "pending", ActionName: "接单", NextStageCode: "quoted", IsManual: true, RequireMaterial: true, SortOrder: 1},
		{StageCode: "quoted", StageName: "已报价", MacroStatus: "quoted", ActionName: "确认收款", NextStageCode: "paid", IsManual: true, RequireMaterial: false, SortOrder: 2},
		{StageCode: "paid", StageName: "已收款", MacroStatus: "paid", ActionName: "开始履约", NextStageCode: "in_progress", IsManual: true, RequireMaterial: true, SortOrder: 3},
		{StageCode: "in_progress", StageName: "履约中", MacroStatus: "in_progress", ActionName: "完成订单", NextStageCode: "completed", IsManual: true, RequireMaterial: false, SortOrder: 4},
	},
}

// seedWorkflowNodes 动态注入流程节点。
// 1.意图 -> 让节点配置与代码解耦，新增服务只需扩展 workflowTemplates。
// 2.步骤 -> 遍历已存在的服务，按 ServiceCode 匹配模板；先清理旧节点再幂等写入。
// 3.返回 -> 无返回，失败写日志。
func seedWorkflowNodes(db *gorm.DB, services map[string]models.SysService) {
	for code, service := range services {
		templates, ok := workflowTemplates[code]
		if !ok || len(templates) == 0 {
			log.Printf("[seed] no workflow template for service_code=%s, skip", code)
			continue
		}

		// 清理该服务已存在的节点，保证幂等
		if err := db.Where("service_id = ?", service.ID).Delete(&models.SysWorkflowNode{}).Error; err != nil {
			log.Printf("[seed] clear workflow nodes for service_id=%d failed: %v", service.ID, err)
			continue
		}

		for _, tpl := range templates {
			node := models.SysWorkflowNode{
				ServiceID:       service.ID,
				StageCode:       tpl.StageCode,
				StageName:       tpl.StageName,
				MacroStatus:     tpl.MacroStatus,
				ActionName:      tpl.ActionName,
				NextStageCode:   tpl.NextStageCode,
				IsManual:        tpl.IsManual,
				RequireMaterial: tpl.RequireMaterial,
				NotifyType:      tpl.NotifyType,
				SortOrder:       tpl.SortOrder,
			}
			if err := db.Create(&node).Error; err != nil {
				log.Printf("[seed] create workflow node for service_id=%d stage=%s failed: %v", service.ID, tpl.StageCode, err)
			}
		}
		log.Printf("[seed] workflow nodes seeded for service_code=%s (%d nodes)", code, len(templates))
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

// seedDictTypes 注入基础字典类型与字典数据。
// 1.意图 -> 让服务分类、资讯分类和订单状态等枚举具备后台可配置基础数据。
// 2.步骤 -> 幂等写入 sys_dict_types 与 sys_dict_data，按 dict_code 和 dict_value 去重。
// 3.返回 -> 无返回。
func seedDictTypes(db *gorm.DB) {
	types := []models.SysDictType{
		{DictName: "服务分类", DictCode: "service_category", Remark: "C 端服务入口分类", Status: 1},
		{DictName: "资讯分类", DictCode: "article_category", Remark: "C 端资讯频道分类", Status: 1},
		{DictName: "订单状态", DictCode: "order_status", Remark: "订单履约流程状态", Status: 1},
	}
	for _, seed := range types {
		var item models.SysDictType
		if db.Where("dict_code = ?", seed.DictCode).First(&item).Error != nil {
			db.Create(&seed)
		} else {
			db.Model(&item).Updates(seed)
		}
	}
	data := []models.SysDictData{
		{DictCode: "service_category", DictLabel: "出行交通", DictValue: "travel", SortOrder: 1, Status: 1, Remark: "接机、包车等移动服务"},
		{DictCode: "service_category", DictLabel: "商务合规", DictValue: "business", SortOrder: 2, Status: 1, Remark: "签证、注册、财税等商务服务"},
		{DictCode: "service_category", DictLabel: "语言协作", DictValue: "language", SortOrder: 3, Status: 1, Remark: "翻译、陪同、会议支持"},
		{DictCode: "article_category", DictLabel: "落地指南", DictValue: "guide", SortOrder: 1, Status: 1, Remark: "越南商务与生活落地知识"},
		{DictCode: "article_category", DictLabel: "城市灵感", DictValue: "city", SortOrder: 2, Status: 1, Remark: "胡志明、河内、岘港等城市内容"},
		{DictCode: "article_category", DictLabel: "服务公告", DictValue: "notice", SortOrder: 3, Status: 1, Remark: "平台服务与活动公告"},
		{DictCode: "order_status", DictLabel: "待受理", DictValue: "pending", SortOrder: 1, Status: 1, Remark: "客户刚提交订单"},
		{DictCode: "order_status", DictLabel: "已报价", DictValue: "quoted", SortOrder: 2, Status: 1, Remark: "管家已完成报价"},
		{DictCode: "order_status", DictLabel: "已收款", DictValue: "paid", SortOrder: 3, Status: 1, Remark: "客户付款完成"},
		{DictCode: "order_status", DictLabel: "履约中", DictValue: "in_progress", SortOrder: 4, Status: 1, Remark: "服务正在履约"},
		{DictCode: "order_status", DictLabel: "已完成", DictValue: "completed", SortOrder: 5, Status: 1, Remark: "订单履约结束"},
	}
	for _, seed := range data {
		var item models.SysDictData
		if db.Where("dict_code = ? AND dict_value = ?", seed.DictCode, seed.DictValue).First(&item).Error != nil {
			db.Create(&seed)
		} else {
			db.Model(&item).Updates(seed)
		}
	}
}

// seedArticles 注入 C 端首页和资讯 Tab 演示内容。
// 1.意图 -> 让资讯模块在首次启动后即可从数据库动态渲染。
// 2.步骤 -> 按 title 幂等写入多条热带奢华风越南商务服务资讯。
// 3.返回 -> 无返回。
func seedArticles(db *gorm.DB) {
	articles := []models.SysArticle{
		{Title: "抵达胡志明后的 6 小时黄金动线", CoverImg: "/static/img.png", Summary: "从机场接送、酒店入住到商务晚宴，Yesok 管家为高净值客户拆解首日抵达节奏。", Content: "抵达越南后的第一天决定了整趟行程的效率。建议提前锁定航班信息、车辆规格、酒店入住窗口与晚宴动线，由双语管家统一协调司机、酒店和餐厅。", Category: "guide", Author: "Yesok Vietnam", Status: 1, SortOrder: 1, ViewCount: 168},
		{Title: "越南商务包车如何选择车型与路线", CoverImg: "/static/img.png", Summary: "商务拜访、工厂考察与城市转场的车型、司机语言和路线规划建议。", Content: "胡志明与周边工业园路况差异明显。商务包车应优先明确乘坐人数、行李件数、工厂地址、等待时长与返程节点，并预留跨区通勤缓冲。", Category: "guide", Author: "Yesok Vietnam", Status: 1, SortOrder: 2, ViewCount: 132},
		{Title: "岘港海岸线上的高端生活灵感", CoverImg: "/static/img.png", Summary: "把商务行程和热带度假融合，让越南目的地服务更从容。", Content: "岘港适合在紧凑商务行程中安排短暂停留。高端客户可组合机场接送、半日包车、会客翻译和海岸线餐厅预订，形成更有记忆点的目的地体验。", Category: "city", Author: "Yesok Vietnam", Status: 1, SortOrder: 3, ViewCount: 98},
		{Title: "签证加急资料准备清单", CoverImg: "/static/img.png", Summary: "护照、入境日期、酒店地址与联系人信息，提前准备可显著缩短办理时间。", Content: "签证加急的关键在于资料准确性。客户应提前确认护照有效期、入境日期、停留天数、越南联系人和酒店地址，管家会在提交前完成二次校验。", Category: "notice", Author: "Yesok Vietnam", Status: 1, SortOrder: 4, ViewCount: 76},
	}
	for _, seed := range articles {
		var item models.SysArticle
		if db.Where("title = ?", seed.Title).First(&item).Error != nil {
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

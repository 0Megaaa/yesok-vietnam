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
		{ServiceCode: "translation", ServiceName: "商务翻译", DisplayName: "随行翻译", Icon: "🌐", CoverImage: "/static/img.png", Description: "中越英随行翻译，会议陪同与商务谈判支持。", BasePrice: 150000000, Currency: "VND", Unit: "天", SortOrder: 4, Status: 1, IsHot: false, FormSchema: makeFormSchema([]FieldSchema{
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
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Placeholder string   `json:"placeholder"`
	Options     []string `json:"options,omitempty"`
}

// makeFormSchema 构造服务动态表单的 schema JSON。
func makeFormSchema(fields []FieldSchema) []byte {
	b, _ := json.Marshal(map[string]interface{}{"fields": fields})
	return b
}

// WorkflowNodeTemplate 描述单个流程节点的静态配置模板。
type WorkflowNodeTemplate struct {
	StageCode    string                // 当前节点编码（触发入口）
	StageName    string                // 当前节点名称
	MacroStatus  string                // 映射主状态
	ActionName   string                // 动作标识（内部使用，唯一键）
	ButtonLabel  string                // 按钮名称（UI 显示）
	ExecutorRole string                // admin/client/both
	ActionType   string                // button_click / form_input / wx_pay
	FormFields   []models.FormFieldDef // form_input 时的字段定义
	NeedAudit    bool                  // 提交后是否需人工审核确认才推进
	TargetStatus string                // 流转目标状态
	NotifyType   string                // 触发通知类型
	SortOrder    int64                 // 按钮排序
}

// workflowTemplates 按服务编码聚合流程节点模板。
var workflowTemplates = map[string][]WorkflowNodeTemplate{
	"airport_transfer": {
		// C 端提交需求 → wait_quote
		{
			StageCode: "start", StageName: "开始", MacroStatus: "pending",
			ActionName: "submit_request", ButtonLabel: "提交需求",
			ExecutorRole: "client", ActionType: "form_input",
			TargetStatus: "wait_quote", NotifyType: "client_to_admin",
			FormFields: []models.FormFieldDef{
				{Key: "contact_name", Label: "联系人", Type: "text", Required: true},
				{Key: "contact_phone", Label: "联系电话", Type: "phone", Required: true},
				{Key: "flight_no", Label: "航班号", Type: "text", Required: true},
				{Key: "arrival_datetime", Label: "到达时间", Type: "datetime", Required: true},
				{Key: "hotel_address", Label: "酒店地址", Type: "textarea", Required: true},
				{Key: "passenger_count", Label: "乘客人数", Type: "number", Required: true},
				{Key: "remark", Label: "补充说明", Type: "textarea", Required: false},
			},
			SortOrder: 1,
		},
		// B 端发送报价 → wait_pay
		{
			StageCode: "wait_quote", StageName: "待报价", MacroStatus: "quoted",
			ActionName: "send_quote", ButtonLabel: "发送报价",
			ExecutorRole: "admin", ActionType: "form_input",
			TargetStatus: "wait_pay", NotifyType: "admin_to_client",
			FormFields: []models.FormFieldDef{
				{Key: "quote_amount", Label: "报价金额", Type: "number", Required: true},
				{Key: "quote_remark", Label: "报价说明", Type: "textarea", Required: false},
			},
			SortOrder: 2,
		},
		// C 端支付 → paid
		{
			StageCode: "wait_pay", StageName: "待支付", MacroStatus: "paid",
			ActionName: "pay_order", ButtonLabel: "立即支付",
			ExecutorRole: "client", ActionType: "wx_pay",
			TargetStatus: "paid", NotifyType: "client_to_admin",
			SortOrder: 3,
		},
		// B 端派车 → dispatching
		{
			StageCode: "paid", StageName: "已支付", MacroStatus: "paid",
			ActionName: "dispatch_driver", ButtonLabel: "安排司机",
			ExecutorRole: "admin", ActionType: "button_click",
			TargetStatus: "dispatching", NotifyType: "admin_to_client",
			SortOrder: 4,
		},
		// B 端开始服务 → in_progress
		{
			StageCode: "dispatching", StageName: "派车中", MacroStatus: "in_progress",
			ActionName: "start_service", ButtonLabel: "开始服务",
			ExecutorRole: "admin", ActionType: "button_click",
			TargetStatus: "in_progress", NotifyType: "admin_to_client",
			SortOrder: 5,
		},
		// B 端完成订单 → completed
		{
			StageCode: "in_progress", StageName: "服务中", MacroStatus: "completed",
			ActionName: "complete_order", ButtonLabel: "完成订单",
			ExecutorRole: "admin", ActionType: "button_click",
			TargetStatus: "completed", NotifyType: "admin_to_client",
			SortOrder: 6,
		},
	},
	"visa": {
		// C 端提交需求 → wait_material
		{
			StageCode: "start", StageName: "开始", MacroStatus: "pending",
			ActionName: "submit_request", ButtonLabel: "提交需求",
			ExecutorRole: "client", ActionType: "form_input",
			TargetStatus: "wait_material", NotifyType: "client_to_admin",
			FormFields: []models.FormFieldDef{
				{Key: "passport_name", Label: "护照姓名", Type: "text", Required: true},
				{Key: "passport_no", Label: "护照号码", Type: "text", Required: true},
				{Key: "entry_date", Label: "预计入境日期", Type: "date", Required: true},
				{Key: "visa_type", Label: "签证类型", Type: "select", Required: true,
					Options: []struct {
						Label string `json:"label"`
						Value string `json:"value"`
					}{
						{Label: "旅游签", Value: "tourist"},
						{Label: "商务签", Value: "business"},
					}},
			},
			SortOrder: 1,
		},
		// C 端上传材料 → reviewing (需审核)
		{
			StageCode: "wait_material", StageName: "待收资料", MacroStatus: "pending",
			ActionName: "upload_material", ButtonLabel: "上传签证材料",
			ExecutorRole: "client", ActionType: "form_input", NeedAudit: true,
			TargetStatus: "reviewing", NotifyType: "client_to_admin",
			FormFields: []models.FormFieldDef{
				{Key: "passport_img", Label: "护照首页", Type: "image", Required: true},
				{Key: "notes", Label: "备注说明", Type: "textarea", Required: false},
			},
			SortOrder: 2,
		},
		// B 端确认收齐 → reviewing
		{
			StageCode: "wait_material", StageName: "待收资料", MacroStatus: "pending",
			ActionName: "material_received", ButtonLabel: "确认收齐",
			ExecutorRole: "admin", ActionType: "button_click",
			TargetStatus: "reviewing", NotifyType: "admin_to_client",
			SortOrder: 3,
		},
		// B 端审核通过并报价 → quoted
		{
			StageCode: "reviewing", StageName: "资料审核中", MacroStatus: "quoted",
			ActionName: "approve", ButtonLabel: "审核通过并报价",
			ExecutorRole: "admin", ActionType: "form_input",
			TargetStatus: "quoted", NotifyType: "admin_to_client",
			FormFields: []models.FormFieldDef{
				{Key: "quote_amount", Label: "报价金额", Type: "number", Required: true},
				{Key: "quote_remark", Label: "报价说明", Type: "textarea", Required: false},
			},
			SortOrder: 4,
		},
		// C 端支付 → paid
		{
			StageCode: "quoted", StageName: "已报价", MacroStatus: "paid",
			ActionName: "pay_order", ButtonLabel: "立即支付",
			ExecutorRole: "client", ActionType: "wx_pay",
			TargetStatus: "paid", NotifyType: "client_to_admin",
			SortOrder: 5,
		},
		// B 端开始履约 → in_progress
		{
			StageCode: "paid", StageName: "已支付", MacroStatus: "in_progress",
			ActionName: "start_service", ButtonLabel: "开始履约",
			ExecutorRole: "admin", ActionType: "button_click",
			TargetStatus: "in_progress", NotifyType: "admin_to_client",
			SortOrder: 6,
		},
		// B 端完成订单 → completed
		{
			StageCode: "in_progress", StageName: "服务中", MacroStatus: "completed",
			ActionName: "complete_order", ButtonLabel: "完成订单",
			ExecutorRole: "admin", ActionType: "button_click",
			TargetStatus: "completed", NotifyType: "admin_to_client",
			SortOrder: 7,
		},
	},
	"charter": {
		// C 端提交需求 → wait_quote
		{
			StageCode: "start", StageName: "开始", MacroStatus: "pending",
			ActionName: "submit_request", ButtonLabel: "提交需求",
			ExecutorRole: "client", ActionType: "form_input",
			TargetStatus: "wait_quote", NotifyType: "client_to_admin",
			FormFields: []models.FormFieldDef{
				{Key: "contact_name", Label: "联系人", Type: "text", Required: true},
				{Key: "contact_phone", Label: "联系电话", Type: "phone", Required: true},
				{Key: "city", Label: "目的地城市", Type: "text", Required: true},
				{Key: "use_date", Label: "用车日期", Type: "date", Required: true},
				{Key: "passenger_count", Label: "乘客人数", Type: "number", Required: true},
				{Key: "remark", Label: "补充说明", Type: "textarea", Required: false},
			},
			SortOrder: 1,
		},
		// B 端发送报价 → wait_pay
		{
			StageCode: "wait_quote", StageName: "待报价", MacroStatus: "quoted",
			ActionName: "send_quote", ButtonLabel: "发送报价",
			ExecutorRole: "admin", ActionType: "form_input",
			TargetStatus: "wait_pay", NotifyType: "admin_to_client",
			FormFields: []models.FormFieldDef{
				{Key: "quote_amount", Label: "报价金额", Type: "number", Required: true},
				{Key: "quote_remark", Label: "报价说明", Type: "textarea", Required: false},
			},
			SortOrder: 2,
		},
		// C 端支付 → paid
		{
			StageCode: "wait_pay", StageName: "待支付", MacroStatus: "paid",
			ActionName: "pay_order", ButtonLabel: "立即支付",
			ExecutorRole: "client", ActionType: "wx_pay",
			TargetStatus: "paid", NotifyType: "client_to_admin",
			SortOrder: 3,
		},
		// B 端开始服务 → in_progress
		{
			StageCode: "paid", StageName: "已支付", MacroStatus: "in_progress",
			ActionName: "start_service", ButtonLabel: "开始服务",
			ExecutorRole: "admin", ActionType: "button_click",
			TargetStatus: "in_progress", NotifyType: "admin_to_client",
			SortOrder: 4,
		},
		// B 端完成订单 → completed
		{
			StageCode: "in_progress", StageName: "服务中", MacroStatus: "completed",
			ActionName: "complete_order", ButtonLabel: "完成订单",
			ExecutorRole: "admin", ActionType: "button_click",
			TargetStatus: "completed", NotifyType: "admin_to_client",
			SortOrder: 5,
		},
	},
	"translation": {
		// C 端提交需求 → wait_quote
		{
			StageCode: "start", StageName: "开始", MacroStatus: "pending",
			ActionName: "submit_request", ButtonLabel: "提交需求",
			ExecutorRole: "client", ActionType: "form_input",
			TargetStatus: "wait_quote", NotifyType: "client_to_admin",
			FormFields: []models.FormFieldDef{
				{Key: "contact_name", Label: "联系人", Type: "text", Required: true},
				{Key: "contact_phone", Label: "联系电话", Type: "phone", Required: true},
				{Key: "scene", Label: "使用场景", Type: "select", Required: true, Options: []struct {
					Label string `json:"label"`
					Value string `json:"value"`
				}{
					{Label: "会议陪同", Value: "meeting"},
					{Label: "商务谈判", Value: "negotiation"},
				}},
				{Key: "service_date", Label: "服务日期", Type: "date", Required: true},
				{Key: "remark", Label: "补充说明", Type: "textarea", Required: false},
			},
			SortOrder: 1,
		},
		// B 端发送报价 → wait_pay
		{
			StageCode: "wait_quote", StageName: "待报价", MacroStatus: "quoted",
			ActionName: "send_quote", ButtonLabel: "发送报价",
			ExecutorRole: "admin", ActionType: "form_input",
			TargetStatus: "wait_pay", NotifyType: "admin_to_client",
			FormFields: []models.FormFieldDef{
				{Key: "quote_amount", Label: "报价金额", Type: "number", Required: true},
				{Key: "quote_remark", Label: "报价说明", Type: "textarea", Required: false},
			},
			SortOrder: 2,
		},
		// C 端支付 → paid
		{
			StageCode: "wait_pay", StageName: "待支付", MacroStatus: "paid",
			ActionName: "pay_order", ButtonLabel: "立即支付",
			ExecutorRole: "client", ActionType: "wx_pay",
			TargetStatus: "paid", NotifyType: "client_to_admin",
			SortOrder: 3,
		},
		// B 端开始服务 → in_progress
		{
			StageCode: "paid", StageName: "已支付", MacroStatus: "in_progress",
			ActionName: "start_service", ButtonLabel: "开始服务",
			ExecutorRole: "admin", ActionType: "button_click",
			TargetStatus: "in_progress", NotifyType: "admin_to_client",
			SortOrder: 4,
		},
		// B 端完成订单 → completed
		{
			StageCode: "in_progress", StageName: "服务中", MacroStatus: "completed",
			ActionName: "complete_order", ButtonLabel: "完成订单",
			ExecutorRole: "admin", ActionType: "button_click",
			TargetStatus: "completed", NotifyType: "admin_to_client",
			SortOrder: 5,
		},
	},
	"business": {
		// C 端提交需求 → wait_quote
		{
			StageCode: "start", StageName: "开始", MacroStatus: "pending",
			ActionName: "submit_request", ButtonLabel: "提交需求",
			ExecutorRole: "client", ActionType: "form_input",
			TargetStatus: "wait_quote", NotifyType: "client_to_admin",
			FormFields: []models.FormFieldDef{
				{Key: "contact_name", Label: "联系人", Type: "text", Required: true},
				{Key: "contact_phone", Label: "联系电话", Type: "phone", Required: true},
				{Key: "company_name", Label: "公司名称", Type: "text", Required: true},
				{Key: "need_service", Label: "需要的服务", Type: "text", Required: true},
				{Key: "remark", Label: "补充说明", Type: "textarea", Required: false},
			},
			SortOrder: 1,
		},
		// B 端发送报价 → wait_pay
		{
			StageCode: "wait_quote", StageName: "待报价", MacroStatus: "quoted",
			ActionName: "send_quote", ButtonLabel: "发送报价",
			ExecutorRole: "admin", ActionType: "form_input",
			TargetStatus: "wait_pay", NotifyType: "admin_to_client",
			FormFields: []models.FormFieldDef{
				{Key: "quote_amount", Label: "报价金额", Type: "number", Required: true},
				{Key: "quote_remark", Label: "报价说明", Type: "textarea", Required: false},
			},
			SortOrder: 2,
		},
		// C 端支付 → paid
		{
			StageCode: "wait_pay", StageName: "待支付", MacroStatus: "paid",
			ActionName: "pay_order", ButtonLabel: "立即支付",
			ExecutorRole: "client", ActionType: "wx_pay",
			TargetStatus: "paid", NotifyType: "client_to_admin",
			SortOrder: 3,
		},
		// B 端开始服务 → in_progress
		{
			StageCode: "paid", StageName: "已支付", MacroStatus: "in_progress",
			ActionName: "start_service", ButtonLabel: "开始服务",
			ExecutorRole: "admin", ActionType: "button_click",
			TargetStatus: "in_progress", NotifyType: "admin_to_client",
			SortOrder: 4,
		},
		// B 端完成订单 → completed
		{
			StageCode: "in_progress", StageName: "服务中", MacroStatus: "completed",
			ActionName: "complete_order", ButtonLabel: "完成订单",
			ExecutorRole: "admin", ActionType: "button_click",
			TargetStatus: "completed", NotifyType: "admin_to_client",
			SortOrder: 5,
		},
	},
}

// seedWorkflowNodes 动态注入流程节点。先清理旧节点再幂等写入。
func seedWorkflowNodes(db *gorm.DB, services map[string]models.SysService) {
	for code, service := range services {
		templates, ok := workflowTemplates[code]
		if !ok || len(templates) == 0 {
			log.Printf("[seed] no workflow template for service_code=%s, skip", code)
			continue
		}

		if err := db.Where("service_id = ?", service.ID).Delete(&models.SysWorkflowNode{}).Error; err != nil {
			log.Printf("[seed] clear workflow nodes for service_id=%d failed: %v", service.ID, err)
			continue
		}

		for _, tpl := range templates {
			node := models.SysWorkflowNode{
				ServiceID:    service.ID,
				StageCode:    tpl.StageCode,
				StageName:    tpl.StageName,
				MacroStatus:  tpl.MacroStatus,
				ActionName:   tpl.ActionName,
				ButtonLabel:  tpl.ButtonLabel,
				ExecutorRole: tpl.ExecutorRole,
				ActionType:   tpl.ActionType,
				FormFields:   tpl.FormFields,
				NeedAudit:    tpl.NeedAudit,
				TargetStatus: tpl.TargetStatus,
				NotifyType:   tpl.NotifyType,
				SortOrder:    tpl.SortOrder,
			}
			if err := db.Create(&node).Error; err != nil {
				log.Printf("[seed] create workflow node for service_id=%d stage=%s failed: %v", service.ID, tpl.StageCode, err)
			}
		}
		log.Printf("[seed] workflow nodes seeded for service_code=%s (%d nodes)", code, len(templates))
	}
}

// seedConfigs 注入 C 端公开全局配置。
func seedConfigs(db *gorm.DB) {
	seeds := []models.SysConfig{
		{ConfigKey: "app_name", ConfigValue: "Yesok Vietnam", ValueType: "string", GroupName: "brand", Remark: "应用名称", IsPublic: true},
		{ConfigKey: "hero_title", ConfigValue: "越南高端生活服务管家", ValueType: "string", GroupName: "home", Remark: "首页主标题", IsPublic: true},
		{ConfigKey: "hero_subtitle", ConfigValue: "接机、签证、包车、翻译、企业落地一站式托管", ValueType: "string", GroupName: "home", Remark: "首页副标题", IsPublic: true},
		{ConfigKey: "banner_image", ConfigValue: "/static/img.png", ValueType: "string", GroupName: "home", Remark: "首页 Banner 图", IsPublic: true},
		{ConfigKey: "primary_color", ConfigValue: "#0F3D3E", ValueType: "string", GroupName: "theme", Remark: "主色", IsPublic: true},
		{ConfigKey: "hotline", ConfigValue: "+84 888 666 168", ValueType: "string", GroupName: "contact", Remark: "管家热线", IsPublic: true},
	}
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
func seedDictTypes(db *gorm.DB) {
	types := []models.SysDictType{
		{DictName: "服务分类", DictCode: "service_category", Remark: "C 端服务入口分类", Status: 1},
		{DictName: "资讯分类", DictCode: "article_category", Remark: "C 端资讯频道分类", Status: 1},
		{DictName: "订单状态", DictCode: "order_status", Remark: "订单履约流程状态", Status: 1},
		{DictName: "主状态", DictCode: "macro_status", Remark: "订单主状态", Status: 1},
		{DictName: "节点阶段", DictCode: "node_stage", Remark: "工作流节点阶段", Status: 1},
		{DictName: "工作流动作", DictCode: "workflow_action", Remark: "工作流动作名称", Status: 1},
		{DictName: "动作类型", DictCode: "action_type", Remark: "工作流动作类型", Status: 1},
		{DictName: "执行角色", DictCode: "executor_role", Remark: "工作流执行角色", Status: 1},
		{DictName: "通知类型", DictCode: "notify_type", Remark: "工作流通知类型", Status: 1},
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
		{DictCode: "service_category", DictLabel: "语言协作", DictValue: "language", SortOrder: 3, Status: 1, Remark: "翻译、陪同，会议支持"},
		{DictCode: "article_category", DictLabel: "落地指南", DictValue: "guide", SortOrder: 1, Status: 1, Remark: "越南商务与生活落地知识"},
		{DictCode: "article_category", DictLabel: "城市灵感", DictValue: "city", SortOrder: 2, Status: 1, Remark: "胡志明、河内、岘港等城市内容"},
		{DictCode: "article_category", DictLabel: "服务公告", DictValue: "notice", SortOrder: 3, Status: 1, Remark: "平台服务与活动公告"},
		{DictCode: "order_status", DictLabel: "待受理", DictValue: "pending", SortOrder: 1, Status: 1, Remark: "客户刚提交订单"},
		{DictCode: "order_status", DictLabel: "已报价", DictValue: "quoted", SortOrder: 2, Status: 1, Remark: "管家已完成报价"},
		{DictCode: "order_status", DictLabel: "已收款", DictValue: "paid", SortOrder: 3, Status: 1, Remark: "客户付款完成"},
		{DictCode: "order_status", DictLabel: "履约中", DictValue: "in_progress", SortOrder: 4, Status: 1, Remark: "服务正在履约"},
		{DictCode: "order_status", DictLabel: "已完成", DictValue: "completed", SortOrder: 5, Status: 1, Remark: "订单履约结束"},
		// macro_status 字典数据
		{DictCode: "macro_status", DictLabel: "待处理", DictValue: "pending", SortOrder: 1, Status: 1, Remark: "待处理"},
		{DictCode: "macro_status", DictLabel: "已报价", DictValue: "quoted", SortOrder: 2, Status: 1, Remark: "已报价"},
		{DictCode: "macro_status", DictLabel: "已支付", DictValue: "paid", SortOrder: 3, Status: 1, Remark: "已支付"},
		{DictCode: "macro_status", DictLabel: "服务中", DictValue: "in_progress", SortOrder: 4, Status: 1, Remark: "服务中"},
		{DictCode: "macro_status", DictLabel: "已完成", DictValue: "completed", SortOrder: 5, Status: 1, Remark: "已完成"},
		{DictCode: "macro_status", DictLabel: "已取消", DictValue: "cancelled", SortOrder: 6, Status: 1, Remark: "已取消"},
		// node_stage 字典数据
		{DictCode: "node_stage", DictLabel: "开始", DictValue: "start", SortOrder: 1, Status: 1, Remark: "开始节点"},
		{DictCode: "node_stage", DictLabel: "待报价", DictValue: "wait_quote", SortOrder: 2, Status: 1, Remark: "待报价"},
		{DictCode: "node_stage", DictLabel: "待支付", DictValue: "wait_pay", SortOrder: 3, Status: 1, Remark: "待支付"},
		{DictCode: "node_stage", DictLabel: "已支付", DictValue: "paid", SortOrder: 4, Status: 1, Remark: "已支付"},
		{DictCode: "node_stage", DictLabel: "派车中", DictValue: "dispatching", SortOrder: 5, Status: 1, Remark: "派车中"},
		{DictCode: "node_stage", DictLabel: "服务中", DictValue: "in_progress", SortOrder: 6, Status: 1, Remark: "服务中"},
		{DictCode: "node_stage", DictLabel: "已完成", DictValue: "completed", SortOrder: 7, Status: 1, Remark: "已完成"},
		// workflow_action 字典数据
		{DictCode: "workflow_action", DictLabel: "提交需求", DictValue: "submit_request", SortOrder: 1, Status: 1, Remark: "提交需求"},
		{DictCode: "workflow_action", DictLabel: "发送报价", DictValue: "send_quote", SortOrder: 2, Status: 1, Remark: "发送报价"},
		{DictCode: "workflow_action", DictLabel: "立即支付", DictValue: "pay_order", SortOrder: 3, Status: 1, Remark: "立即支付"},
		{DictCode: "workflow_action", DictLabel: "安排司机", DictValue: "dispatch_driver", SortOrder: 4, Status: 1, Remark: "安排司机"},
		{DictCode: "workflow_action", DictLabel: "开始服务", DictValue: "start_service", SortOrder: 5, Status: 1, Remark: "开始服务"},
		{DictCode: "workflow_action", DictLabel: "完成订单", DictValue: "complete_order", SortOrder: 6, Status: 1, Remark: "完成订单"},
		{DictCode: "workflow_action", DictLabel: "上传材料", DictValue: "upload_material", SortOrder: 7, Status: 1, Remark: "上传材料"},
		{DictCode: "workflow_action", DictLabel: "确认收齐", DictValue: "material_received", SortOrder: 8, Status: 1, Remark: "确认收齐"},
		// action_type 字典数据
		{DictCode: "action_type", DictLabel: "按钮点击", DictValue: "button_click", SortOrder: 1, Status: 1, Remark: "按钮点击"},
		{DictCode: "action_type", DictLabel: "表单输入", DictValue: "form_input", SortOrder: 2, Status: 1, Remark: "表单输入"},
		{DictCode: "action_type", DictLabel: "微信支付", DictValue: "wx_pay", SortOrder: 3, Status: 1, Remark: "微信支付"},
		// executor_role 字典数据
		{DictCode: "executor_role", DictLabel: "管理员", DictValue: "admin", SortOrder: 1, Status: 1, Remark: "管理员"},
		{DictCode: "executor_role", DictLabel: "客户", DictValue: "client", SortOrder: 2, Status: 1, Remark: "客户"},
		{DictCode: "executor_role", DictLabel: "双方", DictValue: "both", SortOrder: 3, Status: 1, Remark: "双方"},
		// notify_type 字典数据
		{DictCode: "notify_type", DictLabel: "不通知", DictValue: "none", SortOrder: 1, Status: 1, Remark: "不通知"},
		{DictCode: "notify_type", DictLabel: "客户通知后台", DictValue: "client_to_admin", SortOrder: 2, Status: 1, Remark: "客户通知后台"},
		{DictCode: "notify_type", DictLabel: "后台通知客户", DictValue: "admin_to_client", SortOrder: 3, Status: 1, Remark: "后台通知客户"},
		{DictCode: "notify_type", DictLabel: "系统通知", DictValue: "system", SortOrder: 4, Status: 1, Remark: "系统通知"},
	}
	for _, seed := range data {
		var item models.SysDictData
		if db.Where("dict_code = ? AND dict_value = ?", seed.DictCode, seed.DictValue).First(&item).Error != nil {
			db.Create(&seed)
		} else {
			db.Model(&item).Updates(seed)
		}
	}

	// 标准化旧字典，禁用废弃值
	normalizeWorkflowDictData(db)
}

// normalizeWorkflowDictData 清理旧字典数据，只保留标准白名单值。
// 禁用废弃的 dict_code 和不在白名单内的 dict_value。
func normalizeWorkflowDictData(db *gorm.DB) {
	// 1. 禁用旧字典体系
	db.Exec(`UPDATE sys_dict_data SET status = 0 WHERE dict_code IN ('macro_status_list', 'workflow_node_code')`)

	// 2. 定义各字典的白名单
	whitelist := map[string]map[string]bool{
		"service_category": {
			"airport_transfer": true,
			"visa":             true,
			"charter":          true,
			"translation":      true,
			"business":         true,
		},
		"macro_status": {
			"pending":     true,
			"quoted":      true,
			"paid":        true,
			"in_progress": true,
			"completed":   true,
			"cancelled":   true,
		},
		"node_stage": {
			"start":         true,
			"wait_quote":    true,
			"wait_pay":      true,
			"paid":          true,
			"dispatching":   true,
			"in_progress":   true,
			"completed":     true,
			"wait_material": true,
			"reviewing":     true,
		},
		"workflow_action": {
			"submit_request":    true,
			"send_quote":        true,
			"pay_order":         true,
			"dispatch_driver":   true,
			"start_service":     true,
			"complete_order":    true,
			"upload_material":   true,
			"material_received": true,
			"approve":           true,
		},
		"action_type": {
			"button_click": true,
			"form_input":   true,
			"wx_pay":       true,
		},
		"executor_role": {
			"admin":  true,
			"client": true,
			"both":   true,
		},
		"notify_type": {
			"none":            true,
			"client_to_admin": true,
			"admin_to_client": true,
			"system":          true,
		},
	}

	// 3. 遍历白名单，禁用不在白名单内的值
	for dictCode, allowedValues := range whitelist {
		var items []models.SysDictData
		db.Where("dict_code = ? AND status = 1", dictCode).Find(&items)
		for _, item := range items {
			if !allowedValues[item.DictValue] {
				db.Model(&item).Update("status", 0)
				log.Printf("[normalize] disabled dict_data: %s.%s (not in whitelist)", dictCode, item.DictValue)
			}
		}
	}
}

// seedArticles 注入 C 端首页和资讯 Tab 演示内容。
func seedArticles(db *gorm.DB) {
	articles := []models.SysArticle{
		{Title: "抵达胡志明后的 6 小时黄金动线", CoverImg: "/static/img.png", Summary: "从机场接送、酒店入住到商务晚宴，Yesok 管家为高净值客户拆解首日抵达节奏。", Content: "抵达越南后的第一天决定了整趟行程的效率。建议提前锁定航班信息、车辆规格、酒店入住窗口与晚宴动线，由双语管家统一协调司机、酒店和餐厅。", Category: "guide", Author: "Yesok Vietnam", Status: 1, SortOrder: 1, ViewCount: 168},
		{Title: "越南商务包车如何选择车型与路线", CoverImg: "/static/img.png", Summary: "商务拜访、工厂考察与城市转场的车型、司机语言和路线规划建议。", Content: "胡志明与周边工业园路况差异明显。商务包车应优先明确乘坐人数、行李件数、工厂地址、等待时长与返程节点，并预留跨区通勤缓冲。", Category: "guide", Author: "Yesok Vietnam", Status: 1, SortOrder: 2, ViewCount: 132},
		{Title: "岘港海岸线上的高端生活灵感", CoverImg: "/static/img.png", Summary: "把商务行程和热带度假融合，让越南目的地服务更从容。", Content: "岘港适合在紧凑商务行程中安排短暂停留。高端客户可组合机场接送，半日包车、会客翻译和海岸线餐厅预订，形成更有记忆点的目的地体验。", Category: "city", Author: "Yesok Vietnam", Status: 1, SortOrder: 3, ViewCount: 98},
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
func seedAppUser(db *gorm.DB) {
	var count int64
	db.Model(&models.AppUser{}).Where("phone = ?", "+84901234567").Count(&count)
	if count == 0 {
		db.Create(&models.AppUser{Phone: "+84901234567", Nickname: "陈先生", VipLevel: 2, Balance: 0})
	}
}

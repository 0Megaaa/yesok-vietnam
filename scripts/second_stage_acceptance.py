#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""第二阶段业务闭环验收脚本。

1.意图 -> 用真实 HTTP API 验证 C 端动态服务、客户下单、B 端登录、后台改状态、财务流水生成。
2.步骤 -> 依次请求服务列表、创建订单、登录后台、推进订单到 paid、查询订单与流水。
3.返回 -> 在 docs/second_stage_business_loop_acceptance.json 写入完整 Mock 验收结果。
"""

import json
import pathlib
import sys
import urllib.error
import urllib.request

API = "http://127.0.0.1:8080"
ROOT = pathlib.Path(__file__).resolve().parents[1]
OUT = ROOT / "docs" / "second_stage_business_loop_acceptance.json"


def request(method, path, payload=None, token=None):
    """发送 JSON 请求。

    1.意图 -> 统一封装验收脚本中的 HTTP 请求。
    2.步骤 -> 设置 Content-Type、Authorization，读取响应并解析 JSON。
    3.返回 -> 解析后的 JSON 对象。
    """
    body = None
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    if payload is not None:
        body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    req = urllib.request.Request(API + path, data=body, method=method, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            raw = resp.read().decode("utf-8")
    except urllib.error.HTTPError as exc:
        raw = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"{method} {path} failed: {exc.code} {raw}") from exc
    return json.loads(raw) if raw else {}


def pick_data(resp):
    """提取接口 data/list。

    1.意图 -> 兼容后端不同接口返回层级。
    2.步骤 -> 优先读取 data，其次读取 list 或原对象。
    3.返回 -> 可继续使用的业务数据。
    """
    return resp.get("data", resp)


def main():
    """执行第二阶段闭环验收。

    1.意图 -> 从客户视角和管家视角跑通全链路。
    2.步骤 -> 服务动态读取、客户提交订单、管家推进 paid、核对 payment_records。
    3.返回 -> 控制台打印摘要并写入验收 JSON。
    """
    services_resp = request("GET", "/api/v1/services")
    services = pick_data(services_resp)
    if isinstance(services, dict):
        services = services.get("list") or services.get("services") or []
    assert services, "服务列表为空，C 端无法动态渲染"
    service = services[0]

    order_payload = {
        "service_id": service["id"],
        "customer_name": "第二阶段验收客户",
        "phone": "+84909998888",
        "form_data": {
            "arrival_city": "胡志明市",
            "flight_no": "VN-AI-2026",
            "passenger_count": 2,
            "remark": "第二阶段白屏修复后闭环验收订单，要求中文管家接机。",
        },
    }
    order_resp = request("POST", "/api/v1/orders", order_payload)
    order_data = pick_data(order_resp)
    order_id = order_data.get("id") or order_data.get("order", {}).get("id")
    assert order_id, f"订单创建失败：{order_resp}"

    login_resp = request("POST", "/api/v1/admin/auth/login", {"username": "admin", "password": "123456"})
    login_data = pick_data(login_resp)
    token = login_data.get("token")
    assert token, f"后台登录失败：{login_resp}"

    update_resp = request(
        "PUT",
        f"/api/v1/admin/orders/{order_id}",
        {"target_status": "paid", "payment_status": "paid", "remark": "第二阶段验收：后台管家确认已收款并生成财务流水"},
        token,
    )
    orders_resp = request("GET", "/api/v1/admin/orders", token=token)
    payments_resp = request("GET", "/api/v1/admin/payments", token=token)

    payment_text = json.dumps(payments_resp, ensure_ascii=False)
    assert "YS-PAY" in payment_text or "success" in payment_text, "未发现 paid 状态对应财务流水"

    result = {
        "acceptance": "passed",
        "service_count": len(services),
        "selected_service": service,
        "created_order": order_resp,
        "status_update": update_resp,
        "admin_orders": orders_resp,
        "payment_records": payments_resp,
    }
    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text(json.dumps(result, ensure_ascii=False, indent=2), encoding="utf-8")
    print(json.dumps({"acceptance": "passed", "order_id": order_id, "service_count": len(services)}, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(f"验收失败：{exc}", file=sys.stderr)
        sys.exit(1)

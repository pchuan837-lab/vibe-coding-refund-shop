// R4-① 方案 A 活 import：触发 b2bug.init()，把 Bug1/Bug2 标志真实写入全局。
// 【铁律】本文件仅在 b2-bug 分支存在（main 分支无 internal/b2bug 包，不能出现此文件）。
// 移除 //go:build ignore_inject_never，确保 blank import 生效进入测试二进制。
package domain_test

import _ "refund-shop/internal/b2bug"
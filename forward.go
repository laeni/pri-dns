package pridns

import "net"

// ForwardZone 表示子域转发定义，当定义子域名时为树形结构
type ForwardZone struct {
	Dns      []net.IP
	Enabled  bool
	History  []net.IP // 解析历史
	Order    int      // 排序。值越低，优先级越高
	Children map[string]*ForwardZone
}

// ClientForward 表示每个客户端下面定义的转发配置，其中""表示全局配置。
//
// 结构示例：
//
//		 {
//		   "": ...,
//		   "127.0.0.1": {
//	      "example.org": ...,
//	      "example.com": {
//	        ...,
//	        "children": {
//	          "a": ...,
//	          "foo": ...,
//	          "*": ...,
//	        }
//	      },
//		   },
//		 }
type ClientForward map[string]*map[string]*ForwardZone

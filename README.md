# PriDns

[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20Windows%20%7C%20macOS-cc6600.svg)](release)
[![License](https://img.shields.io/badge/license-Apache%202-blue)](LICENSE)

## Name

*PriDns* - 私人NDS，目标是即使多人同时使用同一个DNS服务，每个人也可以单独对其进行配置，而这些配置不影响其他人。它作为[CoreDNS](https://github.com/coredns/coredns)插件提供服务。

*PriDns*的主要功能有两个：

1. 就像我们平时以为的DNS服务器，可以添加自定义域名的解析。
2. 可以指定将某些特定的域名转发给指定的上层DNS服务器进行解析。这一功能主要是解决某些情况下默认DNS服务器被污染的问题（需要将被污染的域名转发给正常DNS服务器解析）以及某些域名需要使用特定地域的DNS服务器解析（比如当`a.`域名流量走`A`地域代理时，有时候可能期望该域名也由`A`地域的DNS服务器来解析，这样通常得到的IP所在位置可能离`A`地域较近，从而获得更快的访问速度，而*CDN*节点域名一般就属于这类情况）。

## Syntax

```Corefile
pridns {
    dataSourceName DATA_SOURCE_NAME
}
```

- `DATA_SOURCE_NAME` 数据库连接串，只支持*MySQL*。由于该插件的设计目的是面向多个普通用户的，这意味着需要存储大量数据，并且这些数据需要很轻松地随时更改，所以用数据库来存储这些数据。

## Metrics

如果启用监控（通过 _prometheus_ 插件），则导出以下指标：

TODO

## Caveats

* 为了产生最大的匹配性能，我们搜索并返回第一个匹配的上游，因此 dnsredir 之间的块顺序很重要。不像 `proxy` 插件，它总是试图找到最长的匹配，即与位置无关的搜索。

## Examples

使用本地MySQL数据库:

```Corefile
pridns {
    dataSourceName root:@tcp(127.0.0.1:3306)/pridns?parseTime=true&loc=Asia%2FShanghai
}
```

## LICENSE

*PriDns*使用与[CoreDNS](https://github.com/coredns/coredns)相同的[LICENSE](LICENSE)。

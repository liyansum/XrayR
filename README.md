## 本分支需要配合修改过的V2board面板进行使用
https://github.com/wyx2685/v2board

# XrayR
[![](https://img.shields.io/badge/TgChat-@UnOfficialV2board讨论-blue.svg)](https://t.me/unofficialV2board)
[![](https://img.shields.io/badge/TgChat-@XrayR讨论-blue.svg)](https://t.me/XrayR_project)
[![](https://img.shields.io/badge/Channel-@XrayR通知-blue.svg)](https://t.me/XrayR_channel)
![](https://img.shields.io/github/stars/wyx2685/XrayR)
![](https://img.shields.io/github/forks/wyx2685/XrayR)
![](https://github.com/wyx2685/XrayR/actions/workflows/release.yml/badge.svg)
![](https://github.com/wyx2685/XrayR/actions/workflows/docker.yml/badge.svg)
[![Github All Releases](https://img.shields.io/github/downloads/wyx2685/XrayR/total.svg)]()


[English](https://github.com/wyx2685/XrayR/blob/master/README-en.md)|[Iranian](https://github.com/wyx2685/XrayR/blob/master/README_Fa.md)|[Vietnamese](https://github.com/wyx2685/XrayR/blob/master/README-vi.md)

A Xray backend framework that can easily support many panels.

一个基于Xray的后端框架，支持V2ay,Trojan,Shadowsocks协议，极易扩展，支持多面板对接。

如果您喜欢本项目，可以右上角点个star+watch，持续关注本项目的进展。

使用教程：[详细使用教程](https://xrayr-project.github.io/XrayR-doc/)


## 免责声明

本项目只是本人个人学习开发并维护，本人不保证任何可用性，也不对使用本软件造成的任何后果负责。

## 特点

* 永久开源且免费。
* 精简构建仅支持 Trojan、Shadowsocks 和 Shadowsocks-Plugin。
* 支持 TCP、WebSocket 与 TLS；不包含 VMess、VLESS、REALITY、gRPC、mKCP/KCP、HTTPUpgrade 和 XHTTP。
* 支持单实例对接多面板、多节点，无需重复启动。
* 支持限制在线IP
* 支持节点端口级别、用户级别限速。
* 配置简单明了。
* 修改配置自动重启实例。
* 方便编译和升级，可以快速更新核心版本， 支持Xray-core新特性。

## 功能介绍

| 功能 | trojan | shadowsocks |
|---|---|---|
| 获取节点与用户信息 | √ | √ |
| 用户流量与服务器信息上报 | √ | √ |
| 在线人数统计与限制 | √ | √ |
| 审计规则、节点与用户限速 | √ | √ |
| 自定义 DNS | √ | √ |

## 支持前端


各面板仅可使用其 Trojan、Shadowsocks 或 Shadowsocks-Plugin 节点类型；已裁剪的协议节点会在启动阶段被明确拒绝。

## 软件安装

### 本地生产构建

```bash
./build.sh
```

默认使用 `CGO_ENABLED=0`、`GOAMD64=v3`、`-trimpath` 和
`-ldflags="-s -w"`。生成的二进制要求 x86-64-v3 兼容 CPU；需要兼容旧 CPU
时可使用 `GOAMD64=v1 ./build.sh`。

### 一键安装

```
wget -N https://raw.githubusercontent.com/wyx2685/XrayR-release/master/install.sh && bash install.sh
```

### 手动安装

[手动安装教程](https://xrayr-project.github.io/XrayR-doc/xrayr-xia-zai-he-an-zhuang/install/manual)

## 配置文件及详细使用教程

[详细使用教程](https://xrayr-project.github.io/XrayR-doc/)

## Thanks

* [Project X](https://github.com/XTLS/)
* [V2Fly](https://github.com/v2fly)
* [VNet-V2ray](https://github.com/ProxyPanel/VNet-V2ray)
* [Air-Universe](https://github.com/crossfw/Air-Universe)

## Licence

[Mozilla Public License Version 2.0](https://github.com/wyx2685/XrayR/blob/master/LICENSE)

## Telgram

[XrayR后端讨论](https://t.me/XrayR_project)

[XrayR通知](https://t.me/XrayR_channel)

## Stargazers over time

[![Stargazers over time](https://starchart.cc/wyx2685/XrayR.svg)](https://starchart.cc/wyx2685/XrayR)

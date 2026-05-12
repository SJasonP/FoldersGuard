# FoldersGuard

[English](README.md) | 简体中文

> **重要声明：** 本项目的全部源代码均由 AI 撰写。本项目不对安全性、密码学正确性、数据可靠性、生产环境适用性或防止数据丢失作出任何保证。不要把它作为保护重要、敏感或不可替代数据的唯一手段。

FoldersGuard 是一个实验性的桌面端和命令行工具，用于保护文件夹，同时让加密后的数据仍然便于手动移动、上传、下载和分享。

## 为什么做 FoldersGuard

FoldersGuard 围绕一个产品想法设计：加密后的数据在需要存储和搬运时，仍然应该像普通文件和文件夹一样好处理。

加密内容会保留为一个可见的文件夹树，只是名称替换为UUID。
你可以把它复制到另一块硬盘、上传到云盘、只下载其中一部分，或者把选中的加密文件和目录发给别人。
恢复真实名称、元数据和密钥所需的信息，则单独存放在 FoldersGuard 的加密数据库中。

这意味着 FoldersGuard 可以在不提前解密的情况下分享加密文件和目录。分享数据库只描述接收方被允许恢复的内容，所以你可以分享一个加密文件、一个加密目录，或一组选中的文件和目录，而不会暴露父目录、同级条目或无关项目数据。

## 功能特色

- 支持手动处理加密数据：加密输出是普通文件夹树，可以用常规工具复制、上传、下载、备份或分享。
- 直接分享加密内容：分享文件或目录前不需要先导出明文。
- 分享范围受限：`.fgs` 分享数据库只包含选中文件和目录所需的元数据和密钥。
- 完整性校验：无需解密即可验证加密内容，提前发现缺失或被篡改的加密对象。
- 隐藏真实名称：加密后的可见文件名和目录名使用 UUID。
- 元数据独立：FoldersGuard 数据和加密内容分离，因此重命名等纯元数据操作不需要加密内容在本地。
- 保留目录层级：加密树保留原始逻辑结构，方便理解和手动管理存储。
- 大文件分片：大文件可以拆成均衡分片，但在 FoldersGuard 内仍然是一个逻辑文件。
- 同时支持桌面端和 CLI：Wails WebUI 面向交互使用，CLI 面向自动化。

## 当前状态

FoldersGuard 仍在开发中。

当前项目包含：

- 用于扫描、规划、加密、恢复、校验、分享和修改受保护文件夹项目的 Go 核心。
- 名为 `foldersguard` 的 CLI，`fg` 作为短别名。
- 基于同一套 Go 核心的 Wails 桌面 WebUI。
- 使用 React 和 Ant Design 构建的前端，支持英语和简体中文。
- 基于 SQLCipher 的 `.fg` 项目数据库和 `.fgs` 分享数据库。
- 大文件分片、目录层级保留、UUID 可见名称、可移植文件系统元数据保留等能力。

在经过独立审查和充分测试前，本项目应被视为实验性软件。

## 安全模型

FoldersGuard v1 使用：

- SQLCipher 保护项目数据库和分享数据库。
- AES-256-GCM 加密文件内容。
- 每个文件和文件夹使用随机内部 key。
- 密码作为面向用户的解锁方式。
- 使用 UUID 作为加密文件和目录的可见名称。

FoldersGuard 目标是保护文件内容、真实名称、目录元数据和内部 key 材料，避免未授权读取。

FoldersGuard 不尝试隐藏：

- 正在使用 FoldersGuard 这一事实。
- 可见的加密目录层级。
- 目录中的加密条目数量。
- 加密文件或分片的大致大小。
- 存储服务提供方可观察到的修改模式。

再次强调，本项目不作任何安全保证。代码由 AI 撰写，并不等同于经过审计的密码学软件。

## 项目模型

普通 FoldersGuard 项目由一个顶层文件夹创建。

加密内容树仍然是普通文件夹树，可以通过常规存储工具移动、上传、下载或分享。真实名称会被 UUID 名称替代。
把 UUID 名称映射回真实名称所需的元数据，则单独存放在 FoldersGuard 数据中。

在 v1 中：

- `.fg` 用于普通项目数据库，且数据库必须只包含一个顶层目录。
- `.fgs` 用于分享数据库和其它分享范围的数据形态。
- 加密内容和 FoldersGuard 元数据彼此分离。
- 活跃项目数据存放在用户的 FoldersGuard 数据目录中。
- 分享数据库可以描述一个文件、一个文件夹，或一组选中的文件和文件夹。

## 使用界面

### 桌面 WebUI

WebUI 是主要的交互界面，基于 Wails、React 和 Ant Design 构建。

它支持项目创建、导入、导出、检查、解密、校验、删除、分享、加载分享、项目浏览和项目修改流程。

### CLI

CLI 用于自动化和可重复执行的工作流。

主可执行文件名为：

```text
foldersguard
```

短别名为：

```text
fg
```

主要命令包括：

```text
fg encrypt
fg decrypt
fg inspect
fg verify
fg export
fg import
fg share
fg rename
fg add
fg move
fg remove
fg plan encrypt
```

CLI 规范见 `docs/cli.md` 和 `docs/cli/` 下的文件。

## 仓库结构

```text
.
├── cmd/foldersguard/        CLI 入口
├── internal/app/            WebUI 使用的应用服务层
├── internal/cli/            Cobra CLI 命令
├── internal/content/        内容加密和恢复逻辑
├── internal/crypto/         项目使用的密码学基础逻辑
├── internal/db/             SQLite 和 SQLCipher 数据库打开逻辑
├── internal/format/         格式常量和扩展名规则
├── internal/fsmeta/         文件系统元数据采集和恢复辅助逻辑
├── internal/fswalk/         文件系统扫描
├── internal/model/          核心数据结构和分片规划
├── internal/project/        项目规划、执行、恢复和校验逻辑
├── internal/storage/        数据库 schema 和元数据持久化
├── frontend/                React WebUI
├── docs/                    产品、架构、CLI、WebUI 和存储格式文档
└── scripts/                 构建辅助脚本
```

## 开发

需要：

- 与 `go.mod` 声明匹配的 Go 版本。
- 用于前端的 Node.js 和 npm。
- 用于桌面构建的 Wails v2。
- 构建 SQLCipher 支持时，需要可用的 C 编译器。

运行 Go 测试：

```text
go test ./...
```

构建前端：

```text
npm --prefix frontend run build
```

构建 CLI：

```text
make build
```

构建 Wails 桌面应用：

```text
wails build
```

## SQLCipher 和 CGO

FoldersGuard 使用 SQLCipher 保护项目数据库和分享数据库。
SQLCipher 是 CGO 依赖，因此真正的发布构建必须启用 CGO，并准备好目标平台可用的 C 编译器。

如果某个构建产出了可执行文件，但没有可用的 SQLCipher 支持，不应把它视为完整或可用的 FoldersGuard 构建。

Windows AMD64 构建说明见：

```text
docs/build.md
scripts/build-windows-amd64.ps1
```

## 文档

主要文档：

- `docs/product-requirements.md`
- `docs/architecture.md`
- `docs/storage-format.md`
- `docs/security-implementation.md`
- `docs/cli.md`
- `docs/webui.md`
- `docs/build.md`

## 许可证

FoldersGuard 自有源代码使用 MIT License。见 `LICENSE`。

第三方组件按其各自许可证授权。发布或再分发 release 产物前，请查看 `THIRD-PARTY-NOTICES.md`。

## 背景

`cmd/cutter` 当前已支持两类能力：

- `gocli cutter -d <destination>`：克隆整个 Go 项目
- `gocli cutter app -s <source> -n <name>`：在同一项目内克隆 app

现状存在两个限制：

- 整项目 `cutter` 只接受项目根目录存在 `go.mod`
- `cutter app` 虽然会向上查找 `go.work`，但如果根目录只有 `go.work`、没有同级 `go.mod`，仍然会直接报错

本次设计目标是让两条路径都支持以下情况：

- 项目根目录存在 `go.mod`
- 项目根目录没有 `go.mod`，但存在 `go.work`
- `cutter app` 场景下，`apps/<source>` 也可能自带 `go.mod`

## 目标

1. 让整项目 `cutter` 支持在 workspace-only 根目录执行
2. 让 `cutter app` 支持根目录为 workspace-only 的项目
3. 让 `cutter app` 在 app 自身有 `go.mod` 时优先使用该模块信息
4. 在无法唯一推断模块的 workspace 场景下返回明确错误，而不是误判为非法 Go 项目
5. 保持现有复制、忽略、格式化和文本替换逻辑尽量不变

## 非目标

1. 不重写整个 workspace 内所有模块的 module path
2. 不解析并批量修改复制结果中的所有 `go.mod`
3. 不自动重写 `go.work`
4. 不引入新的 CLI 参数
5. 不做与当前问题无关的 cutter 重构

## 方案选择

考虑过三种方案：

1. 仅放宽根目录识别，其他逻辑基本不动
2. 统一拆分“根上下文识别”和“模块来源解析”，并为 `cutter` 与 `cutter app` 分别定义解析规则
3. 完整支持 workspace 多模块重写，包括 `go.work` 与全部 `go.mod` 联动修改

最终选择方案 2。

原因：

- 方案 1 无法满足 workspace 自动推断与 app 自带 `go.mod` 优先级要求
- 方案 3 超出当前 cutter 的复杂度边界，风险和测试成本明显过高
- 方案 2 能覆盖当前需求，同时把复杂度集中在“定位模块来源”这一处，便于控制回归风险

## 总体设计

### 设计原则

将“当前目录是否是合法执行上下文”和“本次操作应该使用哪个 module path”明确拆开。

新的逻辑分成两层：

1. 根上下文识别
2. 命令级模块解析

根上下文识别只回答“这里能不能执行 cutter”；模块解析才回答“导入路径替换和 `go.mod` 修改应该基于哪个模块”。

这可以消除当前 `findProjectRoot` 中“发现 `go.work` 就必须同级有 `go.mod`”这类混合判断。

## 命令行为设计

### 整项目 `gocli cutter -d <destination>`

执行前提：

- 当前目录有 `go.mod`，合法
- 当前目录没有 `go.mod`，但有 `go.work`，也合法

执行流程：

1. 校验当前目录是否为合法项目根
2. 复制项目文件，沿用现有忽略规则
3. 复制 `.go` 文件时继续执行 import 替换
4. 复制完成后按上下文决定是否修改目标根 `go.mod`

修改规则：

- 如果源根目录存在 `go.mod`，保持现有行为，修改目标根 `go.mod` 的 module name
- 如果源根目录只有 `go.work`，不报错，也不伪造目标根 `go.mod`

该模式下，workspace-only 项目的目标是“完整复制原有项目结构”，而不是将 workspace 压成单模块项目。

### `gocli cutter app -s <source> -n <name>`

执行前提：

- 向上查找时，首次命中的合法根可以是 `go.mod` 根，也可以是 `go.work` 根

模块解析优先级：

1. `apps/<source>/go.mod`
2. 根目录 `go.mod`
3. `go.work use` 成员推断

推断规则：

1. 如果 `apps/<source>/go.mod` 存在，直接读取其 module path，作为 import 替换依据
2. 如果 app 自身没有 `go.mod`，但根目录存在 `go.mod`，则沿用根模块路径
3. 如果只有 `go.work`，则解析 `use` 成员
4. 如果 workspace 中只有一个可用模块，直接使用它
5. 如果有多个模块，优先选择目录范围包含 `apps/<source>` 的那个模块
6. 如果仍然无法唯一确定，则返回明确错误

复制规则：

- `.go` 文件仍由 AST 改写 import 路径
- `.yaml` / `.yml` 文本替换逻辑保持不变
- 如果 app 自身带 `go.mod`，复制后要同步修改新 app 目录下 `go.mod` 的 module name

## 组件拆分

保持改动收敛在 `cmd/cutter` 内，并尽量复用现有复制逻辑。

### 1. `detectProjectContext`

职责：识别项目根上下文，不做模块推断。

输入：当前目录

输出建议包含：

- `rootDir`
- `rootGoModPath`
- `goWorkPath`
- `workspaceUseDirs`

说明：

- 只回答“是否为合法执行上下文”
- 找到 `go.mod` 或 `go.work` 任一即可认为是合法 Go 根

### 2. `resolveProjectCloneMode`

职责：给整项目 `cutter` 判断复制后是否需要修改目标根 `go.mod`。

返回结果建议区分：

- root module 项目
- workspace-only 项目

### 3. `resolveAppModulePath`

职责：为 `cutter app` 解析实际用于 import 替换的模块。

输入：

- 项目上下文
- `sourceAppName`

输出建议包含：

- `modulePath`
- `moduleRootDir`
- `moduleSource`

其中 `moduleSource` 可取值类似：

- `app-go-mod`
- `root-go-mod`
- `go-work-inferred`

用于日志输出和错误信息。

### 4. `parseGoWorkUseDirs`

职责：解析 `go.work` 的 `use` 条目并转换为绝对目录路径。

用途：

- 为 workspace 自动推断提供候选模块目录
- 供 `resolveAppModulePath` 判断哪个模块覆盖 `apps/<source>`

### 5. `maybeModifyGoMod`

职责：替代当前直接调用的根 `go.mod` 修改逻辑。

规则：

- 目标目录存在 `go.mod` 才修改
- 不存在则直接跳过

这样可以保证 workspace-only 的整项目复制不会因为目标根没有 `go.mod` 而失败。

### 6. `maybeModifyAppGoMod`

职责：如果 app 本身是独立模块，复制后同步改写该 app 的 `go.mod`。

边界：

- 只处理新 app 自己目录下的 `go.mod`
- 不扫描、不重写整个 workspace 内其他模块

## 数据流

### 整项目 `cutter`

1. 从当前目录构建项目上下文
2. 判断当前是 root module 项目还是 workspace-only 项目
3. 执行项目复制
4. 复制后按模式决定是否修改目标根 `go.mod`
5. 删除目标目录中的 `.git`

### `cutter app`

1. 从当前目录定位项目根上下文
2. 校验 `apps` 目录和源 app 目录存在
3. 解析 app 所属 module path
4. 复制 app 目录
5. 在复制过程中替换 import 与配置文本
6. 如果新 app 自身有 `go.mod`，修改其 module name

## 错误处理

错误文案应区分“根不合法”和“模块无法定位”，避免把 workspace-only 项目误报成非 Go 项目。

建议错误分类如下：

### 1. 非法根目录

场景：向上遍历后既找不到 `go.mod`，也找不到 `go.work`

建议文案：

`current directory is not a Go project root: no go.mod or go.work found`

### 2. 找到 workspace，但无法解析 app 所属模块

场景：

- `apps/<source>` 没有自己的 `go.mod`
- 根目录也没有 `go.mod`
- `go.work` 中存在多个候选模块，且无法唯一确定哪个模块覆盖该 app

建议文案：

`go.work found, but cannot resolve module for apps/<source>: multiple workspace modules match or no module path available`

### 3. 源 app 不存在

保持当前风格即可，不需要额外扩展。

### 4. workspace-only 下执行整项目复制

这是合法场景，不应给出警告或错误。

### 5. app 自带 `go.mod` 但 module 声明与目录结构不完全一致

本次不做强校验。

处理策略：

- 只按读取到的 module 声明做定向改写
- 如果能完成替换，就不额外报结构错误

这样可以避免引入过度设计。

## 测试设计

优先覆盖路径解析和关键回归点。

### 单元测试

建议新增 `util_test.go`，覆盖：

1. 根目录仅有 `go.mod`
2. 根目录仅有 `go.work`
3. 根目录同时有 `go.mod` 和 `go.work`
4. `go.work use` 解析为单模块
5. `go.work use` 解析为多模块
6. `resolveAppModulePath` 命中 app 自身 `go.mod`
7. `resolveAppModulePath` 回退到根 `go.mod`
8. `resolveAppModulePath` 通过 `go.work` 唯一推断成功
9. `resolveAppModulePath` 因多候选无法唯一确定而报错

### 集成测试

建议补充 1 到 2 个文件系统级测试，重点验证主链路。

#### 整项目 `cutter`

1. 根目录只有 `go.mod`：复制成功，目标根 `go.mod` 被改写
2. 根目录只有 `go.work`：复制成功，不因缺少目标根 `go.mod` 失败
3. 根目录同时有 `go.mod` 和 `go.work`：保持当前单模块逻辑优先

#### `cutter app`

1. 根目录 `go.mod` 单模块：import 替换成功
2. 根目录 `go.work`，app 自带 `go.mod`：优先使用 app 自身模块
3. 根目录 `go.work`，app 无 `go.mod`，workspace 仅一个模块：自动推断成功
4. 根目录 `go.work`，app 无 `go.mod`，workspace 多模块但仅一个覆盖 `apps/<source>`：成功
5. 根目录 `go.work`，app 无 `go.mod`，workspace 多模块且无法唯一确定：返回明确错误

### 回归验证点

1. 忽略目录与忽略文件规则不退化
2. `.yaml` / `.yml` 的 app 名替换保持有效
3. `.go` 文件格式化保持有效
4. 现有单模块场景行为不变

## 文档更新

README 与 README_cn 中 `cutter` 章节需要同步更新为：

- 整项目 `cutter` 可在根目录有 `go.mod` 或 `go.work` 时执行
- `cutter app` 可在单模块项目或 workspace 项目中执行
- 如果 workspace 中无法唯一定位 app 所属模块，会返回明确错误
- 如果 app 自身带 `go.mod`，将优先使用 app 的模块路径

## 兼容性与风险

### 兼容性

- 对现有根目录 `go.mod` 的单模块项目保持兼容
- 对已有 `cutter app` import 替换逻辑保持兼容

### 风险

1. `go.work` 成员目录解析错误会导致 app 模块推断错误
2. workspace 多模块覆盖关系判断不清晰时，容易出现误替换
3. app 自带 `go.mod` 后，根模块与子模块并存时需要特别注意优先级

### 风险控制

1. 把根识别和模块解析拆开，减少隐式耦合
2. 在无法唯一推断时显式报错，不做猜测性替换
3. 通过单元测试覆盖优先级和边界场景

## 实施建议

按以下顺序实施：

1. 提取项目上下文识别与 `go.work` 解析函数
2. 重构 `findProjectRoot` 相关逻辑，改为上下文识别 + 模块解析
3. 改造整项目 `cutter` 的根校验与 `go.mod` 修改入口
4. 改造 `cutter app` 的模块优先级解析
5. 增加单元测试与集成测试
6. 更新 README 与 README_cn 文档

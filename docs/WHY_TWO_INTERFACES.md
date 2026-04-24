# 为什么需要两套接口？

## 核心问题

同一个 `gateway.Server`（同一个端口 18790）上，为什么要维护 **WebSocket RPC** 和 **HTTP REST API** 两套独立的接口？

**答案：** 因为使用场景不同，不同的客户端有不同的需求。

---

## 🎯 谁在使用这两套接口

### 1. 前端 Web UI（React 应用）

**文件位置：** `ui/web/src/`

**两个客户端同时在用：**

#### WebSocket 客户端 (`ws-client.ts`)
```typescript
// 用于实时通信、流式输出、事件推送
const ws = new WsClient(
  "ws://localhost:18790/ws",
  () => token,
  () => userId
);

// 实时对话（Agent 执行中实时推送结果）
ws.call('chat.send', { sessionKey, message });

// 监听事件（Agent 执行进度、错误等）
ws.on('agent', (event) => {
  console.log(event.phase); // thinking → acting → observing
});
```

#### HTTP 客户端 (`http-client.ts`)
```typescript
// 用于常规查询、配置管理（无需实时推送）
const http = new HttpClient(
  "http://localhost:18790/v1"
);

// 查询系统配置（只读，不需要实时）
const config = await http.get('/v1/system-configs');

// 更新配置（但无需 Agent 执行进度）
const agents = await http.get('/v1/agents');
```

### 2. 桌面 App（Electron / Wails）

**文件位置：** `ui/desktop/frontend/`

**只用 WebSocket：**
```typescript
// 桌面 App 内嵌的 gateway，需要实时更新
ws.call('chat.send', { sessionKey, message });
ws.on('agent', updateProgressBar);
```

### 3. 第三方集成（其他后端服务）

**使用 HTTP REST API：**
```bash
# 来自另一个服务的调用（例如 Python 脚本、Java 后端）
curl -X POST http://localhost:18790/v1/agents/my-agent/summon \
  -H "Authorization: Bearer API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello"}'

# 无需实时推送，只需要最终结果
```

### 4. 脚本/自动化

**使用 HTTP REST API：**
```bash
# Shell 脚本、自动化工具
for agent in $(curl -X GET http://localhost:18790/v1/agents); do
  echo "Agent: $agent"
done
```

---

## 📊 为什么这样设计

### 场景对比表

| 场景 | 需求 | 选择 | 原因 |
|------|------|------|------|
| **Web 聊天** | 实时看到 Agent 执行进度 | WebSocket | 流式输出、事件推送 |
| **获取配置** | 只需要读一次 | HTTP | 无连接开销 |
| **删除会话** | 简单的增删改查 | HTTP | 标准化、易集成 |
| **Webhook** | Agent 执行完毕通知 | HTTP | 第三方服务只能用 HTTP |
| **桌面 App** | 实时 UI 更新 | WebSocket | 嵌入式需要实时 |
| **后端服务** | 与其他系统集成 | HTTP | 标准 REST API |

### 两套系统的优势

```
┌─ WebSocket RPC
│  ✅ 实时推送（Agent 执行中的每个阶段）
│  ✅ 双向通信（Client ↔ Server）
│  ✅ 事件订阅（监听特定 Agent 的输出）
│  ✅ 流式输出（逐字返回结果，不用等待）
│  ❌ 无连接支持（需要长连接）
│  ❌ 不适合脚本化
│
└─ HTTP REST API
   ✅ 标准化（任何语言都支持）
   ✅ 无状态（无需长连接）
   ✅ 脚本友好（curl、Python requests 等）
   ✅ 第三方集成容易
   ✅ CDN 友好（可缓存）
   ❌ 无实时推送（需要轮询）
   ❌ 不支持事件订阅
```

---

## 💻 实际例子

### 例子 1：发送聊天消息

**前端 Web UI（需要实时看到结果）：**
```typescript
// ui/web/src/pages/chat/chat-page.tsx
const { ws } = useWs();

async function sendMessage(sessionKey, message) {
  // WebSocket：实时看到 Agent 思考、执行、观察的过程
  const stream = ws.call<ChatResponse>('chat.send', {
    sessionKey,
    message,
  });

  // 实时更新 UI
  for await (const chunk of stream) {
    updateUI(chunk); // 逐字返回
  }
}
```

**后端服务（只需要最终结果）：**
```bash
# 第三方后端服务的调用
curl -X POST http://localhost:18790/v1/agents/my-agent/summon \
  -H "Authorization: Bearer TOKEN" \
  -d '{"message": "Hello"}'

# 获得最终结果后继续处理
```

### 例子 2：列出配置

**前端 Web UI（可用任意方式，但这里用 HTTP 因为只需查询）：**
```typescript
const { http } = useWs();

// 简单查询，用 HTTP
const configs = await http.get('/v1/system-configs');
```

**后端服务（标准 HTTP 调用）：**
```python
import requests

response = requests.get(
    "http://localhost:18790/v1/system-configs",
    headers={"Authorization": f"Bearer {api_key}"}
)
config = response.json()
```

---

## 🏗️ 代码架构

### WebSocket RPC 处理器

**位置：** `internal/gateway/methods/`

用于需要**实时互动**的场景：
- `chat.send` - 实时对话
- `agent.wait` - 等待 Agent 状态变化
- 事件推送 - 订阅 Agent 执行进度

### HTTP REST API 处理器

**位置：** `internal/http/`

用于**标准化 API** 的场景：
- `/v1/agents` - CRUD
- `/v1/sessions` - CRUD
- `/v1/chat/completions` - OpenAI 兼容接口

### 共享核心逻辑

**位置：** `internal/agent/`, `internal/store/`, `internal/pipeline/`

两套接口**都调用同一个业务逻辑**：
- Agent 执行引擎
- 数据库操作
- 权限检查

---

## 🔄 Web UI 中的实际使用

**文件：** `ui/web/src/components/providers/ws-provider.tsx`

```typescript
// 同时创建两个客户端
const ws = new WsClient(...);        // WebSocket 实时通信
const http = new HttpClient(...);    // HTTP 查询

// 都暴露给 React 组件使用
return (
  <WsContext.Provider value={{ ws, http }}>
    {children}
  </WsContext.Provider>
);
```

**实际组件使用：**

```typescript
// 对话页面（需要实时）
const { ws } = useWs();
await ws.call('chat.send', { ... });

// 配置页面（只需查询）
const { http } = useWs();
const config = await http.get('/v1/system-configs');

// Agent 列表（可用任意一个，这里用 HTTP）
const agents = await http.get('/v1/agents');
```

---

## 📈 总结

### 两套接口的使用者

| 使用者 | 接口 | 原因 |
|--------|------|------|
| **Web UI** | WebSocket + HTTP | 前者实时推送，后者查询 |
| **桌面 App** | WebSocket | 嵌入式需要实时 |
| **Python 脚本** | HTTP | 标准库支持 |
| **Java 后端** | HTTP | REST 友好 |
| **Webhook** | HTTP | 入站 HTTP 事件 |
| **CLI 工具** | HTTP | 易于脚本化 |
| **第三方集成** | HTTP | 标准化接口 |

### 为什么不用一套

- ❌ **纯 WebSocket**：脚本、后端服务难以集成（需要处理连接、心跳、重连）
- ❌ **纯 HTTP**：无法实时推送 Agent 执行进度，用户体验差（需要轮询）

### 为什么要两套

- ✅ **各取所长**：WebSocket 用于实时场景，HTTP 用于集成场景
- ✅ **兼容性**：支持任何客户端（浏览器、移动、后端、脚本）
- ✅ **性能**：WebSocket 长连接高效，HTTP 无连接开销小
- ✅ **标准化**：HTTP 是互联网标准，便于第三方接入

---

**结论：** 两套接口不是冗余，而是针对不同场景的最优设计。
- **实时场景** → WebSocket RPC
- **集成场景** → HTTP REST API
- **两者共存** → 最大化兼容性和可用性

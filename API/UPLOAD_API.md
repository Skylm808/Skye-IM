# 文件上传 API - 前端对接文档

## 🎯 服务信息

- **Base URL**: `http://localhost:10600/api/v1/upload`
- **认证**: 所有接口需要JWT Token
- **Gateway路由**: 需要Gateway配置转发到10600端口

---

## 📡 API接口

### 1. 上传图片

**端点**: `POST /api/v1/upload/image`

**用途**: 聊天发图、朋友圈图片

**限制**:
- 最大10MB
- 格式：jpeg, png, gif, webp

**请求**:
```javascript
const formData = new FormData();
formData.append('file', imageFile);

const response = await fetch('http://localhost:10600/api/v1/upload/image', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`
  },
  body: formData
});

const data = await response.json();
// data = { url: "http://...", thumbnail: "...", width: 0, height: 0, size: 12345 }
```

---

### 2. 上传文件

**端点**: `POST /api/v1/upload/file`

**用途**: 聊天发文件、文档

**限制**:
- 最大100MB
- 格式：pdf, doc, docx, zip等

**请求**:
```javascript
const formData = new FormData();
formData.append('file', file);

const response = await fetch('http://localhost:10600/api/v1/upload/file', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`
  },
  body: formData
});

const data = await response.json();
// data = { url: "http://...", filename: "xxx.pdf", size: 12345, mimeType: "application/pdf" }
```

---

### 3. 上传头像

**端点**: `POST /api/v1/upload/avatar`

**用途**: 用户头像、群头像

**限制**:
- 最大5MB
- 格式：jpeg, png, gif, webp

**请求**:
```javascript
const formData = new FormData();
formData.append('file', avatarFile);

const response = await fetch('http://localhost:10600/api/v1/upload/avatar', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`
  },
  body: formData
});

const data = await response.json();
// data = { url: "http://...", thumbnail: "..." }
```

---

## 💻 前端代码示例

### 场景1：发送图片消息

```javascript
// 1. 用户选择图片
<input type="file" accept="image/*" onChange={handleImageSelect} />

// 2. 上传并发送
async function handleImageSelect(e) {
  const file = e.target.files[0];
  
  // 上传到服务器
  const formData = new FormData();
  formData.append('file', file);
  
  const res = await fetch('http://localhost:10600/api/v1/upload/image', {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${getToken()}` },
    body: formData
  });
  
  const { url } = await res.json();
  
  // 通过WebSocket发送图片消息
  socket.send(JSON.stringify({
    type: 'chat',
    data: {
      msgId: uuid(),
      toUserId: targetUserId,
      content: url,           // 图片URL
      contentType: 2          // 2=图片
    }
  }));
}
```

### 场景2：设置头像

```react
// React组件示例
function AvatarUpload({ onSuccess }) {
  const handleUpload = async (e) => {
    const file = e.target.files[0];
    
    // 1. 上传头像
    const formData = new FormData();
    formData.append('file', file);
    
    const res = await fetch('http://localhost:10600/api/v1/upload/avatar', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` },
      body: formData
    });
    
    const { url } = await res.json();
    
    // 2. 调用User API更新头像
    await fetch('/api/v1/user/profile', {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ avatar: url })
    });
    
    onSuccess(url);
  };
  
  return (
    <div>
      <img src={currentAvatar} alt="avatar" />
      <input type="file" accept="image/*" onChange={handleUpload} />
    </div>
  );
}
```

### 场景3：发送文件消息

```javascript
async function sendFile(file, toUserId) {
  // 1. 上传文件
  const formData = new FormData();
  formData.append('file', file);
  
  const res = await fetch('http://localhost:10600/api/v1/upload/file', {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${getToken()}` },
    body: formData
  });
  
  const result = await res.json();
  
  // 2. 发送文件消息
  socket.send(JSON.stringify({
    type: 'chat',
    data: {
      msgId: uuid(),
      toUserId: toUserId,
      content: JSON.stringify({
        url: result.url,
        filename: result.filename,
        size: result.size
      }),
      contentType: 3  // 3=文件
    }
  }));
}
```

---

## 🎨 UI组件建议

### 图片上传按钮
```jsx
<Button icon={<ImageIcon />} onClick={() => fileInputRef.current.click()}>
  发送图片
</Button>
<input 
  ref={fileInputRef} 
  type="file" 
  accept="image/*" 
  style={{display: 'none'}}
  onChange={handleImageUpload}
/>
```

### 文件上传按钮
```jsx
<Button icon={<FileIcon />} onClick={() => fileInputRef.current.click()}>
  发送文件
</Button>
<input 
  ref={fileInputRef} 
  type="file" 
  style={{display: 'none'}}
  onChange={handleFileUpload}
/>
```

### 头像编辑
```jsx
<Avatar src={avatar} size={80} onClick={() => inputRef.current.click()} />
<input 
  ref={inputRef} 
  type="file" 
  accept="image/*" 
  style={{display: 'none'}}
  onChange={handleAvatarChange}
/>
```

---

## 📊 消息内容类型

聊天消息的`contentType`字段：

| 类型 | contentType | content内容 |
|------|-------------|-------------|
| 文本 | 1 | 文本字符串 |
| 图片 | 2 | 图片URL |
| 文件 | 3 | JSON字符串：`{url, filename, size}` |
| 语音 | 4 | 语音文件URL |

---

## ⚠️ 注意事项

1. **JWT Token**: 所有上传接口都需要Authorization header
2. **字段名固定**: FormData的字段名必须是`file`
3. **CORS**: 如果跨域，需要在Gateway配置
4. **Gateway转发**: 需要Gateway添加upload服务路由（端口10600）

### Gateway配置示例
```go
// gateway.go 添加：
staticServices := map[string]string{
    "upload-api": "127.0.0.1:10600",  // 新增
    // ... 其他服务
}
```

---

## 🔗 MinIO访问

上传后的文件通过MinIO访问：
- MinIO控制台：`http://localhost:9001`
- 用户名：`admin`
- 密码：`630630`

---

## ✅ 快速测试

```bash
# 测试图片上传
curl -X POST http://localhost:10600/api/v1/upload/image \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@test.jpg"

# 测试头像上传
curl -X POST http://localhost:10600/api/v1/upload/avatar \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@avatar.png"
```

---

完成！前端可以直接使用这些接口实现文件上传功能。

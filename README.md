# Groqai2APi

https://groq.com/

# docker 参数

- SERVER_PORT 端口
- SERVER_HOST 域名
- SUPPORT_APIKEY 支持官方的apikey调用,支持官方apikey调用会跳过Authorization验证（不想被滥用，慎重开启）
- API_KEY_PREFIX 官方apikey前缀（默认不用填）
- SESSION_TOKEN_LEN session token 长度（默认不用填）
- CHINA_PROMPT 是否内置中文提示
- Authorization 是否密钥验证（注意：SUPPORT_APIKEY开启后，密钥验证失效）
- OpenAuthSecret 是否启用auth路由密钥访问，关闭后，auth路由不可访问
- AuthSecret auth路由的密钥

# 模型映射

> 新增模型映射，只为解决模型名称不一致问题或只有openai的模型。



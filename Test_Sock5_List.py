import aiohttp
import asyncio
from aiohttp_socks import ProxyConnector, ProxyType

# 并发量控制
CONCURRENT_TASKS = 1000  # 并发任务数量
# 超时时间设置
TIMEOUT_SECONDS = 10  # 超时时间，单位是秒

# 测试单个代理
async def test_proxy(session, proxy, semaphore):
    async with semaphore:  # 使用信号量限制并发量
        try:
            # 解析代理地址和端口
            proxy_host, proxy_port = proxy.split(':')
            # 设置代理
            proxy_url = f"socks5://{proxy_host}:{proxy_port}"
            # 发送测试请求
            async with session.get("https://www.cloudflare.com/cdn-cgi/trace", proxy=proxy_url, timeout=aiohttp.ClientTimeout(total=TIMEOUT_SECONDS)) as response:
                # 判断响应状态码是否在期望列表中
                if response.status in [200, 403, 502, 503, 101, 429, 204, 301, 302, 304, 504, 500]:
                    return proxy  # 返回工作代理的地址
        except Exception as e:
            print(f"代理测试失败 {proxy}: {str(e)}")
        return None

# 主异步函数
async def main():
    # 从文件读取代理列表
    with open("socks5_unique.txt", "r") as f:
        proxies = [line.strip() for line in f.readlines()]

    semaphore = asyncio.Semaphore(CONCURRENT_TASKS)  # 创建一个Semaphore实例来限制并发量

    async with aiohttp.ClientSession() as session:
        tasks = [test_proxy(session, proxy, semaphore) for proxy in proxies]
        working_proxies = await asyncio.gather(*tasks)

    # 过滤None结果，获取有效代理列表
    working_proxies = [proxy for proxy in working_proxies if proxy is not None]

    # 将可用的代理保存到文件
    with open("socks5_working.txt", "w") as f:
        for proxy in working_proxies:
            f.write(f"{proxy}\n")

    print(f"> 已将可用的Socks5代理保存至 socks5_working.txt，总计 {len(working_proxies)} 个可用代理。")

if __name__ == "__main__":
    asyncio.run(main())

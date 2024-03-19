import aiohttp
import asyncio
import aiosocks
from aiosocks.connector import ProxyConnector, ProxyClientRequest
from concurrent.futures import ThreadPoolExecutor

# 测试单个代理
async def test_proxy(session, proxy):
    try:
        # 设置代理连接器
        connector = ProxyConnector()
        headers = {'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36'}
        async with session.get("http://www.google.com", proxy=proxy, connector=connector, request_class=ProxyClientRequest, timeout=5, headers=headers) as response:
            if response.status in [200, 403, 502, 503, 101, 429, 204, 301, 302, 304, 504, 500]:
                return proxy.split('@')[-1]  # 返回工作代理的地址
    except Exception as e:
        print(f"代理测试失败 {proxy}: {str(e)}")
    return None

# 主异步函数
async def main():
    # 从文件读取代理列表
    with open("socks5_unique.txt", "r") as f:
        proxies = [f"socks5://{line.strip()}" for line in f.readlines()]

    # 创建异步HTTP会话
    async with aiohttp.ClientSession() as session:
        tasks = [test_proxy(session, proxy) for proxy in proxies]
        # 使用asyncio.gather实现并发
        working_proxies = await asyncio.gather(*tasks)

    # 过滤None结果，获取有效代理列表
    working_proxies = [proxy for proxy in working_proxies if proxy is not None]

    # 将可用的代理保存到文件
    with open("socks5_working.txt", "w") as f:
        for proxy in working_proxies:
            f.write(f"{proxy}\n")

    print(f"> 已将可用的Socks5代理保存至 socks5_working.txt，总计 {len(working_proxies)} 个可用代理。")

# 设置并发限制
asyncio.get_event_loop().run_until_complete(main())

import requests
import argparse

# 设置命令行参数解析
parser = argparse.ArgumentParser(description='下载并去重Socks5代理列表。')
parser.add_argument('--proxy', help='使用代理进行下载。格式为[http://user:pass@host:port]或[socks5://user:pass@host:port]', type=str, default='')
args = parser.parse_args()

# 解析代理参数
proxy = args.proxy
proxies = {}
if proxy:
    if proxy.startswith('http://') or proxy.startswith('https://'):
        proxies = {'http': proxy, 'https': proxy}
    elif proxy.startswith('socks5://'):
        proxies = {'http': proxy, 'https': proxy}

# 定义要获取代理的URL列表
urls = [
    "https://api.proxyscrape.com/v2/?request=getproxies&protocol=socks5&timeout=10000&country=all&simplified=true",
    "https://www.proxy-list.download/api/v1/get?type=socks5",
    "https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt",
    "https://raw.githubusercontent.com/hookzof/socks5_list/master/proxy.txt"
]

# 使用集合来存储去重后的代理
proxies_set = set()

# 从每个URL下载代理列表并添加到集合中，自动去重
for url in urls:
    try:
        response = requests.get(url, proxies=proxies, timeout=5)
        # 按行分割响应内容，然后添加到集合中
        proxies_set.update(response.text.splitlines())
    except Exception as e:
        print(f"从 {url} 下载时出现错误: {e}")

# 将去重后的代理列表写入文件
with open("socks5_unique.txt", 'w') as f:
    for proxy in proxies_set:
        f.write(f"{proxy}\n")

print("> 已将去重后的socks5代理列表保存为 socks5_unique.txt")

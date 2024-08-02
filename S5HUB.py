import requests
import concurrent.futures
import time
import sys
import socket

# 定义要抓取内容的URL列表
urls = [
    'https://api.proxyscrape.com/v2/?request=getproxies&protocol=socks5&timeout=10000&country=all&simplified=true',
    'https://www.proxy-list.download/api/v1/get?type=socks5',
    'https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt',
    'https://raw.githubusercontent.com/hookzof/socks5_list/master/proxy.txt',
    'https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/socks5.txt',
    'https://raw.githubusercontent.com/ErcinDedeoglu/proxies/main/proxies/socks5.txt',
    'https://yakumo.rei.my.id/SOCKS5',
    'https://raw.githubusercontent.com/themiralay/Proxy-List-World/master/data.txt',
    'https://raw.githubusercontent.com/officialputuid/KangProxy/KangProxy/socks5/socks5.txt',
    'https://raw.githubusercontent.com/ALIILAPRO/Proxy/main/socks5.txt',
    'https://raw.githubusercontent.com/0x1337fy/fresh-proxy-list/archive/storage/classic/socks5.txt'
]

# 从多个URL抓取代理列表并去重，保存到 raw_proxies.txt 文件
def fetch_and_save():
    all_proxies = []
    proxies = set()

    for url in urls:
        try:
            response = requests.get(url)
            response.raise_for_status()
            proxy_list = response.text.splitlines()
            all_proxies.extend(proxy_list)
            proxies.update(proxy_list)
            print(f"从 {url} 获取并添加代理成功")
        except Exception as e:
            print(f"抓取 {url} 时出错: {e}")
    
    with open("raw_proxies.txt", "w") as f:
        f.write("\n".join(proxies))
    
    print("代理列表已保存到 raw_proxies.txt")
    print(f"一共收集了 {len(all_proxies)} 个代理")
    print(f"去重后剩余 {len(proxies)} 个代理")
    return all_proxies, proxies

# 验证代理的可用性
def check_proxy(proxy):
    try:
        socks5_proxy = {
            "http": f"socks5://{proxy}",
            "https": f"socks5://{proxy}"
        }
        response = requests.get("https://dash.cloudflare.com/cdn-cgi/trace", proxies=socks5_proxy, timeout=5)
        if response.status_code == 200:
            return proxy
    except Exception:
        return None

# 使用 20 线程并发验证代理可用性
def validate_proxies():
    with open("raw_proxies.txt", "r") as f:
        proxies = f.read().splitlines()

    valid_proxies = []
    total_proxies = len(proxies)
    
    def progress_bar(current, total, bar_length=40):
        progress = current / total
        block = int(bar_length * progress)
        bar = "#" * block + "-" * (bar_length - block)
        text = f"\r验证进度: [{bar}] {current}/{total} ({progress:.2%})"
        sys.stdout.write(text)
        sys.stdout.flush()

    start_time = time.time()
    with concurrent.futures.ThreadPoolExecutor(max_workers=20) as executor:
        results = executor.map(check_proxy, proxies)
        for i, result in enumerate(results):
            progress_bar(i + 1, total_proxies)
            if result:
                valid_proxies.append(result)
    end_time = time.time()

    with open("validated_proxies.txt", "w") as f:
        f.write("\n".join(valid_proxies))
    
    print("\n可用代理已保存到 validated_proxies.txt")
    print(f"验证过后共有 {len(valid_proxies)} 个可用代理")
    print(f"验证代理花费了 {end_time - start_time:.2f} 秒")

# 主函数
if __name__ == "__main__":
    start_time = time.time()
    print("开始抓取代理并保存到 raw_proxies.txt")
    all_proxies, unique_proxies = fetch_and_save()
    print("开始验证代理可用性")
    validate_proxies()
    end_time = time.time()
    print(f"所有操作完成，共花费了 {end_time - start_time:.2f} 秒")

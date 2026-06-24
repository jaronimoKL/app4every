import urllib.request
import ssl
import re

def test():
    url = "https://video.sibnet.ru/shell.php?videoid=4287802"
    ctx = ssl.create_default_context()
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE
    
    try:
        req = urllib.request.Request(
            url,
            headers={
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
                'Referer': 'https://shikimori.one/'
            }
        )
        with urllib.request.urlopen(req, context=ctx, timeout=10) as r:
            print(f"Status: {r.status}")
            print(f"Headers:")
            for k, v in r.getheaders():
                print(f"  {k}: {v}")
            raw_body = r.read()
            body = raw_body.decode('windows-1251', errors='ignore')
            print(f"\nBody sample (first 1000 chars):")
            print(body[:1000])
            
            # Search for mp4 or video URLs in text
            print("\nSearching for mp4 or video URLs in text:")
            links = re.findall(r'src:\s*["\']([^"\']+\.mp4[^"\']*)["\']', body)
            for link in links:
                print(f"- Found source src: {link}")
                
            # Search for any file: "/upload/..." or similar
            links_all = re.findall(r'file:\s*["\']([^"\']+)["\']', body)
            for link in links_all:
                print(f"- Found file: {link}")
                
            # Let's search for "video" or "mp4" in other patterns
            patterns = re.findall(r'["\'](/upload/share/[^"\']+)["\']', body)
            for p in patterns:
                print(f"- Found upload share path: {p}")
                
            # Search for player setup
            print("\nSearching player configs:")
            jw_setup = re.findall(r'jwplayer\([^\)]*\)\.setup\((.*?)\);', body, re.DOTALL)
            for js_obj in jw_setup:
                print(f"JWPlayer setup: {js_obj.strip()}")
                
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    test()

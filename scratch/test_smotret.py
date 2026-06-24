import urllib.request
import ssl
import re

def test():
    url = "https://smotret-anime.org/translations/embed/845524"
    ctx = ssl.create_default_context()
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE
    
    try:
        req = urllib.request.Request(
            url,
            headers={
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
            }
        )
        with urllib.request.urlopen(req, context=ctx, timeout=10) as r:
            body = r.read().decode('utf-8')
            
            # Print script tags content or look for typical videojs configurations
            print("Finding script tags in embed page...")
            scripts = re.findall(r'<script[^>]*>(.*?)</script>', body, re.DOTALL)
            print(f"Total script tags: {len(scripts)}")
            
            for i, script in enumerate(scripts):
                script_trimmed = script.strip()
                if not script_trimmed:
                    continue
                # Look for video URLs or configs
                if any(x in script_trimmed for x in ["video", "source", "playlist", "mp4", "m3u8"]):
                    print(f"\n--- Script {i+1} containing keywords (len={len(script_trimmed)}) ---")
                    lines = script_trimmed.split("\n")
                    # print first 50 lines
                    print("\n".join(lines[:50]))
                    if len(lines) > 50:
                        print("...")
                        
            # Search for any raw link in the body
            print("\nSearching for any mp4/m3u8 or video URLs in text:")
            links = re.findall(r'https?://[^\s"\']+\.(?:mp4|m3u8|m4v)[^\s"\']*', body)
            for link in links:
                print(f"- {link}")
                
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    test()

import urllib.request
import ssl

def test():
    url = "https://smotret-anime.org/js/app/embed.min.js?1780826954"
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
            
            for keyword in ["requireActivation", "sources", "alternativeSources", "backgroundSources", "activation", "activate"]:
                idx = body.find(keyword)
                if idx != -1:
                    print(f"Keyword '{keyword}' found at index {idx}. Snippet:")
                    start = max(0, idx - 150)
                    end = min(len(body), idx + 200)
                    print(body[start:end])
                    print("-" * 50)
                else:
                    print(f"Keyword '{keyword}' NOT found")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    test()

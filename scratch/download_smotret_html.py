import urllib.request
import ssl

def download():
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
            with open("/Users/alexandr/Projects/app4every/scratch/smotret_embed.html", "w", encoding="utf-8") as f:
                f.write(body)
            print("Successfully saved embed page HTML to scratch/smotret_embed.html")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    download()

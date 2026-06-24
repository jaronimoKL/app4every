import urllib.request
import ssl

def check_link(referer):
    url = "https://video.sibnet.ru/v/91c7649a48cda1815bf942ae780c3d95/4287802.mp4"
    ctx = ssl.create_default_context()
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE
    
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
    }
    if referer:
        headers['Referer'] = referer
        
    try:
        req = urllib.request.Request(url, headers=headers, method='HEAD')
        with urllib.request.urlopen(req, context=ctx, timeout=10) as r:
            print(f"Referer: {referer} -> Status: {r.status}")
            print(f"Content-Length: {r.getheader('Content-Length')}")
            print(f"Content-Type: {r.getheader('Content-Type')}")
    except Exception as e:
        print(f"Referer: {referer} -> Error: {e}")

if __name__ == "__main__":
    print("Testing Sibnet direct MP4 links:")
    check_link("https://video.sibnet.ru/shell.php?videoid=4287802")
    check_link(None)
    check_link("https://jaronimo.work.gd/")

import urllib.request
import json

def test():
    url = "https://smarthard.net/v1/api/shikivideos/21"
    try:
        req = urllib.request.Request(
            url, 
            headers={'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)'}
        )
        with urllib.request.urlopen(req, timeout=10) as r:
            body = r.read().decode('utf-8')
            data = json.loads(body)
            
            sibnet_videos = []
            smotret_videos = []
            
            for v in data:
                video_url = v.get("url", "")
                if "sibnet.ru" in video_url:
                    sibnet_videos.append(v)
                elif "smotret-anime" in video_url:
                    smotret_videos.append(v)
            
            print(f"Total Sibnet videos: {len(sibnet_videos)}")
            for i, v in enumerate(sibnet_videos[:5]):
                print(f"Sibnet {i+1}: Ep {v.get('episode')} - {v.get('author')} - {v.get('url')}")
                
            print(f"\nTotal Smotret-Anime videos: {len(smotret_videos)}")
            for i, v in enumerate(smotret_videos[:5]):
                print(f"Smotret-Anime {i+1}: Ep {v.get('episode')} - {v.get('author')} - {v.get('url')}")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    test()

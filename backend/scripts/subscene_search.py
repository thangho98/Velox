#!/usr/bin/env python3
"""Subscene subtitle scraper for Velox.

Uses DrissionPage to bypass Cloudflare protection on sub-scene.com.
Called as a subprocess from Go backend.

Usage:
    subscene_search.py search --query "Inception" --lang en
    subscene_search.py download --url "https://..." --output /tmp/sub.srt
"""

import argparse
import json
import os
import sys
import time
import zipfile
import io

def _has_display():
    """Check if a display is available (X11/Wayland or macOS)."""
    import platform
    if platform.system() == 'Darwin':
        return True
    return bool(os.environ.get('DISPLAY') or os.environ.get('WAYLAND_DISPLAY'))


def get_page():
    """Create a DrissionPage ChromiumPage with Cloudflare bypass.

    Strategy:
    - macOS / Linux with display: normal Chrome (non-headless)
    - Docker / no display: Xvfb virtual framebuffer (still non-headless to CF)
    - Fallback: headless (may get blocked by CF)
    """
    from DrissionPage import ChromiumPage, ChromiumOptions

    co = ChromiumOptions().auto_port()

    # Use system Chrome/Chromium
    chrome_paths = [
        '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
        '/usr/bin/google-chrome',
        '/usr/bin/google-chrome-stable',
        '/usr/bin/chromium-browser',
        '/usr/bin/chromium',
    ]
    for p in chrome_paths:
        if os.path.exists(p):
            co.set_browser_path(p)
            break

    if not _has_display():
        # No display (Docker/headless server) — try Xvfb first
        xvfb_started = _start_xvfb()
        if not xvfb_started:
            # Xvfb not available — fall back to headless
            co.headless(True)
            co.set_argument('--no-sandbox')
            co.set_argument('--disable-gpu')

    co.set_argument('--no-sandbox')
    co.set_argument('--disable-dev-shm-usage')

    return ChromiumPage(co)


def _start_xvfb():
    """Start Xvfb on :99 if not already running. Returns True if display is set."""
    import subprocess
    import shutil

    if os.environ.get('DISPLAY'):
        return True

    xvfb_path = shutil.which('Xvfb')
    if not xvfb_path:
        return False

    try:
        # Start Xvfb on display :99
        subprocess.Popen(
            ['Xvfb', ':99', '-screen', '0', '1920x1080x24', '-nolisten', 'tcp'],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        os.environ['DISPLAY'] = ':99'
        import time
        time.sleep(1)  # Wait for Xvfb to start
        return True
    except Exception:
        return False


def wait_cloudflare(page, timeout=15):
    """Wait for Cloudflare challenge to resolve."""
    for _ in range(timeout):
        title = page.title.lower() if page.title else ''
        if 'just a moment' not in title and 'cloudflare' not in title and title:
            return True
        time.sleep(1)
    return False


ORDINALS = {
    1: 'First', 2: 'Second', 3: 'Third', 4: 'Fourth', 5: 'Fifth',
    6: 'Sixth', 7: 'Seventh', 8: 'Eighth', 9: 'Ninth', 10: 'Tenth',
    11: 'Eleventh', 12: 'Twelfth', 13: 'Thirteenth', 14: 'Fourteenth',
    15: 'Fifteenth', 16: 'Sixteenth', 17: 'Seventeenth', 18: 'Eighteenth',
    19: 'Nineteenth', 20: 'Twentieth',
}


def build_search_query(title, season=0):
    """Build Subscene search query. For TV shows: '{title} - {ordinal} Season'."""
    if season > 0:
        ordinal = ORDINALS.get(season, f'{season}th')
        return f'{title} - {ordinal} Season'
    return title


def search(query, lang, season=0):
    """Search Subscene for subtitles. Returns JSON array to stdout."""
    import urllib.parse

    page = get_page()
    results = []

    try:
        search_title = build_search_query(query, season)
        encoded = urllib.parse.quote_plus(search_title)
        search_url = f"https://sub-scene.com/search?query={encoded}"

        page.get(search_url)
        if not wait_cloudflare(page):
            print(json.dumps({"error": "Cloudflare challenge failed"}))
            return

        time.sleep(2)

        # Find the best matching title link from search results.
        # Links point to /subscene/{slug} pages. Match by comparing
        # the link text against the query — require the query words
        # to appear in the link text (don't blindly pick the first link).
        from bs4 import BeautifulSoup as BS
        search_soup = BS(page.html, 'html.parser')
        series_url = None
        query_lower = query.lower()
        query_words = set(query_lower.split())
        best_score = 0

        for a_tag in search_soup.find_all('a', href=True):
            href = a_tag['href']
            # Title pages use /subscene/{id} pattern
            if '/subscene/' not in href:
                continue
            text = a_tag.get_text(strip=True).lower()
            if not text:
                continue
            # Score: how many query words appear in the link text
            score = sum(1 for w in query_words if w in text)
            if score > best_score:
                best_score = score
                if not href.startswith('http'):
                    href = f"https://sub-scene.com{href}" if href.startswith('/') else f"https://sub-scene.com/{href}"
                series_url = href

        if not series_url:
            print(json.dumps([]))
            return

        # Navigate to subtitle listing page
        page.get(series_url)
        if not wait_cloudflare(page):
            print(json.dumps({"error": "Cloudflare on title page"}))
            return
        time.sleep(2)

        # Parse subtitle rows using BeautifulSoup for speed
        from bs4 import BeautifulSoup
        soup = BeautifulSoup(page.html, 'html.parser')

        # Language filter
        lang_names = {
            'en': 'english', 'vi': 'vietnamese', 'fr': 'french', 'de': 'german',
            'es': 'spanish', 'pt': 'portuguese', 'it': 'italian', 'nl': 'dutch',
            'pl': 'polish', 'ru': 'russian', 'ja': 'japanese', 'ko': 'korean',
            'zh': 'chinese', 'ar': 'arabic', 'tr': 'turkish', 'sv': 'swedish',
            'th': 'thai', 'id': 'indonesian',
        }
        lang_filter = lang_names.get(lang, '').lower()

        rows = soup.find_all('tr')
        for row in rows:
            text_content = row.get_text(separator=' ', strip=True).lower()

            if lang_filter and lang_filter not in text_content:
                continue

            a_elems = row.find_all('a')
            if not a_elems:
                continue

            target_a = a_elems[0]
            href = target_a.get('href', '')
            if not href:
                continue
            if not href.startswith('http'):
                href = f"https://sub-scene.com{href}" if href.startswith('/') else f"https://sub-scene.com/{href}"

            name_text = target_a.get_text(strip=True) or 'Unknown'

            # Detect language
            detected_lang = lang
            for iso, name in lang_names.items():
                if name in text_content:
                    detected_lang = iso
                    break

            results.append({
                'provider': 'subscene',
                'external_id': href,  # details page URL
                'title': name_text.strip(),
                'language': detected_lang,
                'format': 'srt',
                'downloads': 0,
                'rating': 0,
            })

        # Resolve download links for top results (limit to 10 to avoid slowness)
        resolved = []
        for item in results[:10]:
            details_url = item['external_id']
            try:
                page.get(details_url)
                time.sleep(1)
                dl_elem = page.ele('@href:download', timeout=3)
                if dl_elem and dl_elem.link:
                    item['external_id'] = dl_elem.link
                    resolved.append(item)
            except Exception:
                continue

        print(json.dumps(resolved))

    except Exception as e:
        print(json.dumps({"error": str(e)}))
    finally:
        try:
            page.quit()
        except Exception:
            pass


def download(url, output_path):
    """Download a subtitle file from Subscene. Handles ZIP extraction."""
    import requests

    try:
        resp = requests.get(url, headers={
            'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
            'Referer': 'https://sub-scene.com',
        }, timeout=20)

        if resp.status_code != 200:
            print(json.dumps({"error": f"HTTP {resp.status_code}"}))
            return

        # Try ZIP extraction
        valid_exts = ('.srt', '.vtt', '.ass', '.sub', '.ssa')
        try:
            with zipfile.ZipFile(io.BytesIO(resp.content)) as z:
                for name in z.namelist():
                    if '__MACOSX' in name or name.endswith('/'):
                        continue
                    if name.lower().endswith(valid_exts):
                        data = z.read(name)
                        with open(output_path, 'wb') as f:
                            f.write(data)
                        print(json.dumps({"file": output_path, "original_name": name}))
                        return
                print(json.dumps({"error": "No subtitle file in ZIP"}))
        except zipfile.BadZipFile:
            # Not a ZIP, save raw content
            with open(output_path, 'wb') as f:
                f.write(resp.content)
            print(json.dumps({"file": output_path, "original_name": "subtitle.srt"}))

    except Exception as e:
        print(json.dumps({"error": str(e)}))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Subscene subtitle scraper')
    sub = parser.add_subparsers(dest='command')

    search_p = sub.add_parser('search')
    search_p.add_argument('--query', required=True)
    search_p.add_argument('--lang', default='en')
    search_p.add_argument('--season', type=int, default=0, help='Season number for TV shows')

    dl_p = sub.add_parser('download')
    dl_p.add_argument('--url', required=True)
    dl_p.add_argument('--output', required=True)

    args = parser.parse_args()

    if args.command == 'search':
        search(args.query, args.lang, args.season)
    elif args.command == 'download':
        download(args.url, args.output)
    else:
        parser.print_help()
        sys.exit(1)

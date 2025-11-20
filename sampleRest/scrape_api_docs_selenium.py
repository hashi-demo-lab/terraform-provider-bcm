#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
Advanced scraper for BCM API Documentation using Selenium
Handles JavaScript-rendered content (React apps)
"""

import json
import sys
import time
from datetime import datetime
import urllib3

# Disable SSL warnings
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

try:
    from selenium import webdriver
    from selenium.webdriver.chrome.options import Options
    from selenium.webdriver.chrome.service import Service
    from selenium.webdriver.common.by import By
    from selenium.webdriver.support.ui import WebDriverWait
    from selenium.webdriver.support import expected_conditions as EC
    from selenium.common.exceptions import TimeoutException
    SELENIUM_AVAILABLE = True
except ImportError:
    SELENIUM_AVAILABLE = False
    print("Warning: Selenium not installed. Install with: pip install selenium")


class SeleniumScraper:
    """Scraper using Selenium for JavaScript-rendered pages"""

    def __init__(self, headless=True):
        if not SELENIUM_AVAILABLE:
            raise ImportError("Selenium is required but not installed")

        self.headless = headless
        self.driver = None

    def setup_driver(self):
        """Initialize Chrome/Chromium driver"""
        print("Setting up Selenium WebDriver...")

        options = Options()
        if self.headless:
            options.add_argument('--headless')
        options.add_argument('--no-sandbox')
        options.add_argument('--disable-dev-shm-usage')
        options.add_argument('--ignore-certificate-errors')
        options.add_argument('--allow-insecure-localhost')
        options.add_argument('--disable-gpu')

        try:
            # Try chromium first (common in Linux)
            self.driver = webdriver.Chrome(options=options)
            print("✓ Chrome WebDriver initialized")
            return True
        except Exception as e:
            print(f"✗ Failed to initialize WebDriver: {e}")
            print("\nTroubleshooting:")
            print("1. Install Chrome/Chromium: apt-get install chromium-browser")
            print("2. Install ChromeDriver: apt-get install chromium-chromedriver")
            print("3. Or use pip: pip install webdriver-manager")
            return False

    def fetch_page(self, url, wait_time=10):
        """Fetch page and wait for JavaScript to render"""
        if not self.driver:
            print("Driver not initialized")
            return None

        print(f"\nFetching: {url}")
        print(f"Waiting up to {wait_time}s for page to load...")

        try:
            self.driver.get(url)

            # Wait for the React app to load
            # Look for the root div to have content
            wait = WebDriverWait(self.driver, wait_time)

            # Try multiple wait conditions
            try:
                # Wait for root div to have children
                wait.until(lambda d: d.find_element(By.ID, 'root').get_attribute('innerHTML') != '')
                print("✓ React app loaded")
            except TimeoutException:
                print("Warning: Timeout waiting for content, capturing current state")

            # Additional wait for any async content
            time.sleep(2)

            # Get page source after JavaScript rendering
            html_content = self.driver.page_source

            # Get page title
            title = self.driver.title
            print(f"Page title: {title}")

            # Try to get any network requests or console logs
            try:
                logs = self.driver.get_log('performance')
                print(f"Network logs captured: {len(logs)} entries")
            except:
                logs = []

            return {
                'html': html_content,
                'title': title,
                'url': self.driver.current_url,
                'logs': logs
            }

        except Exception as e:
            print(f"Error fetching page: {e}")
            return None

    def extract_api_data(self):
        """Try to extract API documentation from the rendered page"""
        if not self.driver:
            return None

        data = {
            'text_content': None,
            'tables': [],
            'code_blocks': [],
            'lists': [],
            'headings': []
        }

        try:
            # Get all text
            body = self.driver.find_element(By.TAG_NAME, 'body')
            data['text_content'] = body.text

            # Extract headings
            for level in range(1, 7):
                headings = self.driver.find_elements(By.TAG_NAME, f'h{level}')
                for h in headings:
                    data['headings'].append({
                        'level': level,
                        'text': h.text
                    })

            # Extract tables
            tables = self.driver.find_elements(By.TAG_NAME, 'table')
            for idx, table in enumerate(tables):
                rows = table.find_elements(By.TAG_NAME, 'tr')
                table_data = []
                for row in rows:
                    cells = row.find_elements(By.TAG_NAME, 'td')
                    if not cells:
                        cells = row.find_elements(By.TAG_NAME, 'th')
                    table_data.append([cell.text for cell in cells])
                if table_data:
                    data['tables'].append(table_data)

            # Extract code blocks
            code_blocks = self.driver.find_elements(By.TAG_NAME, 'code')
            for code in code_blocks:
                data['code_blocks'].append(code.text)

            # Extract pre blocks
            pre_blocks = self.driver.find_elements(By.TAG_NAME, 'pre')
            for pre in pre_blocks:
                data['code_blocks'].append(pre.text)

            # Extract lists
            lists = self.driver.find_elements(By.TAG_NAME, 'ul')
            lists.extend(self.driver.find_elements(By.TAG_NAME, 'ol'))
            for lst in lists:
                items = lst.find_elements(By.TAG_NAME, 'li')
                data['lists'].append([item.text for item in items])

        except Exception as e:
            print(f"Error extracting data: {e}")

        return data

    def close(self):
        """Close the browser"""
        if self.driver:
            self.driver.quit()
            print("Browser closed")


def save_data(data, filename):
    """Save data to JSON file"""
    with open(filename, 'w', encoding='utf-8') as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
    print(f"Data saved to: {filename}")


def main():
    """Main execution function"""

    target_url = "https://172.21.15.254:8081/api/search?type=service&name=CMDevice"

    print("="*80)
    print("BCM API Documentation Scraper (Selenium)")
    print("="*80)
    print(f"Target URL: {target_url}\n")

    if not SELENIUM_AVAILABLE:
        print("✗ Selenium is not installed")
        print("\nInstall with:")
        print("  pip install selenium")
        print("  apt-get install chromium-browser chromium-chromedriver")
        sys.exit(1)

    # Initialize scraper
    scraper = SeleniumScraper(headless=True)

    if not scraper.setup_driver():
        print("\n✗ Failed to setup WebDriver")
        sys.exit(1)

    try:
        # Fetch the page
        result = scraper.fetch_page(target_url, wait_time=15)

        if not result:
            print("Failed to fetch page")
            sys.exit(1)

        # Generate timestamp
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

        # Save raw HTML (after JavaScript rendering)
        html_filename = f"cmdevice_selenium_html_{timestamp}.html"
        with open(html_filename, 'w', encoding='utf-8') as f:
            f.write(result['html'])
        print(f"\nRendered HTML saved to: {html_filename}")

        # Extract structured data
        print("\nExtracting structured data from rendered page...")
        api_data = scraper.extract_api_data()

        if api_data:
            data_filename = f"cmdevice_selenium_data_{timestamp}.json"
            save_data(api_data, data_filename)

            # Print summary
            print("\n" + "="*80)
            print("Extraction Summary")
            print("="*80)
            print(f"Headings found: {len(api_data['headings'])}")
            print(f"Tables found: {len(api_data['tables'])}")
            print(f"Code blocks found: {len(api_data['code_blocks'])}")
            print(f"Lists found: {len(api_data['lists'])}")

            if api_data['headings']:
                print("\nHeadings:")
                for h in api_data['headings'][:10]:
                    print(f"  H{h['level']}: {h['text']}")

            if api_data['text_content']:
                print(f"\nText content preview (first 500 chars):")
                print("-" * 80)
                print(api_data['text_content'][:500])
                print("-" * 80)

        print("\n✓ Scraping completed successfully")

    except Exception as e:
        print(f"\n✗ Error during scraping: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

    finally:
        scraper.close()


if __name__ == "__main__":
    main()

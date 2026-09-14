import os
from pathlib import Path

from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.remote.file_detector import LocalFileDetector
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait

ROOT = Path(__file__).resolve().parents[1]
XLSX = ROOT / "test.xlsx"
DOCX = ROOT / "test.docx"
BASE_URL = os.environ.get("BASE_URL", "https://templater.local:8443")
SELENIUM_URL = os.environ.get("SELENIUM_URL", "http://127.0.0.1:4444")


def test_open_template_in_onlyoffice():
    assert XLSX.is_file(), "test.xlsx is missing"
    assert DOCX.is_file(), "test.docx is missing"

    options = webdriver.ChromeOptions()
    options.add_argument("--disable-dev-shm-usage")
    driver = webdriver.Remote(command_executor=SELENIUM_URL, options=options)
    driver.file_detector = LocalFileDetector()
    try:
        driver.get(BASE_URL)
        wait = WebDriverWait(driver, 30)
        wait.until(EC.presence_of_element_located((By.CSS_SELECTOR, '[data-testid="session-label"]')))
        wait.until(
            lambda d: "нет сессии" not in d.find_element(By.CSS_SELECTOR, '[data-testid="session-label"]').text
        )
        driver.find_element(By.CSS_SELECTOR, 'input[accept=".xlsx"]').send_keys(str(XLSX))
        wait.until(EC.presence_of_element_located((By.CSS_SELECTOR, 'input[accept=".docx"]')))
        driver.find_element(By.CSS_SELECTOR, 'input[accept=".docx"]').send_keys(str(DOCX))
        open_btn = wait.until(EC.element_to_be_clickable((By.CSS_SELECTOR, '[data-testid="open-task"]')))
        open_btn.click()
        WebDriverWait(driver, 90).until(EC.presence_of_element_located((By.CSS_SELECTOR, ".v-chip")))
        WebDriverWait(driver, 90).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, '[data-testid="doc-editor"] iframe'))
        )
    finally:
        driver.quit()

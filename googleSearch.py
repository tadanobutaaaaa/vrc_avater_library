import requests
import time
import os
from searchIcons import downloadImages
from dotenv import load_dotenv

load_dotenv()

API_KEY = os.getenv('GOOGLE_API_KEY')
CSE_ID = '100d459ed822542ed'
url = 'https://www.googleapis.com/customsearch/v1'

def google_search(pathNameList):
    for name in pathNameList:
        try:
            paramas = {
                'q': name['path'],
                'key': API_KEY,
                'cx': CSE_ID
            }
            response = requests.get(url, params=paramas)
            results = response.json()
            name['url'] = results['items'][0]['formattedUrl']
        except KeyError:
            continue
        
        downloadImages(results['items'][0]['pagemap']['metatags'][0]['og:image'], name['path'])
        time.sleep(2)
    print(pathNameList)
    return pathNameList
